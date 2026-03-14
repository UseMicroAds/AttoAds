package settlement

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
)

type Settler struct {
	client          *ethclient.Client
	privateKey      *ecdsa.PrivateKey
	operatorAddress common.Address
	escrowAddress   common.Address
	chainID         *big.Int
}

func NewSettler(rpcURL, privateKeyHex, escrowContractAddr string, chainID int64) (*Settler, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}

	pk, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	addr := crypto.PubkeyToAddress(pk.PublicKey)

	return &Settler{
		client:          client,
		privateKey:      pk,
		operatorAddress: addr,
		escrowAddress:   common.HexToAddress(escrowContractAddr),
		chainID:         big.NewInt(chainID),
	}, nil
}

func (s *Settler) Release(ctx context.Context, dealID string, campaignID string, commenterAddress string, amountCents int) (string, error) {
	slog.Info("settlement: releasing funds",
		"deal_id", dealID,
		"campaign_id", campaignID,
		"commenter", commenterAddress,
		"amount_cents", amountCents,
		"operator", s.operatorAddress.Hex(),
		"escrow", s.escrowAddress.Hex(),
		"chain_id", s.chainID.String(),
	)

	nonce, err := s.client.PendingNonceAt(ctx, s.operatorAddress)
	if err != nil {
		slog.Error("settlement: failed to get pending nonce", "operator", s.operatorAddress.Hex(), "error", err)
		return "", fmt.Errorf("get nonce: %w", err)
	}
	slog.Info("settlement: fetched pending nonce", "nonce", nonce, "operator", s.operatorAddress.Hex())

	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		slog.Error("settlement: failed to suggest gas price", "error", err)
		return "", fmt.Errorf("suggest gas price: %w", err)
	}
	slog.Info("settlement: suggested gas price", "gas_price_wei", gasPrice.String())

	// USDC has 6 decimals; amountCents / 100 = dollars, * 1e6 = USDC units
	amount := new(big.Int).Mul(
		big.NewInt(int64(amountCents)),
		big.NewInt(1e4), // cents -> 6-decimal USDC
	)

	// ABI-encode: release(bytes32 dealId, bytes32 campaignId, address commenter, uint256 amount)
	dealIDBytes, err := uuidStringToBytes32(dealID)
	if err != nil {
		slog.Error("settlement: invalid deal id", "deal_id", dealID, "error", err)
		return "", fmt.Errorf("parse deal id: %w", err)
	}
	// Must match frontend funding flow: keccak256(stringToHex(campaign.id))
	campaignIDBytes := stringKeccakToBytes32(campaignID)
	commenterAddr := common.HexToAddress(commenterAddress)
	slog.Info("settlement: normalized release params",
		"deal_id_bytes32", dealIDBytes.Hex(),
		"campaign_id_bytes32", campaignIDBytes.Hex(),
		"commenter_hex", commenterAddr.Hex(),
		"amount_usdc_6dec", amount.String(),
	)

	data, err := packReleaseCall(dealIDBytes, campaignIDBytes, commenterAddr, amount)
	if err != nil {
		slog.Error("settlement: failed to pack calldata", "error", err)
		return "", fmt.Errorf("pack release call: %w", err)
	}
	slog.Info("settlement: packed release calldata",
		"calldata_len", len(data),
		"method_id", fmt.Sprintf("0x%x", data[:4]),
		"calldata_hex", common.Bytes2Hex(data),
	)

	callMsg := ethereum.CallMsg{
		From:     s.operatorAddress,
		To:       &s.escrowAddress,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     data,
	}
	gasLimit, err := s.client.EstimateGas(ctx, callMsg)
	if err != nil {
		slog.Error("settlement: gas estimation failed",
			"from", s.operatorAddress.Hex(),
			"to", s.escrowAddress.Hex(),
			"gas_price_wei", gasPrice.String(),
			"calldata_hex", common.Bytes2Hex(data),
			"error", err,
		)
		return "", fmt.Errorf("estimate gas: %w", err)
	}
	slog.Info("settlement: gas estimated", "gas_limit", gasLimit)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &s.escrowAddress,
		Value:    big.NewInt(0),
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(s.chainID), s.privateKey)
	if err != nil {
		slog.Error("settlement: failed to sign tx", "nonce", nonce, "error", err)
		return "", fmt.Errorf("sign transaction: %w", err)
	}
	slog.Info("settlement: signed release tx",
		"tx_hash", signedTx.Hash().Hex(),
		"nonce", signedTx.Nonce(),
		"to", signedTx.To().Hex(),
		"gas_limit", signedTx.Gas(),
		"gas_price_wei", signedTx.GasPrice().String(),
	)

	if err := s.client.SendTransaction(ctx, signedTx); err != nil {
		slog.Error("settlement: failed to broadcast tx",
			"tx_hash", signedTx.Hash().Hex(),
			"nonce", signedTx.Nonce(),
			"error", err,
		)
		return "", fmt.Errorf("send transaction: %w", err)
	}

	slog.Info("settlement: release transaction submitted",
		"tx_hash", signedTx.Hash().Hex(),
		"nonce", nonce,
		"gas_price", gasPrice,
		"gas_limit", gasLimit,
		"amount_usdc", amount.String(),
	)

	return signedTx.Hash().Hex(), nil
}

func (s *Settler) TransactionReceiptStatus(ctx context.Context, txHash string) (mined bool, success bool, err error) {
	if txHash == "" {
		return false, false, fmt.Errorf("tx hash is required")
	}

	receipt, err := s.client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("fetch tx receipt: %w", err)
	}

	return true, receipt.Status == types.ReceiptStatusSuccessful, nil
}

func packReleaseCall(dealID common.Hash, campaignID common.Hash, commenter common.Address, amount *big.Int) ([]byte, error) {
	// Method signature: release(bytes32,bytes32,address,uint256)
	methodID := crypto.Keccak256([]byte("release(bytes32,bytes32,address,uint256)"))[:4]

	data := make([]byte, 4+32+32+32+32)
	copy(data[0:4], methodID)
	copy(data[4:36], dealID.Bytes())
	copy(data[36:68], campaignID.Bytes())
	copy(data[68:100], common.LeftPadBytes(commenter.Bytes(), 32))
	copy(data[100:132], common.LeftPadBytes(amount.Bytes(), 32))

	return data, nil
}

func uuidStringToBytes32(value string) (common.Hash, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return common.Hash{}, err
	}
	return common.BytesToHash(id[:]), nil
}

func stringKeccakToBytes32(value string) common.Hash {
	return crypto.Keccak256Hash([]byte(value))
}
