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

package core

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/TesraSupernet/TesraMainChain/accounts"
	"github.com/TesraSupernet/TesraMainChain/common"
	"github.com/TesraSupernet/TesraMainChain/consensus"
	"github.com/TesraSupernet/TesraMainChain/consensus/ethash"
	"github.com/TesraSupernet/TesraMainChain/core/state"
	"github.com/TesraSupernet/TesraMainChain/core/types"
	"github.com/TesraSupernet/TesraMainChain/core/vm"
	"github.com/TesraSupernet/TesraMainChain/crypto"
	"github.com/TesraSupernet/TesraMainChain/ethdb"
	"github.com/TesraSupernet/TesraMainChain/params"
)

// So we can deterministically seed different blockchains
var (
	canonicalSeed             = 1
	forkSeed                  = 2
	fakedAddr                 = common.HexToAddress("0xf9b32578b4420a36f132db32b56f3831a7cc1804")
	fakedAccountPrivateKey, _ = crypto.HexToECDSA("f1572f76b75b40a7da72d6f2ee7fda3d1189c2d28f0a2f096347055abe344d7f")
	extraVanity               = 32
	extraSeal                 = 65
)

func fakeSignerFn(signer accounts.Account, hash []byte) ([]byte, error) {
	return crypto.Sign(hash, fakedAccountPrivateKey)
}

type ChainEnv struct {
	config       *params.ChainConfig
	genesis      *Genesis
	engine       *ethash.Ethash
	blockChain   *BlockChain
	db           ethdb.Database
	mapSigners   map[common.Address]struct{}
	arraySigners []common.Address
	signerKeys   []*ecdsa.PrivateKey
	//set signers here for testing convenience
	// validSigners
}

// add for testing permission proof of work
var (
	totalSigner = 20
	signerSet   = make(map[common.Address]*ecdsa.PrivateKey)
	addrSigners = make([]common.Address, 0)
)

func init() {
	for i := 0; i < totalSigner; i++ {
		private, _ := crypto.GenerateKey()
		addr := crypto.PubkeyToAddress(private.PublicKey)
		signerSet[addr] = private
		addrSigners = append(addrSigners, addr)
	}
}

func NewChainEnv(config *params.ChainConfig, g *Genesis, engine consensus.Engine, bc *BlockChain, db ethdb.Database) *ChainEnv {
	ce := &ChainEnv{
		config:       config,
		genesis:      g,
		engine:       engine.(*ethash.Ethash),
		blockChain:   bc,
		db:           db,
		mapSigners:   make(map[common.Address]struct{}),
		arraySigners: make([]common.Address, len(g.ExtraData)/common.AddressLength),
	}

	for i := 0; i < len(ce.arraySigners); i++ {
		copy(ce.arraySigners[i][:], g.ExtraData[i*common.AddressLength:])
	}
	for _, s := range ce.arraySigners {
		ce.mapSigners[s] = struct{}{}
	}

	return ce
}

// BlockGen creates blocks for testing.
// See GenerateChain for a detailed explanation.
type BlockGen struct {
	i       int
	parent  *types.Block
	chain   []*types.Block
	header  *types.Header
	statedb *state.StateDB

	gasPool  *GasPool
	txs      []*types.Transaction
	receipts []*types.Receipt
	uncles   []*types.Header

	config *params.ChainConfig
}

// SetCoinbase sets the coinbase of the generated block.
// It can be called at most once.
func (b *BlockGen) SetCoinbase(addr common.Address) {
	if b.gasPool != nil {
		if len(b.txs) > 0 {
			panic("coinbase must be set before adding transactions")
		}
		panic("coinbase can only be set once")
	}
	b.header.Coinbase = addr
	b.gasPool = new(GasPool).AddGas(b.header.GasLimit)
}

// SetExtra sets the extra data field of the generated block.
func (b *BlockGen) SetExtra(data []byte) {
	// ensure the extra data has all its components
	l := len(data)
	if l > extraVanity {
		fmt.Println("extra data too long")
		return
	}

	copy(b.header.Extra[extraVanity-l:extraVanity], data)
}

