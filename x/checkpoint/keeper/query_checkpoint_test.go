package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"powpos/x/checkpoint/keeper"
	"powpos/x/checkpoint/types"
)

func TestQueryLatestCheckpoint(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)

	checkpoint := types.Checkpoint{
		Height:    40,
		AppHash:   []byte{0xab, 0xcd},
		Timestamp: 1700000040,
	}
	require.NoError(t, f.keeper.Checkpoints.Set(f.ctx, checkpoint.Height, checkpoint))
	require.NoError(t, f.keeper.LatestCheckpointHeight.Set(f.ctx, checkpoint.Height))

	resp, err := qs.QueryLatestCheckpoint(f.ctx, &types.QueryLatestCheckpointRequest{})
	require.NoError(t, err)
	require.Equal(t, checkpoint, resp.Checkpoint)
}

func TestQueryCheckpointByHeight(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)

	checkpoint := types.Checkpoint{
		Height:    60,
		AppHash:   []byte{0x11, 0x22, 0x33},
		Timestamp: 1700000060,
	}
	require.NoError(t, f.keeper.Checkpoints.Set(f.ctx, checkpoint.Height, checkpoint))

	resp, err := qs.QueryCheckpointByHeight(f.ctx, &types.QueryCheckpointByHeightRequest{Height: checkpoint.Height})
	require.NoError(t, err)
	require.Equal(t, checkpoint, resp.Checkpoint)

	_, err = qs.QueryCheckpointByHeight(f.ctx, &types.QueryCheckpointByHeightRequest{Height: 61})
	require.Error(t, err)
	require.Equal(t, codes.NotFound, status.Code(err))
}
