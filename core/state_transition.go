// Copyright 2018 TesraSupernet Foundation Ltd
// Copyright 2014 The go-ethereum Authors
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
	"errors"
	"math/big"

	"crypto/ecdsa"
	"strings"

	"github.com/TesraSupernet/TesraMainChain/accounts/abi"
	"github.com/TesraSupernet/TesraMainChain/common"
	"github.com/TesraSupernet/TesraMainChain/common/math"
	"github.com/TesraSupernet/TesraMainChain/core/types"
	"github.com/TesraSupernet/TesraMainChain/core/vm"
	"github.com/TesraSupernet/TesraMainChain/crypto"
	"github.com/TesraSupernet/TesraMainChain/params"
	"github.com/TesraSupernet/TesraMainChain/pos/incentive"
	"github.com/TesraSupernet/TesraMainChain/pos/util"
)

var (
	Big0                         = big.NewInt(0)
	errInsufficientBalanceForGas = errors.New("insufficient balance to pay for gas")
)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay gas
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==
  4a) Attempt to run transaction data
  4b) If valid, use result as code for the new state object
== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp         *GasPool
	msg        Message
	gas        uint64
	gasPrice   *big.Int
	initialGas *big.Int
	value      *big.Int
	data       []byte
	state      vm.StateDB
	evm        *vm.EVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	//FromFrontier() (common.Address, error)
	To() *common.Address

	GasPrice() *big.Int
	Gas() *big.Int
	Value() *big.Int

	Nonce() uint64
	CheckNonce() bool
	Data() []byte

	TxType() uint64
}

