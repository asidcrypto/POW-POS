package keeper_test

import (
	"testing"

	"powpos/x/checkpoint/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		Checkpoints: []types.Checkpoint{
			{
				Height:    20,
				AppHash:   []byte{0x01, 0x02, 0x03},
				Timestamp: 1700000000,
			},
		},
	}

	f := initFixture(t)
	err := f.keeper.InitGenesis(f.ctx, genesisState)
	require.NoError(t, err)
	got, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
	require.EqualExportedValues(t, genesisState.Checkpoints, got.Checkpoints)
}
