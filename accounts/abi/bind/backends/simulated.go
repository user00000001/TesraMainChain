// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package backends

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	ethereum "github.com/TesraSupernet/TesraMainChain"
	"github.com/TesraSupernet/TesraMainChain/accounts/abi/bind"
	"github.com/TesraSupernet/TesraMainChain/common"
	"github.com/TesraSupernet/TesraMainChain/common/math"
	"github.com/TesraSupernet/TesraMainChain/consensus/ethash"
	"github.com/TesraSupernet/TesraMainChain/core"
	"github.com/TesraSupernet/TesraMainChain/core/state"
	"github.com/TesraSupernet/TesraMainChain/core/types"
	"github.com/TesraSupernet/TesraMainChain/core/vm"
	"github.com/TesraSupernet/TesraMainChain/crypto"
	"github.com/TesraSupernet/TesraMainChain/ethdb"
	"github.com/TesraSupernet/TesraMainChain/params"
)

// This nil assignment ensures compile time that SimulatedBackend implements bind.ContractBackend.
var _ bind.ContractBackend = (*SimulatedBackend)(nil)

var errBlockNumberUnsupported = errors.New("SimulatedBackend cannot access blocks other than the latest block")

var (
	key, _      = crypto.HexToECDSA("f1572f76b75b40a7da72d6f2ee7fda3d1189c2d28f0a2f096347055abe344d7f")
	coinbase    = crypto.PubkeyToAddress(key.PublicKey)
	extraVanity = 32
	extraSeal   = 65
)

// SimulatedBackend implements bind.ContractBackend, simulating a blockchain in
// the background. Its main purpose is to allow easily testing contract bindings.
type SimulatedBackend struct {
	// database   ethdb.Database   // In memory database to store our testing data
	// blockchain *core.BlockChain // Ethereum blockchain to handle the consensus

	mu           sync.Mutex
	pendingBlock *types.Block   // Currently pending block that will be imported on request
	pendingState *state.StateDB // Currently pending state that will be the active on on request

	env *core.ChainEnv

	BlockEnv *core.ChainEnv
	// config *params.ChainConfig
}

// NewSimulatedBackend creates a new binding backend using a simulated blockchain
// for testing purposes.
func NewSimulatedBackend() *SimulatedBackend {
	db, _ := ethdb.NewMemDatabase()
	gspec := core.DefaultPPOWTestingGenesisBlock()
	gspec.MustCommit(db)

	ce := ethash.NewFaker(db)
	bc, _ := core.NewBlockChain(db, gspec.Config, ce, vm.Config{}, nil)
	env := core.NewChainEnv(gspec.Config, gspec, ce, bc, db)

	backend := &SimulatedBackend{env: env}
	backend.BlockEnv = env
	backend.rollback()
	return backend
}

func NewSimulatedBackendEx(alloc core.GenesisAlloc) *SimulatedBackend {
	db, _ := ethdb.NewMemDatabase()
	gspec := core.DefaultPPOWTestingGenesisBlock()
	for k, v := range alloc {
		gspec.Alloc[k] = v
	}
	gspec.MustCommit(db)

	ce := ethash.NewFaker(db)
	bc, _ := core.NewBlockChain(db, gspec.Config, ce, vm.Config{}, nil)
	env := core.NewChainEnv(gspec.Config, gspec, ce, bc, db)

	backend := &SimulatedBackend{env: env}
	backend.rollback()
	return backend
}

// Commit imports all the pending transactions as a single block and starts a
// fresh new state.
func (b *SimulatedBackend) Commit() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, err := b.env.Blockchain().InsertChain([]*types.Block{b.pendingBlock}); err != nil {
		panic(err) // This cannot happen unless the simulator is wrong, fail in that case
	}
	b.rollback()
}

// Rollback aborts all pending transactions, reverting to the last committed state.
func (b *SimulatedBackend) Rollback() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.rollback()
}

func (b *SimulatedBackend) rollback() {
	blocks, _ := b.env.GenerateChain(b.env.Blockchain().CurrentBlock(), 1, nil)
	b.pendingBlock = blocks[0]
	b.pendingState, _ = state.New(b.pendingBlock.Root(), state.NewDatabase(b.env.Database()))
}

