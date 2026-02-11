package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"powpos/x/checkpoint/types"
)

func TestMaybeCreateCheckpoint(t *testing.T) {
	f := initFixture(t)

	baseCtx := sdk.UnwrapSDKContext(f.ctx)
	header := baseCtx.BlockHeader()
	header.AppHash = []byte{0xde, 0xad, 0xbe, 0xef}

	ctxBefore := baseCtx.
		WithBlockHeader(header).
		WithBlockHeight(19).
		WithBlockTime(time.Unix(1700000000, 0).UTC()).
		WithEventManager(sdk.NewEventManager())

	require.NoError(t, f.keeper.MaybeCreateCheckpoint(ctxBefore))
	_, err := f.keeper.GetCheckpointByHeight(ctxBefore, 19)
	require.Error(t, err)

	ctxTarget := ctxBefore.
		WithBlockHeight(20).
		WithBlockTime(time.Unix(1700000020, 0).UTC()).
		WithEventManager(sdk.NewEventManager())

	require.NoError(t, f.keeper.MaybeCreateCheckpoint(ctxTarget))

	got, err := f.keeper.GetCheckpointByHeight(ctxTarget, 20)
	require.NoError(t, err)
	require.Equal(t, uint64(20), got.Height)
	require.Equal(t, header.AppHash, got.AppHash)
	require.Equal(t, int64(1700000020), got.Timestamp)

	latest, err := f.keeper.GetLatestCheckpoint(ctxTarget)
	require.NoError(t, err)
	require.Equal(t, got, latest)

	require.True(t, hasCheckpointCreatedEvent(sdk.UnwrapSDKContext(ctxTarget).EventManager().Events()))
}

func hasCheckpointCreatedEvent(events sdk.Events) bool {
	for _, event := range events {
		if event.Type == types.EventTypeCheckpointCreated {
			return true
		}
	}

	return false
}

func TestFormatBitcoinCommitmentWithStoredCheckpoint(t *testing.T) {
	f := initFixture(t)
	ctx := sdk.UnwrapSDKContext(f.ctx)
	header := ctx.BlockHeader()
	header.AppHash = []byte{0x01, 0x02, 0x03, 0x04}
	ctx = ctx.WithBlockHeader(header).WithBlockHeight(20).WithBlockTime(time.Unix(1700000020, 0).UTC())

	require.NoError(t, f.keeper.MaybeCreateCheckpoint(ctx))

	checkpoint, err := f.keeper.GetLatestCheckpoint(ctx)
	require.NoError(t, err)

	payload := types.FormatBitcoinCommitment(checkpoint.AppHash)
	require.Less(t, len(payload), 80)
}