// AddTx adds a transaction to the generated block. If no coinbase has
// been set, the block's coinbase is set to the zero address.
//
// AddTx panics if the transaction cannot be executed. In addition to
// the protocol-imposed limitations (gas limit, etc.), there are some
// further limitations on the content of transactions that can be
// added. Notably, contract code relying on the BLOCKHASH instruction
// will panic during execution.
func (b *BlockGen) AddTx(tx *types.Transaction) {
	if b.gasPool == nil {
		b.SetCoinbase(b.parent.Coinbase())
	}
	b.statedb.Prepare(tx.Hash(), common.Hash{}, len(b.txs))
	receipt, _, err := ApplyTransaction(b.config, nil, &b.header.Coinbase, b.gasPool, b.statedb, b.header, tx, b.header.GasUsed, vm.Config{})
	if err != nil {
		panic(err)
	}
	b.txs = append(b.txs, tx)
	b.receipts = append(b.receipts, receipt)
}

func (b *BlockGen) AddTxAndCalcGasUsed(tx *types.Transaction) *big.Int {
	if b.gasPool == nil {
		b.SetCoinbase(b.parent.Coinbase())
	}
	b.statedb.Prepare(tx.Hash(), common.Hash{}, len(b.txs))
	receipt, gasUsed, err := ApplyTransaction(b.config, nil, &b.header.Coinbase, b.gasPool, b.statedb, b.header, tx, b.header.GasUsed, vm.Config{})
	if err != nil {
		panic(err)
	}
	b.txs = append(b.txs, tx)
	b.receipts = append(b.receipts, receipt)
	return gasUsed
}

// Number returns the block number of the block being generated.
func (b *BlockGen) Number() *big.Int {
	return new(big.Int).Set(b.header.Number)
}

// AddUncheckedReceipt forcefully adds a receipts to the block without a
// backing transaction.
//
// AddUncheckedReceipt will cause consensus failures when used during real
// chain processing. This is best used in conjunction with raw block insertion.
func (b *BlockGen) AddUncheckedReceipt(receipt *types.Receipt) {
	b.receipts = append(b.receipts, receipt)
}

// TxNonce returns the next valid transaction nonce for the
// account at addr. It panics if the account does not exist.
func (b *BlockGen) TxNonce(addr common.Address) uint64 {
	if !b.statedb.Exist(addr) {
		panic("account does not exist")
	}
	return b.statedb.GetNonce(addr)
}

// AddUncle adds an uncle header to the generated block.
func (b *BlockGen) AddUncle(h *types.Header) {
	b.uncles = append(b.uncles, h)
}

// PrevBlock returns a previously generated block by number. It panics if
// num is greater or equal to the number of the block being generated.
// For index -1, PrevBlock returns the parent block given to GenerateChain.
func (b *BlockGen) PrevBlock(index int) *types.Block {
	if index >= b.i {
		panic("block index out of range")
	}
	if index == -1 {
		return b.parent
	}
	return b.chain[index]
}

// OffsetTime modifies the time instance of a block, implicitly changing its
// associated difficulty. It's useful to test scenarios where forking is not
// tied to chain length directly.
func (b *BlockGen) OffsetTime(seconds int64) {
	b.header.Time.Add(b.header.Time, new(big.Int).SetInt64(seconds))
	if b.header.Time.Cmp(b.parent.Header().Time) <= 0 {
		panic("block time out of range")
	}
	b.header.Difficulty = ethash.CalcDifficulty(b.config, b.header.Time.Uint64(), b.parent.Header())
}