// CodeAt returns the code associated with a certain account in the blockchain.
func (b *SimulatedBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if blockNumber != nil && blockNumber.Cmp(b.env.Blockchain().CurrentBlock().Number()) != 0 {
		return nil, errBlockNumberUnsupported
	}
	statedb, _ := b.env.Blockchain().State()
	return statedb.GetCode(contract), nil
}

// BalanceAt returns the wei balance of a certain account in the blockchain.
func (b *SimulatedBackend) BalanceAt(ctx context.Context, contract common.Address, blockNumber *big.Int) (*big.Int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if blockNumber != nil && blockNumber.Cmp(b.env.Blockchain().CurrentBlock().Number()) != 0 {
		return nil, errBlockNumberUnsupported
	}
	statedb, _ := b.env.Blockchain().State()
	return statedb.GetBalance(contract), nil
}

// NonceAt returns the nonce of a certain account in the blockchain.
func (b *SimulatedBackend) NonceAt(ctx context.Context, contract common.Address, blockNumber *big.Int) (uint64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if blockNumber != nil && blockNumber.Cmp(b.env.Blockchain().CurrentBlock().Number()) != 0 {
		return 0, errBlockNumberUnsupported
	}
	statedb, _ := b.env.Blockchain().State()
	return statedb.GetNonce(contract), nil
}

// StorageAt returns the value of key in the storage of an account in the blockchain.
func (b *SimulatedBackend) StorageAt(ctx context.Context, contract common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if blockNumber != nil && blockNumber.Cmp(b.env.Blockchain().CurrentBlock().Number()) != 0 {
		return nil, errBlockNumberUnsupported
	}
	statedb, _ := b.env.Blockchain().State()
	val := statedb.GetState(contract, key)
	return val[:], nil
}

// TransactionReceipt returns the receipt of a transaction.
func (b *SimulatedBackend) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	receipt, _, _, _ := core.GetReceipt(b.env.Database(), txHash)
	return receipt, nil
}

// PendingCodeAt returns the code associated with an account in the pending state.
func (b *SimulatedBackend) PendingCodeAt(ctx context.Context, contract common.Address) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.pendingState.GetCode(contract), nil
}

// CallContract executes a contract call.
func (b *SimulatedBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if blockNumber != nil && blockNumber.Cmp(b.env.Blockchain().CurrentBlock().Number()) != 0 {
		return nil, errBlockNumberUnsupported
	}
	state, err := b.env.Blockchain().State()
	if err != nil {
		return nil, err
	}
	rval, _, _, err := b.callContract(ctx, call, b.env.Blockchain().CurrentBlock(), state)
	return rval, err
}

// PendingCallContract executes a contract call on the pending state.
// func (b *SimulatedBackend) PendingCallContract(ctx context.Context, call ethereum.CallMsg) ([]byte, error) {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()
// 	defer b.pendingState.RevertToSnapshot(b.pendingState.Snapshot())

// 	rval, _, _, err := b.callContract(ctx, call, b.pendingBlock, b.pendingState)
// 	return rval, err
// }

// PendingNonceAt implements PendingStateReader.PendingNonceAt, retrieving
// the nonce currently pending for the account.
func (b *SimulatedBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.pendingState.GetOrNewStateObject(account).Nonce(), nil
}

// SuggestGasPrice implements ContractTransactor.SuggestGasPrice. Since the simulated
// chain doens't have miners, we just return a gas price of 1 for any call.
func (b *SimulatedBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}

// EstimateGas executes the requested code against the currently pending block/state and
// returns the used amount of gas.
func (b *SimulatedBackend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (*big.Int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo uint64 = params.TxGas - 1
		hi uint64
	)
	if call.Gas != nil && call.Gas.Uint64() >= params.TxGas {
		hi = call.Gas.Uint64()
	} else {
		hi = b.pendingBlock.GasLimit().Uint64()
	}
	for lo+1 < hi {
		// Take a guess at the gas, and check transaction validity
		mid := (hi + lo) / 2
		call.Gas = new(big.Int).SetUint64(mid)

		snapshot := b.pendingState.Snapshot()
		_, _, failed, err := b.callContract(ctx, call, b.pendingBlock, b.pendingState)
		b.pendingState.RevertToSnapshot(snapshot)

		// If the transaction became invalid or execution failed, raise the gas limit
		if err != nil || failed {
			lo = mid
			continue
		}
		// Otherwise assume the transaction succeeded, lower the gas limit
		hi = mid
	}
	return new(big.Int).SetUint64(hi), nil
}