// IntrinsicGas computes the 'intrinsic gas' for a message
// with the given data.
//
// TODO convert to uint64
func IntrinsicGas(data []byte, to *common.Address, homestead bool) *big.Int {
	contractCreation := to == nil

	igas := new(big.Int)
	if contractCreation && homestead {
		igas.SetUint64(params.TxGasContractCreation)
	} else {
		igas.SetUint64(params.TxGas)
	}
	if len(data) > 0 {
		var nz int64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		m := big.NewInt(nz)
		m.Mul(m, new(big.Int).SetUint64(params.TxDataNonZeroGas))
		igas.Add(igas, m)
		m.SetInt64(int64(len(data)) - nz)
		m.Mul(m, new(big.Int).SetUint64(params.TxDataZeroGas))
		igas.Add(igas, m)
	}

	// reduce gas used for pos tx
	if vm.IsPosPrecompiledAddr(to) {
		igas = igas.Div(igas, big.NewInt(10))
	}

	return igas
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(evm *vm.EVM, msg Message, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:         gp,
		evm:        evm,
		msg:        msg,
		gasPrice:   msg.GasPrice(),
		initialGas: new(big.Int),
		value:      msg.Value(),
		data:       msg.Data(),
		state:      evm.StateDB,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) ([]byte, *big.Int, bool, error) {
	st := NewStateTransition(evm, msg, gp)

	ret, _, gasUsed, failed, err := st.TransitionDb()
	return ret, gasUsed, failed, err
}

func (st *StateTransition) from() vm.AccountRef {
	f := st.msg.From()
	if !st.state.Exist(f) {
		st.state.CreateAccount(f)
	}
	return vm.AccountRef(f)
}

func (st *StateTransition) to() vm.AccountRef {
	if st.msg == nil {
		return vm.AccountRef{}
	}
	to := st.msg.To()
	if to == nil {
		return vm.AccountRef{} // contract creation
	}

	reference := vm.AccountRef(*to)
	if !st.state.Exist(*to) {
		st.state.CreateAccount(*to)
	}
	return reference
}

func (st *StateTransition) useGas(amount uint64) error {
	if st.gas < amount {
		return vm.ErrOutOfGas
	}
	st.gas -= amount

	return nil
}

func (st *StateTransition) buyGas() error {
	mgas := st.msg.Gas()
	if mgas.BitLen() > 64 {
		return vm.ErrOutOfGas
	}

	mgval := new(big.Int).Mul(mgas, st.gasPrice)

	var (
		state  = st.state
		sender = st.from()
	)

	if state.GetBalance(sender.Address()).Cmp(mgval) < 0 {
		return errInsufficientBalanceForGas
	}

	if err := st.gp.SubGas(mgas); err != nil {
		return err
	}
	st.gas += mgas.Uint64()

	st.initialGas.Set(mgas)
	state.SubBalance(sender.Address(), mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	msg := st.msg
	sender := st.from()

	// Make sure this transaction's nonce is correct
	if msg.CheckNonce() {
		nonce := st.state.GetNonce(sender.Address())
		if nonce < msg.Nonce() {
			return ErrNonceTooHigh
		} else if nonce > msg.Nonce() {
			return ErrNonceTooLow
		}
	}
	return st.buyGas()
}

// TransitionDb will transition the state by applying the current message and returning the result
// including the required gas for the operation as well as the used gas. It returns an error if it
// failed. An error indicates a consensus issue.
func (st *StateTransition) TransitionDb() (ret []byte, requiredGas, usedGas *big.Int, failed bool, err error) {

	if types.IsNormalTransaction(st.msg.TxType()) || types.IsPosTransaction(st.msg.TxType()) {
		if err = st.preCheck(); err != nil {
			return
		}
	}

	//log.Trace("after preCheck", "txType", st.msg.TxType(), "gas pool", st.gp.String())

	msg := st.msg
	sender := st.from() // err checked in preCheck

	//homestead := st.evm.ChainConfig().IsHomestead(st.evm.BlockNumber)
	contractCreation := msg.To() == nil

	// Pay intrinsic gas
	// TODO convert to uint64
	intrinsicGas := IntrinsicGas(st.data, msg.To(), true /*homestead*/)
	//log.Trace("get intrinsic gas", "gas", intrinsicGas.String())
	if intrinsicGas.BitLen() > 64 {
		return nil, nil, nil, false, vm.ErrOutOfGas
	}

	var stampTotalGas uint64
	if types.IsPrivacyTransaction(st.msg.TxType()) {
		pureCallData, totalUseableGas, evmUseableGas, err := PreProcessPrivacyTx(st.evm.StateDB,
			sender.Address().Bytes(),
			st.data, st.gasPrice, st.value)
		if err != nil {
			return nil, nil, nil, false, err
		}

		stampTotalGas = totalUseableGas
		st.gas = evmUseableGas
		st.initialGas.SetUint64(evmUseableGas)
		st.data = pureCallData[:]
		//log.Trace("pre process privacy tx", "stampTotalGas", stampTotalGas, "evmUseableGas", evmUseableGas)
		//sub gas from total gas of curent block,prevent gas is overhead gaslimit
		if err := st.gp.SubGas(new(big.Int).SetUint64(totalUseableGas)); err != nil {
			return nil, nil, nil, false, err
		}
	}

	if err = st.useGas(intrinsicGas.Uint64()); err != nil {
		return nil, nil, nil, false, err
	}

	//log.Trace("subed intrinsic gas", "gas pool left", st.gp.String())

	var (
		evm = st.evm
		// vm errors do not effect consensus and are therefor
		// not assigned to err, except for insufficient balance
		// error.
		vmerr error
	)
	if contractCreation {
		ret, _, st.gas, vmerr = evm.Create(sender, st.data, st.gas, st.value)
		//log.Trace("create contract", "left gas", st.gas, "err", vmerr)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(sender.Address(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = evm.Call(sender, st.to().Address(), st.data, st.gas, st.value)
		//log.Trace("no create contract", "left gas", st.gas, "err", vmerr)
	}

	if vmerr != nil {
		//log.Debug("VM returned with error", "err", vmerr)
		// The only possible consensus-error would be if there wasn't
		// sufficient balance to make the transfer happen. The first
		// balance transfer may never fail.
		if vmerr == vm.ErrInsufficientBalance {
			return nil, nil, nil, false, vmerr
		}
	}

	if types.IsNormalTransaction(st.msg.TxType()) || types.IsPosTransaction(st.msg.TxType()) {
		requiredGas = st.gasUsed()
		st.refundGas()
		usedGas = new(big.Int).Set(st.gasUsed())
		//log.Trace("calc used gas, normal tx", "required gas", requiredGas, "used gas", usedGas)
	} else if types.IsPrivacyTransaction(st.msg.TxType()) {
		requiredGas = new(big.Int).SetUint64(stampTotalGas)
		usedGas = requiredGas
		//log.Trace("calc used gas, pos tx", "required gas", requiredGas, "used gas", usedGas)
	}

	if !params.IsPosActive() {
		st.state.AddBalance(st.evm.Coinbase, new(big.Int).Mul(usedGas, st.gasPrice))
	} else {
		epochID, _ := util.GetEpochSlotIDFromDifficulty(st.evm.Context.Difficulty)
		incentive.AddEpochGas(st.state, new(big.Int).Mul(usedGas, st.gasPrice), epochID)
	}
	return ret, requiredGas, usedGas, vmerr != nil, err
}

func (st *StateTransition) refundGas() {
	// Return eth for remaining gas to the sender account,
	// exchanged at the original rate.
	sender := st.from() // err already checked
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(sender.Address(), remaining)

	// Apply refund counter, capped to half of the used gas.
	uhalf := remaining.Div(st.gasUsed(), common.Big2)
	refund := math.BigMin(uhalf, st.state.GetRefund())
	st.gas += refund.Uint64()

	st.state.AddBalance(sender.Address(), refund.Mul(refund, st.gasPrice))

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(new(big.Int).SetUint64(st.gas))
}

func (st *StateTransition) gasUsed() *big.Int {
	return new(big.Int).Sub(st.initialGas, new(big.Int).SetUint64(st.gas))
}

///////////////////////added for privacy tx /////////////////////////////
var (
	utilAbiDefinition = `[{"constant":false,"type":"function","inputs":[{"name":"RingSignedData","type":"string"},{"name":"CxtCallParams","type":"bytes"}],"name":"combine","outputs":[{"name":"RingSignedData","type":"string"},{"name":"CxtCallParams","type":"bytes"}]}]`

	utilAbi, errAbiInit = abi.JSON(strings.NewReader(utilAbiDefinition))

	TokenAbi = utilAbi
)

func init() {
	if errAbiInit != nil {
		panic(errAbiInit)
	}
}

type PrivacyTxInfo struct {
	PublicKeys         []*ecdsa.PublicKey
	KeyImage           *ecdsa.PublicKey
	W_Random           []*big.Int
	Q_Random           []*big.Int
	CallData           []byte
	StampBalance       *big.Int
	StampTotalGas      uint64
	GasLeftSubRingSign uint64
}

func FetchPrivacyTxInfo(stateDB vm.StateDB, hashInput []byte, in []byte, gasPrice *big.Int) (info *PrivacyTxInfo, err error) {
	if len(in) < 4 {
		return nil, vm.ErrInvalidRingSigned
	}

	var TxDataWithRing struct {
		RingSignedData string
		CxtCallParams  []byte
	}

	err = utilAbi.Unpack(&TxDataWithRing, "combine", in[4:])
	if err != nil {
		return
	}

	ringSignInfo, err := vm.FetchRingSignInfo(stateDB, hashInput, TxDataWithRing.RingSignedData)
	if err != nil {
		return
	}

	stampGasBigInt := new(big.Int).Div(ringSignInfo.OTABalance, gasPrice)
	if stampGasBigInt.BitLen() > 64 {
		return nil, vm.ErrOutOfGas
	}

	StampTotalGas := stampGasBigInt.Uint64()
	mixLen := len(ringSignInfo.PublicKeys)
	ringSigDiffRequiredGas := params.RequiredGasPerMixPub * (uint64(mixLen))

	// ringsign compute gas + ota image key store setting gas
	preSubGas := ringSigDiffRequiredGas + params.SstoreSetGas
	if StampTotalGas < preSubGas {
		return nil, vm.ErrOutOfGas
	}

	GasLeftSubRingSign := StampTotalGas - preSubGas
	info = &PrivacyTxInfo{
		ringSignInfo.PublicKeys,
		ringSignInfo.KeyImage,
		ringSignInfo.W_Random,
		ringSignInfo.Q_Random,
		TxDataWithRing.CxtCallParams[:],
		ringSignInfo.OTABalance,
		StampTotalGas,
		GasLeftSubRingSign,
	}

	return
}

func ValidPrivacyTx(stateDB vm.StateDB, hashInput []byte, in []byte, gasPrice *big.Int,
	intrGas *big.Int, txValue *big.Int, gasLimit *big.Int) error {
	if intrGas == nil || intrGas.BitLen() > 64 {
		return vm.ErrOutOfGas
	}

	if txValue.Sign() != 0 {
		return vm.ErrInvalidPrivacyValue
	}

	if gasPrice == nil || gasPrice.Cmp(common.Big0) <= 0 {
		return vm.ErrInvalidGasPrice
	}

	info, err := FetchPrivacyTxInfo(stateDB, hashInput, in, gasPrice)
	if err != nil {
		return err
	}

	if info.StampTotalGas > gasLimit.Uint64() {
		return ErrGasLimit
	}

	kix := crypto.FromECDSAPub(info.KeyImage)
	exist, _, err := vm.CheckOTAImageExist(stateDB, kix)
	if err != nil {
		return err
	} else if exist {
		return errors.New("stamp has been spended")
	}

	if info.GasLeftSubRingSign < intrGas.Uint64() {
		return vm.ErrOutOfGas
	}

	return nil
}

func PreProcessPrivacyTx(stateDB vm.StateDB, hashInput []byte, in []byte, gasPrice *big.Int, txValue *big.Int) (callData []byte, totalUseableGas uint64, evmUseableGas uint64, err error) {
	if txValue.Sign() != 0 {
		return nil, 0, 0, vm.ErrInvalidPrivacyValue
	}

	info, err := FetchPrivacyTxInfo(stateDB, hashInput, in, gasPrice)
	if err != nil {
		return nil, 0, 0, err
	}

	kix := crypto.FromECDSAPub(info.KeyImage)
	exist, _, err := vm.CheckOTAImageExist(stateDB, kix)
	if err != nil || exist {
		return nil, 0, 0, err
	}

	vm.AddOTAImage(stateDB, kix, info.StampBalance.Bytes())

	return info.CallData, info.StampTotalGas, info.GasLeftSubRingSign, nil
}
