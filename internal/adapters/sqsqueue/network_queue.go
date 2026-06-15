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

type NetworkQueue struct {
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

var _ ports.NetworkProducerPort = (*NetworkQueue)(nil)

func NewNetworkQueue(cfg *appconfig.Config, logger *zap.Logger) (*NetworkQueue, error) {
	if cfg == nil {
		cfg = &appconfig.Config{}
	}

	queue := &NetworkQueue{
		cfg:               cfg,
		logger:            logger,
		queueURL:          cfg.NetworkQueueURL,
		producerEnabled:   cfg.SQSEnabled,
		consumerEnabled:   cfg.SQSConsumerEnabled,
		waitTimeSeconds:   cfg.SQSWaitTimeSeconds,
		visibilityTimeout: cfg.SQSVisibilityTimeout,
	}

	if !cfg.SQSEnabled && !cfg.SQSConsumerEnabled {
		logger.Info("sqs disabled")
		return queue, nil
	}
	if cfg.NetworkQueueName == "" && cfg.NetworkQueueURL == "" {
		return nil, errors.New("NETWORK_QUEUE_NAME or NETWORK_QUEUE_URL is required when SQS is enabled")
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

func (q *NetworkQueue) Enabled() bool {
	return q != nil && q.producerEnabled && q.client != nil
}

func (q *NetworkQueue) EnqueueNetworkRegistration(ctx context.Context, network *entity.Network) error {
	if !q.Enabled() {
		return errors.New("network queue producer is disabled")
	}
	if network == nil {
		return errors.New("network is required")
	}

	queueURL, err := q.ensureQueue(ctx)
	if err != nil {
		return err
	}

	body, err := json.Marshal(network)
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

	q.logger.Info("network registration sent to sqs", zap.String("network_id", network.ID), zap.String("queue_url", queueURL))
	return nil
}

type NetworkConsumer struct {
	queue  *NetworkQueue
	repo   ports.NetworkRepositoryPort
	logger *zap.Logger
}

func NewNetworkConsumer(queue *NetworkQueue, repo ports.NetworkRepositoryPort, logger *zap.Logger) *NetworkConsumer {
	return &NetworkConsumer{
		queue:  queue,
		repo:   repo,
		logger: logger,
	}
}

func (c *NetworkConsumer) Start(ctx context.Context) {
	if c == nil {
		return
	}
	if c.queue == nil || !c.queue.consumerEnabled || c.queue.client == nil {
		if c.logger != nil {
			c.logger.Info("network sqs consumer disabled")
		}
		return
	}
	go c.run(ctx)
}

func (c *NetworkConsumer) run(ctx context.Context) {
	c.logger.Info("network sqs consumer started")

	for {
		if ctx.Err() != nil {
			c.logger.Info("network sqs consumer stopped")
			return
		}

		if err := c.poll(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("network sqs consumer poll failed", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func (c *NetworkConsumer) poll(ctx context.Context) error {
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
			c.logger.Error("failed process network registration message", zap.Error(err))
		}
	}
	return nil
}

func (c *NetworkConsumer) handleMessage(ctx context.Context, queueURL string, body string, receiptHandle string) error {
	var network entity.Network
	if err := json.Unmarshal([]byte(body), &network); err != nil {
		if receiptHandle != "" {
			if deleteErr := c.deleteMessage(ctx, queueURL, receiptHandle); deleteErr != nil {
				return fmt.Errorf("delete invalid message after decode failure: %w", deleteErr)
			}
		}
		return fmt.Errorf("invalid network message body: %w", err)
	}

	if err := c.repo.SaveNetwork(ctx, &network); err != nil {
		return err
	}

	if receiptHandle != "" {
		if err := c.deleteMessage(ctx, queueURL, receiptHandle); err != nil {
			return err
		}
	}

	c.logger.Info("network registration consumed", zap.String("network_id", network.ID), zap.String("name", network.Name), zap.Int64("chain_id", network.ChainID))
	return nil
}

func (c *NetworkConsumer) deleteMessage(ctx context.Context, queueURL string, receiptHandle string) error {
	_, err := c.queue.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

func (q *NetworkQueue) ensureQueue(ctx context.Context) (string, error) {
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
		QueueName: aws.String(q.cfg.NetworkQueueName),
	})
	if err != nil {
		createOutput, createErr := q.client.CreateQueue(ctx, &sqs.CreateQueueInput{
			QueueName: aws.String(q.cfg.NetworkQueueName),
		})
		if createErr != nil {
			return "", fmt.Errorf("get or create queue %s: %w", q.cfg.NetworkQueueName, createErr)
		}
		q.queueURL = aws.ToString(createOutput.QueueUrl)
	} else {
		q.queueURL = aws.ToString(output.QueueUrl)
	}

	if q.queueURL == "" {
		return "", errors.New("sqs returned an empty queue url")
	}
	q.logger.Info("network sqs queue ready", zap.String("queue_url", q.queueURL))
	return q.queueURL, nil
}