// GenerateChain creates a chain of n blocks. The first block's
// parent will be the provided parent. db is used to store
// intermediate states and should contain the parent's state trie.
//
// The generator function is called with a new block generator for
// every block. Any transactions and uncles added to the generator
// become part of the block. If gen is nil, the blocks will be empty
// and their coinbase will be the zero address.
//
// Blocks created by GenerateChain do not contain valid proof of work
// values. Inserting them into BlockChain requires use of FakePow or
// a similar non-validating proof of work implementation.
func (self *ChainEnv) GenerateChain(parent *types.Block, n int, gen func(int, *BlockGen)) ([]*types.Block, []types.Receipts) {
	blocks, receipts := make(types.Blocks, n), make([]types.Receipts, n)
	genblock := func(i int, h *types.Header, statedb *state.StateDB) (*types.Block, types.Receipts) {
		b := &BlockGen{parent: parent, i: i, chain: blocks, header: h, statedb: statedb, config: self.config}

		// Execute any user modifications to the block and finalize it
		if gen != nil {
			gen(i, b)
		}

		ethash.AccumulateRewards(self.config, statedb, h, b.uncles)
		root, err := statedb.CommitTo(self.db, true)
		if err != nil {
			panic(fmt.Sprintf("state write error: %v", err))
		}
		h.Root = root

		self.engine.Authorize(fakedAddr, fakeSignerFn)
		h.Coinbase.Set(fakedAddr)
		rawBlock := types.NewBlock(h, b.txs, b.uncles, b.receipts)
		sealBlock, _ := self.engine.Seal(self.blockChain, rawBlock, nil)
		return sealBlock, b.receipts
	}
	for i := 0; i < n; i++ {
		statedb, err := state.New(parent.Root(), state.NewDatabase(self.db))
		if err != nil {
			panic(err)
		}
		header := makeHeader(self.config, parent, statedb)
		block, receipt := genblock(i, header, statedb)
		blocks[i] = block
		receipts[i] = receipt
		parent = block
	}
	return blocks, receipts
}

func (self *ChainEnv) GenerateChainMulti(parent *types.Block, n int, gen func(int, *BlockGen)) ([]*types.Block, []types.Receipts) {
	blocks, receipts := make(types.Blocks, n), make([]types.Receipts, n)
	genblock := func(i int, h *types.Header, statedb *state.StateDB) (*types.Block, types.Receipts) {
		b := &BlockGen{parent: parent, i: i, chain: blocks, header: h, statedb: statedb, config: self.config}

		// Execute any user modifications to the block and finalize it
		if gen != nil {
			gen(i, b)
		}

		ethash.AccumulateRewards(self.config, statedb, h, b.uncles)
		root, err := statedb.CommitTo(self.db, true)
		if err != nil {
			panic(fmt.Sprintf("state write error: %v", err))
		}
		h.Root = root

		self.engine.Authorize(fakedAddr, fakeSignerFn)
		rawBlock := types.NewBlock(h, b.txs, b.uncles, b.receipts)
		sealBlock, _ := self.engine.Seal(self.blockChain, rawBlock, nil)
		return sealBlock, b.receipts
	}
	for i := 0; i < n; i++ {
		statedb, err := state.New(parent.Root(), state.NewDatabase(self.db))
		if err != nil {
			panic(err)
		}
		header := makeHeader(self.config, parent, statedb)
		block, receipt := genblock(i, header, statedb)
		blocks[i] = block
		receipts[i] = receipt
		parent = block
	}
	return blocks, receipts
}

func fakeSignerFnEx(signer accounts.Account, hash []byte) ([]byte, error) {
	return crypto.Sign(hash, signerSet[signer.Address])
}

func (self *ChainEnv) GenerateChainEx(parent *types.Block, signerSequence []int, gen func(int, *BlockGen)) ([]*types.Block, []types.Receipts) {
	blocks, receipts := make(types.Blocks, len(signerSequence)), make([]types.Receipts, len(signerSequence))
	genblock := func(i int, h *types.Header, statedb *state.StateDB) (*types.Block, types.Receipts) {
		b := &BlockGen{parent: parent, i: i, chain: blocks, header: h, statedb: statedb, config: self.config}

		// Execute any user modifications to the block and finalize it
		if gen != nil {
			gen(i, b)
		}

		ethash.AccumulateRewards(self.config, statedb, h, b.uncles)
		root, err := statedb.CommitTo(self.db, true)
		if err != nil {
			panic(fmt.Sprintf("state write error: %v", err))
		}
		h.Root = root

		self.engine.Authorize(addrSigners[i], fakeSignerFnEx)
		h.Coinbase.Set(addrSigners[i])
		rawBlock := types.NewBlock(h, b.txs, b.uncles, b.receipts)
		sealBlock, _ := self.engine.Seal(self.blockChain, rawBlock, nil)
		return sealBlock, b.receipts
	}
	for i := 0; i < len(signerSequence); i++ {
		statedb, err := state.New(parent.Root(), state.NewDatabase(self.db))
		if err != nil {
			panic(err)
		}
		header := makeHeader(self.config, parent, statedb)
		block, receipt := genblock(signerSequence[i], header, statedb)
		blocks[i] = block
		receipts[i] = receipt
		parent = block
	}
	return blocks, receipts
}

