package slotleader

import (
	"fmt"
	"github.com/TesraSupernet/TesraMainChain/pos/posconfig"
	"testing"

	"github.com/TesraSupernet/TesraMainChain/accounts/keystore"

	"github.com/TesraSupernet/TesraMainChain/consensus/ethash"
	"github.com/TesraSupernet/TesraMainChain/core"
	"github.com/TesraSupernet/TesraMainChain/core/vm"
	"github.com/TesraSupernet/TesraMainChain/ethdb"
	"github.com/TesraSupernet/TesraMainChain/rpc"
)

var s *SLS

func testInitSlotleader() {
	SlsInit()
	s = GetSlotLeaderSelection()

	// Create the database in memory or in a temporary directory.
	db, _ := ethdb.NewMemDatabase()
	gspec := core.DefaultPPOWTestingGenesisBlock()
	gspec.MustCommit(db)

	ce := ethash.NewFaker(db)
	bc, _ := core.NewBlockChain(db, gspec.Config, ce, vm.Config{},nil)

	s.Init(bc, &rpc.Client{}, &keystore.Key{})

	s.sendTransactionFn = testSender

}

func TestGetCurrentStateDb(t *testing.T) {

	posconfig.SelfTestMode = true
	testInitSlotleader()

	posconfig.SelfTestMode = false
	stateDb, err := s.GetCurrentStateDb()
	if err != nil || stateDb == nil {
		t.FailNow()
	}

	epochID := s.getLastEpochIDFromChain()
	slotID := s.getLastSlotIDFromChain()
	number := s.getBlockChainHeight()
	if number != 0 || epochID != 0 || slotID != 0 {
		t.FailNow()
	}

	fmt.Println(epochID, slotID)
	RmDB("epochGendb")
	posconfig.SelfTestMode = false
}
