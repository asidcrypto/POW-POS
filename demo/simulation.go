package demo

import (
	"fmt"
)

type DemoEvent struct {
	Step    int
	Message string
}

type SecurityRow struct {
	L2Height         int64
	L2Hash           string
	PoSFinal         bool
	AnchoredToPoW    bool
	PoWConfirmations int64
	PoWFinal         bool
}

func DefaultValidators() []Validator {
	return []Validator{
		{Name: "alice", Stake: 40},
		{Name: "bob", Stake: 30},
		{Name: "carol", Stake: 20},
		{Name: "dave", Stake: 10},
	}
}

type RunConfig struct {
	L2Blocks             int64
	CheckpointInterval   int64
	RequiredPoWConfirms  int64
	MineBtcEveryL2Blocks int64
}

func DefaultRunConfig() RunConfig {
	return RunConfig{
		L2Blocks:             12,
		CheckpointInterval:   3,
		RequiredPoWConfirms:  2,
		MineBtcEveryL2Blocks: 2,
	}
}

func RunDemo(cfg RunConfig) (*TendermintL2Chain, *BitcoinPoWChain, *CheckpointBridge, []DemoEvent, error) {
	if cfg.L2Blocks <= 0 {
		return nil, nil, nil, nil, fmt.Errorf("L2Blocks must be positive")
	}
	if cfg.MineBtcEveryL2Blocks <= 0 {
		return nil, nil, nil, nil, fmt.Errorf("MineBtcEveryL2Blocks must be positive")
	}

	l2, err := NewTendermintL2Chain("pow-pos-demo", DefaultValidators())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	btc := NewBitcoinPoWChain("bitcoin-regtest-demo")
	bridge, err := NewCheckpointBridge(cfg.CheckpointInterval, cfg.RequiredPoWConfirms)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	events := make([]DemoEvent, 0, cfg.L2Blocks*3)
	step := 1

	for blockIndex := int64(1); blockIndex <= cfg.L2Blocks; blockIndex++ {
		l2.QueueTransaction(fmt.Sprintf("user_tx_%03d_a", blockIndex))
		l2.QueueTransaction(fmt.Sprintf("user_tx_%03d_b", blockIndex))
		block, err := l2.ProposeBlock(nil, 100)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		events = append(events, DemoEvent{
			Step: step,
			Message: fmt.Sprintf(
				"L2 block %d committed (%d/%d stake, hash=%s)",
				block.Height,
				block.SignedStake,
				block.TotalStake,
				ShortHash(block.BlockHash, 10),
			),
		})
		step++

		anchor := bridge.MaybeAnchorLatestBlock(l2, btc)
		if anchor != nil {
			events = append(events, DemoEvent{
				Step: step,
				Message: fmt.Sprintf(
					"Checkpoint submitted for L2 block %d (txid=%s)",
					anchor.L2Height,
					ShortHash(anchor.TxID, 10),
				),
			})
			step++
		}

		if blockIndex%cfg.MineBtcEveryL2Blocks == 0 {
			mined := btc.MineBlock()
			events = append(events, DemoEvent{
				Step:    step,
				Message: fmt.Sprintf("Bitcoin block %d mined (%d txs)", mined.Height, len(mined.TxIDs)),
			})
			step++
		}
	}

	for i := int64(0); i < cfg.RequiredPoWConfirms+1; i++ {
		mined := btc.MineBlock()
		events = append(events, DemoEvent{
			Step:    step,
			Message: fmt.Sprintf("Bitcoin block %d mined (confirmation extension)", mined.Height),
		})
		step++
	}

	return l2, btc, bridge, events, nil
}

func BuildSecurityReport(l2 *TendermintL2Chain, btc *BitcoinPoWChain, bridge *CheckpointBridge) []SecurityRow {
	blocks := l2.Blocks()
	if len(blocks) <= 1 {
		return nil
	}

	report := make([]SecurityRow, 0, len(blocks)-1)
	for _, block := range blocks[1:] {
		_, anchored := bridge.AnchorTxIDByHeight[block.Height]
		report = append(report, SecurityRow{
			L2Height:         block.Height,
			L2Hash:           ShortHash(block.BlockHash, 10),
			PoSFinal:         block.HasSupermajority(),
			AnchoredToPoW:    anchored,
			PoWConfirmations: bridge.PoWConfirmationsForHeight(block.Height, btc),
			PoWFinal:         bridge.IsPoWFinal(block.Height, btc),
		})
	}
	return report
}
