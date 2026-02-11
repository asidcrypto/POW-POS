package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		Checkpoints: []Checkpoint{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenHeights := make(map[uint64]struct{}, len(gs.Checkpoints))
	for _, checkpoint := range gs.Checkpoints {
		if checkpoint.Height == 0 {
			return fmt.Errorf("checkpoint height must be > 0")
		}

		if _, exists := seenHeights[checkpoint.Height]; exists {
			return fmt.Errorf("duplicate checkpoint height %d", checkpoint.Height)
		}
		seenHeights[checkpoint.Height] = struct{}{}
	}

	return nil
}
