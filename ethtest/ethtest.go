// Copyright (c) 2021 the ethier authors (github.com/divergencetech/ethier)

// The ethtest package provides helpers for testing Ethereum smart contracts.
package ethtest

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// A SimulatedBackend embeds a go-ethereum SimulatedBackend and extends its
// functionality to simplify standard testing.
type SimulatedBackend struct {
	*backends.SimulatedBackend
	AutoCommit  bool
	PrivateKeys []*ecdsa.PrivateKey
}

// NewSimulatedBackend returns a new simulated Ethereum backend with the specified
// number of accounts. Transactions are automatically committed. Close()
// must be called to free resources after use.
func NewSimulatedBackend(numAccounts int) (*SimulatedBackend, error) {
	keys := make([]*ecdsa.PrivateKey, numAccounts)
	alloc := make(core.GenesisAlloc)

	for i := 0; i < numAccounts; i++ {
		privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("unable to generate private key: %w", err)
		}

		keys[i] = privateKey

		alloc[crypto.PubkeyToAddress(privateKey.PublicKey)] = core.GenesisAccount{
			Balance: big.NewInt(0).Mul(big.NewInt(100), big.NewInt(1e18)),
		}
	}

	sb := SimulatedBackend{
		SimulatedBackend: backends.NewSimulatedBackend(alloc, 3e7),
		AutoCommit:       true,
		PrivateKeys:      keys,
	}

	sb.AdjustTime(365 * 24 * time.Hour)
	sb.Commit()

	return &sb, nil
}

// SendTransaction functions pipes its parameters to the embedded backend and
// also calls Commit() if sb.AutoCommit==true.
func (sb *SimulatedBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if err := sb.SimulatedBackend.SendTransaction(ctx, tx); err != nil {
		return err
	}

	if sb.AutoCommit {
		sb.Commit()
	}

	return nil
}

// NewTxOpts returns a TransactOpts for the account.
func (sb *SimulatedBackend) NewTxOpts(account int) (*bind.TransactOpts, error) {
	txOpts, err := bind.NewKeyedTransactorWithChainID(sb.PrivateKeys[account], big.NewInt(1337))
	if err != nil {
		return nil, fmt.Errorf("unable to create a transaction opts: %w", err)
	}

	return txOpts, nil
}

// CallOpts returns a CallOpts for the account.
func (sb *SimulatedBackend) CallOpts(account int) *bind.CallOpts {
	return &bind.CallOpts{
		From: crypto.PubkeyToAddress(sb.PrivateKeys[account].PublicKey),
	}
}
