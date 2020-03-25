// Copyright 2018 TesraSupernet Foundation Ltd

package vm

import (
	"math/big"

	"github.com/TesraSupernet/TesraMainChain/common"
	"github.com/TesraSupernet/TesraMainChain/core/types"
)

// Precompiled contracts address or
// Reserved contracts address.
// Should prevent overwriting to them.
var (
	ecrecoverPrecompileAddr      = common.BytesToAddress([]byte{1})
	sha256hashPrecompileAddr     = common.BytesToAddress([]byte{2})
	ripemd160hashPrecompileAddr  = common.BytesToAddress([]byte{3})
	dataCopyPrecompileAddr       = common.BytesToAddress([]byte{4})
	bigModExpPrecompileAddr      = common.BytesToAddress([]byte{5})
	bn256AddPrecompileAddr       = common.BytesToAddress([]byte{6})
	bn256ScalarMulPrecompileAddr = common.BytesToAddress([]byte{7})
	bn256PairingPrecompileAddr   = common.BytesToAddress([]byte{8})

	tsrCoinPrecompileAddr  = common.BytesToAddress([]byte{100})
	tsrStampPrecompileAddr = common.BytesToAddress([]byte{200})

	TsrCscPrecompileAddr  = common.BytesToAddress([]byte{218})
	StakersInfoAddr       = common.BytesToAddress(big.NewInt(400).Bytes())
	StakingCommonAddr     = common.BytesToAddress(big.NewInt(401).Bytes())
	StakersFeeAddr        = common.BytesToAddress(big.NewInt(402).Bytes())
	StakersMaxFeeAddr     = common.BytesToAddress(big.NewInt(403).Bytes())
	otaBalanceStorageAddr = common.BytesToAddress(big.NewInt(300).Bytes())
	otaImageStorageAddr   = common.BytesToAddress(big.NewInt(301).Bytes())

	// 0.01tsr --> "0x0000000000000000000000010000000000000000"
	otaBalancePercentdot001WStorageAddr = common.HexToAddress(TsrStampdot001)
	otaBalancePercentdot002WStorageAddr = common.HexToAddress(TsrStampdot002)
	otaBalancePercentdot005WStorageAddr = common.HexToAddress(TsrStampdot005)

	otaBalancePercentdot003WStorageAddr = common.HexToAddress(TsrStampdot003)
	otaBalancePercentdot006WStorageAddr = common.HexToAddress(TsrStampdot006)
	otaBalancePercentdot009WStorageAddr = common.HexToAddress(TsrStampdot009)

	otaBalancePercentdot03WStorageAddr = common.HexToAddress(TsrStampdot03)
	otaBalancePercentdot06WStorageAddr = common.HexToAddress(TsrStampdot06)
	otaBalancePercentdot09WStorageAddr = common.HexToAddress(TsrStampdot09)
	otaBalancePercentdot2WStorageAddr  = common.HexToAddress(TsrStampdot2)
	otaBalancePercentdot5WStorageAddr  = common.HexToAddress(TsrStampdot5)

	otaBalance10WStorageAddr  = common.HexToAddress(Tsrcoin10)
	otaBalance20WStorageAddr  = common.HexToAddress(Tsrcoin20)
	otaBalance50WStorageAddr  = common.HexToAddress(Tsrcoin50)
	otaBalance100WStorageAddr = common.HexToAddress(Tsrcoin100)

	otaBalance200WStorageAddr   = common.HexToAddress(Tsrcoin200)
	otaBalance500WStorageAddr   = common.HexToAddress(Tsrcoin500)
	otaBalance1000WStorageAddr  = common.HexToAddress(Tsrcoin1000)
	otaBalance5000WStorageAddr  = common.HexToAddress(Tsrcoin5000)
	otaBalance50000WStorageAddr = common.HexToAddress(Tsrcoin50000)

	//pos
	slotLeaderPrecompileAddr = common.BytesToAddress(big.NewInt(600).Bytes())

	IncentivePrecompileAddr = common.BytesToAddress(big.NewInt(606).Bytes()) //0x25E

	randomBeaconPrecompileAddr = common.BytesToAddress(big.NewInt(610).Bytes())
	PosControlPrecompileAddr   = common.BytesToAddress(big.NewInt(612).Bytes())

	// TODO: remove one?
	RandomBeaconPrecompileAddr = randomBeaconPrecompileAddr
	SlotLeaderPrecompileAddr   = slotLeaderPrecompileAddr
)

// PrecompiledContract is the basic interface for native Go contracts. The implementation
// requires a deterministic gas count based on the input size of the Run method of the
// contract.
type PrecompiledContract interface {
	RequiredGas(input []byte) uint64                                // RequiredPrice calculates the contract gas use
	Run(input []byte, contract *Contract, evm *EVM) ([]byte, error) // Run runs the precompiled contract
	ValidTx(stateDB StateDB, signer types.Signer, tx *types.Transaction) error
}

// PrecompiledContractsHomestead contains the default set of pre-compiled Ethereum
// contracts used in the Frontier and Homestead releases.
var PrecompiledContractsHomestead = map[common.Address]PrecompiledContract{
	ecrecoverPrecompileAddr:     &ecrecover{},
	sha256hashPrecompileAddr:    &sha256hash{},
	ripemd160hashPrecompileAddr: &ripemd160hash{},
	dataCopyPrecompileAddr:      &dataCopy{},

	tsrCoinPrecompileAddr:  &tsrCoinSC{},
	tsrStampPrecompileAddr: &tesramainchainStampSC{},
}

// PrecompiledContractsByzantium contains the default set of pre-compiled Ethereum
// contracts used in the Byzantium release.
var PrecompiledContractsByzantium = map[common.Address]PrecompiledContract{
	ecrecoverPrecompileAddr:      &ecrecover{},
	sha256hashPrecompileAddr:     &sha256hash{},
	ripemd160hashPrecompileAddr:  &ripemd160hash{},
	dataCopyPrecompileAddr:       &dataCopy{},
	bigModExpPrecompileAddr:      &bigModExp{},
	bn256AddPrecompileAddr:       &bn256Add{},
	bn256ScalarMulPrecompileAddr: &bn256ScalarMul{},
	bn256PairingPrecompileAddr:   &bn256Pairing{},

	tsrCoinPrecompileAddr:  &tsrCoinSC{},
	tsrStampPrecompileAddr: &tesramainchainStampSC{},

	//pos
	TsrCscPrecompileAddr:       &PosStaking{},
	PosControlPrecompileAddr:   &PosControl{},
	slotLeaderPrecompileAddr:   &slotLeaderSC{},
	randomBeaconPrecompileAddr: &RandomBeaconContract{},
}

func IsPosPrecompiledAddr(addr *common.Address) bool {
	if addr == nil {
		return false
	}

	if (*addr) == slotLeaderPrecompileAddr ||
		(*addr) == IncentivePrecompileAddr ||
		(*addr) == randomBeaconPrecompileAddr {
		return true
	}

	return false
}
