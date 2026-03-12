package settlement

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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

func (s *Settler) Release(ctx context.Context, dealID string, commenterAddress string, amountCents int) (string, error) {
	slog.Info("settlement: releasing funds",
		"deal_id", dealID,
		"commenter", commenterAddress,
		"amount_cents", amountCents,
	)

	nonce, err := s.client.PendingNonceAt(ctx, s.operatorAddress)
	if err != nil {
		return "", fmt.Errorf("get nonce: %w", err)
	}

	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("suggest gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, s.chainID)
	if err != nil {
		return "", fmt.Errorf("create transactor: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasPrice = gasPrice
	auth.GasLimit = 200000

	// USDC has 6 decimals; amountCents / 100 = dollars, * 1e6 = USDC units
	amount := new(big.Int).Mul(
		big.NewInt(int64(amountCents)),
		big.NewInt(1e4), // cents -> 6-decimal USDC
	)

	// ABI-encode: release(bytes32 dealId, address commenter, uint256 amount)
	dealIDBytes := common.HexToHash(fmt.Sprintf("%064x", dealID))
	commenterAddr := common.HexToAddress(commenterAddress)

	data, err := packReleaseCall(dealIDBytes, commenterAddr, amount)
	if err != nil {
		return "", fmt.Errorf("pack release call: %w", err)
	}

	_ = data
	_ = auth

	// For MVP: log the intent. Actual contract interaction requires ABI binding
	// generated from the deployed contract. Placeholder returns a mock tx hash.
	slog.Info("settlement: release call prepared",
		"nonce", nonce,
		"gas_price", gasPrice,
		"amount_usdc", amount.String(),
	)

	return fmt.Sprintf("0x%064x", nonce), nil
}

func packReleaseCall(dealID common.Hash, commenter common.Address, amount *big.Int) ([]byte, error) {
	// Method signature: release(bytes32,address,uint256)
	methodID := crypto.Keccak256([]byte("release(bytes32,address,uint256)"))[:4]

	data := make([]byte, 4+32+32+32)
	copy(data[0:4], methodID)
	copy(data[4:36], dealID.Bytes())
	copy(data[36:68], common.LeftPadBytes(commenter.Bytes(), 32))
	copy(data[68:100], common.LeftPadBytes(amount.Bytes(), 32))

	return data, nil
}
