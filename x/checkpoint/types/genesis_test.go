package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"powpos/x/checkpoint/types"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Checkpoints: []types.Checkpoint{
					{
						Height:    20,
						AppHash:   []byte{0x01, 0x02},
						Timestamp: 1700000000,
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicate checkpoint height is invalid",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Checkpoints: []types.Checkpoint{
					{Height: 20},
					{Height: 20},
				},
			},
			valid: false,
		},
		{
			desc: "height zero checkpoint is invalid",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Checkpoints: []types.Checkpoint{
					{Height: 0},
				},
			},
			valid: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