// callContract implemens common code between normal and pending contract calls.
// state is modified during execution, make sure to copy it if necessary.
func (b *SimulatedBackend) callContract(ctx context.Context, call ethereum.CallMsg, block *types.Block, statedb *state.StateDB) ([]byte, *big.Int, bool, error) {
	// Ensure message is initialized properly.
	if call.GasPrice == nil {
		call.GasPrice = big.NewInt(1)
	}
	if call.Gas == nil || call.Gas.Sign() == 0 {
		call.Gas = big.NewInt(50000000)
	}
	if call.Value == nil {
		call.Value = new(big.Int)
	}
	// Set infinite balance to the fake caller account.
	from := statedb.GetOrNewStateObject(call.From)
	from.SetBalance(math.MaxBig256)
	// Execute the call.
	msg := callmsg{call}

	evmContext := core.NewEVMContext(msg, block.Header(), b.env.Blockchain(), nil)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(evmContext, statedb, b.env.Config(), vm.Config{})
	gaspool := new(core.GasPool).AddGas(math.MaxBig256)
	ret, gasUsed, _, failed, err := core.NewStateTransition(vmenv, msg, gaspool).TransitionDb()
	return ret, gasUsed, failed, err
}

// SendTransaction updates the pending block to include the given transaction.
// It panics if the transaction is invalid.
func (b *SimulatedBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	sender, err := types.Sender(types.NewEIP155Signer(big.NewInt(1)), tx)
	if err != nil {
		panic(fmt.Errorf("invalid transaction: %v", err))
	}
	nonce := b.pendingState.GetNonce(sender)
	if tx.Nonce() != nonce {
		panic(fmt.Errorf("invalid transaction nonce: got %d, want %d", tx.Nonce(), nonce))
	}

	blocks, _ := b.env.GenerateChain(b.env.Blockchain().CurrentBlock(), 1, func(number int, block *core.BlockGen) {
		for _, tx := range b.pendingBlock.Transactions() {
			block.AddTx(tx)
		}
		block.AddTx(tx)
	})

	b.pendingBlock = blocks[0]
	b.pendingState, _ = state.New(b.pendingBlock.Root(), state.NewDatabase(b.env.Database()))
	return nil
}

// JumpTimeInSeconds adds skip seconds to the clock
func (b *SimulatedBackend) AdjustTime(adjustment time.Duration) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	blocks, _ := b.env.GenerateChain(b.env.Blockchain().CurrentBlock(), 1, func(number int, block *core.BlockGen) {
		for _, tx := range b.pendingBlock.Transactions() {
			block.AddTx(tx)
		}
		block.OffsetTime(int64(adjustment.Seconds()))
	})
	b.pendingBlock = blocks[0]
	b.pendingState, _ = state.New(b.pendingBlock.Root(), state.NewDatabase(b.env.Database()))

	return nil
}

// callmsg implements core.Message to allow passing it as a transaction simulator.
type callmsg struct {
	ethereum.CallMsg
}

func (m callmsg) From() common.Address { return m.CallMsg.From }
func (m callmsg) Nonce() uint64        { return 0 }
func (m callmsg) CheckNonce() bool     { return false }
func (m callmsg) To() *common.Address  { return m.CallMsg.To }
func (m callmsg) GasPrice() *big.Int   { return m.CallMsg.GasPrice }
func (m callmsg) Gas() *big.Int        { return m.CallMsg.Gas }
func (m callmsg) Value() *big.Int      { return m.CallMsg.Value }
func (m callmsg) Data() []byte         { return m.CallMsg.Data }
func (m callmsg) TxType() uint64       { return m.CallMsg.TxType }
