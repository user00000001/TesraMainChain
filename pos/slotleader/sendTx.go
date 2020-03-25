package slotleader

import (
	"errors"
	"math/big"

	"github.com/TesraSupernet/TesraMainChain/common/hexutil"
	"github.com/TesraSupernet/TesraMainChain/core"
	"github.com/TesraSupernet/TesraMainChain/core/types"
	"github.com/TesraSupernet/TesraMainChain/core/vm"
	"github.com/TesraSupernet/TesraMainChain/log"
	"github.com/TesraSupernet/TesraMainChain/rpc"
)

var (
	errRCNotReady = errors.New("rc is not ready")
)

type SendTxFn func(rc *rpc.Client, tx map[string]interface{})

func (s *SLS) sendSlotTx(payload []byte, posSender SendTxFn) error {
	if s.rc == nil {
		return errRCNotReady
	}

	to := vm.GetSlotLeaderSCAddress()
	data := hexutil.Bytes(payload)
	gas := core.IntrinsicGas(data, &to, true)

	arg := map[string]interface{}{}
	arg["from"] = s.key.Address
	arg["to"] = vm.GetSlotLeaderSCAddress()
	arg["value"] = (*hexutil.Big)(big.NewInt(0))
	arg["gas"] = (*hexutil.Big)(gas)
	arg["txType"] = types.POS_TX
	arg["data"] = data
	log.Debug("Write data of payload", "length", len(data))

	go posSender(s.rc, arg)
	return nil
}
