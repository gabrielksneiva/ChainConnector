package sqsqueue

import (
	appconfig "ChainConnector/internal/config"
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

type BlockQueue struct {
	client            *sqs.Client
	cfg               *appconfig.Config
	logger            *zap.Logger
	queueMu           sync.Mutex
	queueURL          string
	producerEnabled   bool
	consumerEnabled   bool
	waitTimeSeconds   int32
	visibilityTimeout int32
}

var _ ports.BlockProducerPort = (*BlockQueue)(nil)

func NewBlockQueue(cfg *appconfig.Config, logger *zap.Logger) (*BlockQueue, error) {
	if cfg == nil {
		cfg = &appconfig.Config{}
	}
	queue := &BlockQueue{
		cfg:               cfg,
		logger:            logger,
		queueURL:          cfg.BlockQueueURL,
		producerEnabled:   cfg.BlockProducerEnabled,
		consumerEnabled:   cfg.BlockConsumerEnabled,
		waitTimeSeconds:   cfg.SQSWaitTimeSeconds,
		visibilityTimeout: cfg.SQSVisibilityTimeout,
	}

	if !cfg.BlockProducerEnabled && !cfg.BlockConsumerEnabled {
		logger.Info("block sqs queue disabled")
		return queue, nil
	}
	if cfg.BlockQueueName == "" && cfg.BlockQueueURL == "" {
		return nil, errors.New("BLOCK_QUEUE_NAME or BLOCK_QUEUE_URL is required when block monitoring is enabled")
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.AWSRegion),
	}
	if cfg.AWSAccessKeyID != "" || cfg.AWSSecretAccessKey != "" {
		if cfg.AWSAccessKeyID == "" || cfg.AWSSecretAccessKey == "" {
			return nil, errors.New("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set together")
		}
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), loadOptions...)
	if err != nil {
		return nil, err
	}
	queue.client = sqs.NewFromConfig(awsCfg, func(options *sqs.Options) {
		if cfg.AWSEndpointURL != "" {
			options.BaseEndpoint = aws.String(cfg.AWSEndpointURL)
		}
	})
	return queue, nil
}

func (q *BlockQueue) Enabled() bool {
	return q != nil && q.producerEnabled && q.client != nil
}

func (q *BlockQueue) EnqueueBlockEvent(ctx context.Context, event *entity.BlockEvent) error {
	if !q.Enabled() {
		return errors.New("block queue producer is disabled")
	}
	if event == nil {
		return errors.New("block event is required")
	}

	queueURL, err := q.ensureQueue(ctx)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = q.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return err
	}
	q.logger.Info("block event sent to sqs", zap.String("chain", event.Chain), zap.Uint64("block_number", event.BlockNumber), zap.String("block_hash", event.BlockHash))
	return nil
}

type BlockQueueConsumer struct {
	queue     *BlockQueue
	processor ports.BlockEventProcessorPort
	logger    *zap.Logger
}

func NewBlockQueueConsumer(queue *BlockQueue, processor ports.BlockEventProcessorPort, logger *zap.Logger) *BlockQueueConsumer {
	return &BlockQueueConsumer{
		queue:     queue,
		processor: processor,
		logger:    logger,
	}
}

func (c *BlockQueueConsumer) Start(ctx context.Context) {
	if c == nil {
		return
	}
	if c.queue == nil || !c.queue.consumerEnabled || c.queue.client == nil {
		if c.logger != nil {
			c.logger.Info("block sqs consumer disabled")
		}
		return
	}
	go c.run(ctx)
}

func (c *BlockQueueConsumer) run(ctx context.Context) {
	c.logger.Info("block sqs consumer started")
	for {
		if ctx.Err() != nil {
			c.logger.Info("block sqs consumer stopped")
			return
		}
		if err := c.poll(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("block sqs consumer poll failed", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func (c *BlockQueueConsumer) poll(ctx context.Context) error {
	queueURL, err := c.queue.ensureQueue(ctx)
	if err != nil {
		return err
	}
	output, err := c.queue.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     c.queue.waitTimeSeconds,
		VisibilityTimeout:   c.queue.visibilityTimeout,
	})
	if err != nil {
		return err
	}
	for _, message := range output.Messages {
		if err := c.handleMessage(ctx, queueURL, aws.ToString(message.Body), aws.ToString(message.ReceiptHandle)); err != nil {
			c.logger.Error("failed process block event message", zap.Error(err))
		}
	}
	return nil
}

func (c *BlockQueueConsumer) handleMessage(ctx context.Context, queueURL string, body string, receiptHandle string) error {
	var event entity.BlockEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		if receiptHandle != "" {
			if deleteErr := c.deleteMessage(ctx, queueURL, receiptHandle); deleteErr != nil {
				return fmt.Errorf("delete invalid block message after decode failure: %w", deleteErr)
			}
		}
		return fmt.Errorf("invalid block event message body: %w", err)
	}

	if c.processor == nil {
		return errors.New("block event processor is required")
	}
	if err := c.processor.ProcessBlockEvent(ctx, &event); err != nil {
		return err
	}
	if receiptHandle != "" {
		if err := c.deleteMessage(ctx, queueURL, receiptHandle); err != nil {
			return err
		}
	}
	return nil
}

func (c *BlockQueueConsumer) deleteMessage(ctx context.Context, queueURL string, receiptHandle string) error {
	_, err := c.queue.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

func (q *BlockQueue) ensureQueue(ctx context.Context) (string, error) {
	if q == nil || q.client == nil {
		return "", errors.New("sqs client is not configured")
	}
	q.queueMu.Lock()
	defer q.queueMu.Unlock()

	if q.queueURL != "" {
		return q.queueURL, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	output, err := q.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(q.cfg.BlockQueueName),
	})
	if err != nil {
		createOutput, createErr := q.client.CreateQueue(ctx, &sqs.CreateQueueInput{
			QueueName: aws.String(q.cfg.BlockQueueName),
		})
		if createErr != nil {
			return "", fmt.Errorf("get or create queue %s: %w", q.cfg.BlockQueueName, createErr)
		}
		q.queueURL = aws.ToString(createOutput.QueueUrl)
	} else {
		q.queueURL = aws.ToString(output.QueueUrl)
	}

	if q.queueURL == "" {
		return "", errors.New("sqs returned an empty block queue url")
	}
	q.logger.Info("block sqs queue ready", zap.String("queue_url", q.queueURL))
	return q.queueURL, nil
}