func makeHeader(config *params.ChainConfig, parent *types.Block, state *state.StateDB) *types.Header {
	var time *big.Int
	if parent.Time() == nil {
		time = big.NewInt(10)
	} else {
		time = new(big.Int).Add(parent.Time(), big.NewInt(10)) // block time is fixed at 10 seconds
	}

	return &types.Header{
		Root:       state.IntermediateRoot(true /*config.IsEIP158(parent.Number())*/),
		ParentHash: parent.Hash(),
		Coinbase:   parent.Coinbase(),
		Difficulty: ethash.CalcDifficulty(config, time.Uint64(), &types.Header{
			Number:     parent.Number(),
			Time:       new(big.Int).Sub(time, big.NewInt(10)),
			Difficulty: parent.Difficulty(),
			UncleHash:  parent.UncleHash(),
		}),
		GasLimit: CalcGasLimit(parent),
		GasUsed:  new(big.Int),
		Number:   new(big.Int).Add(parent.Number(), common.Big1),
		Time:     time,
		Extra:    make([]byte, 97),
	}
}

// newCanonical creates a chain database, and injects a deterministic canonical
// chain. Depending on the full flag, if creates either a full block chain or a
// header only chain.
func newCanonical(n int, full bool) (ethdb.Database, *BlockChain, error, *ChainEnv) {
	// Initialize a fresh chain with only a genesis block
	gspec := DefaultPPOWTestingGenesisBlock()
	db, _ := ethdb.NewMemDatabase()
	genesis := gspec.MustCommit(db)
	engine := ethash.NewFaker(db)

	blockchain, _ := NewBlockChain(db, params.TestChainConfig, engine, vm.Config{}, nil)
	chainEnv := NewChainEnv(params.TestChainConfig, gspec, engine, blockchain, db)
	// Create and inject the requested chain
	if n == 0 {
		return db, blockchain, nil, chainEnv
	}
	if full {
		// Full block-chain requested
		blocks := chainEnv.makeBlockChain(genesis, n, canonicalSeed)
		_, err := blockchain.InsertChain(blocks)
		return db, blockchain, err, chainEnv
	}
	// Header-only chain requested
	headers := chainEnv.makeHeaderChain(genesis.Header(), n, canonicalSeed)
	_, err := blockchain.InsertHeaderChain(headers, 1)
	return db, blockchain, err, chainEnv
}

// makeHeaderChain creates a deterministic chain of headers rooted at parent.
func (self *ChainEnv) makeHeaderChain(parent *types.Header, n int, seed int) []*types.Header {
	blocks := self.makeBlockChain(types.NewBlockWithHeader(parent), n, seed)
	headers := make([]*types.Header, len(blocks))
	for i, block := range blocks {
		headers[i] = block.Header()
	}
	return headers
}

// makeBlockChain creates a deterministic chain of blocks rooted at parent.
func (self *ChainEnv) makeBlockChain(parent *types.Block, n int, seed int) []*types.Block {
	// blocks, _ := self.GenerateChain(parent, n, func(i int, b *BlockGen) {
	// b.SetCoinbase(common.Address{0: byte(seed), 19: byte(i)})
	// })
	blocks, _ := self.GenerateChain(parent, n, nil)
	return blocks
}

func (self *ChainEnv) Blockchain() *BlockChain {
	return self.blockChain
}

func (self *ChainEnv) Database() ethdb.Database {
	return self.blockChain.chainDb
}

func (self *ChainEnv) Config() *params.ChainConfig {
	return self.config
}
