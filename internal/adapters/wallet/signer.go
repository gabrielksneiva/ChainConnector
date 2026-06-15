package wallet

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type privateKeySigner struct {
	privateKey string
}

func NewSignerFromPrivateKey(privateKeyHex string) (ports.WalletSignerPort, error) {
	key := strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x")
	if key == "" {
		return nil, errors.New("missing private key")
	}
	if _, err := hex.DecodeString(key); err != nil {
		return nil, err
	}
	return &privateKeySigner{privateKey: key}, nil
}

func (s *privateKeySigner) Address(ctx context.Context) (string, error) {
	key, err := crypto.HexToECDSA(s.privateKey)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(key.PublicKey).Hex(), nil
}

func (s *privateKeySigner) SignHash(ctx context.Context, hash []byte) ([]byte, error) {
	key, err := crypto.HexToECDSA(s.privateKey)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(hash, key)
}

func (s *privateKeySigner) SignTransaction(ctx context.Context, tx *entity.Transaction) ([]byte, error) {
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	key, err := crypto.HexToECDSA(s.privateKey)
	if err != nil {
		return nil, err
	}
	var ethTx *types.Transaction
	if tx.MaxFeePerGas != nil || tx.MaxPriorityFeePerGas != nil {
		ethTx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   tx.ChainID,
			Nonce:     tx.Nonce,
			GasTipCap: tx.MaxPriorityFeePerGas,
			GasFeeCap: tx.MaxFeePerGas,
			Gas:       tx.Gas,
			To:        toAddress(tx.To),
			Value:     tx.Value,
			Data:      tx.Data,
		})
	} else {
		ethTx = types.NewTx(&types.LegacyTx{
			Nonce:    tx.Nonce,
			GasPrice: tx.GasPrice,
			Gas:      tx.Gas,
			To:       toAddress(tx.To),
			Value:    tx.Value,
			Data:     tx.Data,
		})
	}
	signer := types.LatestSignerForChainID(tx.ChainID)
	signed, err := types.SignTx(ethTx, signer, key)
	if err != nil {
		return nil, err
	}
	return signed.MarshalBinary()
}

func toAddress(addr *string) *common.Address {
	if addr == nil || *addr == "" {
		return nil
	}
	a := common.HexToAddress(*addr)
	return &a
}
