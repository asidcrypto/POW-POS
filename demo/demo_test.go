package demo

import (
	"errors"
	"testing"
)

func TestBlockCommitsWithSupermajorityStake(t *testing.T) {
	chain, err := NewTendermintL2Chain("test-chain", []Validator{
		{Name: "alice", Stake: 40},
		{Name: "bob", Stake: 35},
		{Name: "carol", Stake: 25},
	})
	if err != nil {
		t.Fatalf("new chain: %v", err)
	}

	chain.QueueTransaction("tx-1")
	block, err := chain.ProposeBlock([]string{"alice", "bob"}, 100)
	if err != nil {
		t.Fatalf("propose block: %v", err)
	}
	if !block.HasSupermajority() {
		t.Fatalf("expected supermajority block")
	}
	if block.Height != 1 {
		t.Fatalf("expected height 1, got %d", block.Height)
	}
}

func TestBlockRejectedWithoutSupermajority(t *testing.T) {
	chain, err := NewTendermintL2Chain("test-chain", []Validator{
		{Name: "alice", Stake: 40},
		{Name: "bob", Stake: 35},
		{Name: "carol", Stake: 25},
	})
	if err != nil {
		t.Fatalf("new chain: %v", err)
	}

	chain.QueueTransaction("tx-1")
	_, err = chain.ProposeBlock([]string{"alice", "carol"}, 100) // 65%, not > 2/3
	if !errors.Is(err, ErrConsensusFailure) {
		t.Fatalf("expected ErrConsensusFailure, got %v", err)
	}
}

func TestCheckpointFinalityAfterConfirmations(t *testing.T) {
	cfg := RunConfig{
		L2Blocks:             6,
		CheckpointInterval:   3,
		RequiredPoWConfirms:  2,
		MineBtcEveryL2Blocks: 1,
	}
	l2, btc, bridge, _, err := RunDemo(cfg)
	if err != nil {
		t.Fatalf("run demo: %v", err)
	}

	if _, ok := bridge.AnchorTxIDByHeight[3]; !ok {
		t.Fatalf("missing checkpoint at height 3")
	}
	if _, ok := bridge.AnchorTxIDByHeight[6]; !ok {
		t.Fatalf("missing checkpoint at height 6")
	}
	if bridge.PoWConfirmationsForHeight(3, btc) < 2 {
		t.Fatalf("expected >=2 confirmations for height 3")
	}
	if bridge.PoWConfirmationsForHeight(6, btc) < 2 {
		t.Fatalf("expected >=2 confirmations for height 6")
	}
	if !bridge.IsPoWFinal(3, btc) || !bridge.IsPoWFinal(6, btc) {
		t.Fatalf("expected heights 3 and 6 to be PoW-final")
	}
	if bridge.IsPoWFinal(5, btc) {
		t.Fatalf("height 5 should not be PoW-final")
	}
	if bridge.PoWConfirmationsForHeight(5, btc) != 0 {
		t.Fatalf("height 5 should have zero PoW confirmations")
	}

	report := BuildSecurityReport(l2, btc, bridge)
	var found bool
	for _, row := range report {
		if row.L2Height == 6 {
			found = true
			if !row.AnchoredToPoW || !row.PoWFinal {
				t.Fatalf("expected height 6 anchored and PoW-final, got %+v", row)
			}
		}
	}
	if !found {
		t.Fatalf("missing report row for height 6")
	}
}

func TestBridgeValidatesInputs(t *testing.T) {
	if _, err := NewCheckpointBridge(0, 2); err == nil {
		t.Fatalf("expected error for checkpoint interval 0")
	}
	if _, err := NewCheckpointBridge(3, 0); err == nil {
		t.Fatalf("expected error for required confirmations 0")
	}
}
