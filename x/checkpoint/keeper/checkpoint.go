package keeper

import (
	"context"
	"encoding/hex"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"powpos/x/checkpoint/types"
)

// MaybeCreateCheckpoint persists a checkpoint every configured interval.
func (k Keeper) MaybeCreateCheckpoint(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := uint64(sdkCtx.BlockHeight())

	if height == 0 || height%types.CheckpointInterval != 0 {
		return nil
	}

	header := sdkCtx.BlockHeader()
	checkpoint := types.Checkpoint{
		Height:    height,
		AppHash:   append([]byte(nil), header.AppHash...),
		Timestamp: timestamppb.New(sdkCtx.BlockTime().UTC()),
	}

	if err := k.Checkpoints.Set(ctx, height, checkpoint); err != nil {
		return err
	}

	if err := k.LatestCheckpointHeight.Set(ctx, height); err != nil {
		return err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCheckpointCreated,
			sdk.NewAttribute(types.AttributeKeyHeight, strconv.FormatUint(height, 10)),
			sdk.NewAttribute(types.AttributeKeyAppHash, hex.EncodeToString(checkpoint.AppHash)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, checkpoint.Timestamp.AsTime().UTC().Format(time.RFC3339Nano)),
		),
	)

	return nil
}

func (k Keeper) GetLatestCheckpoint(ctx context.Context) (types.Checkpoint, error) {
	height, err := k.LatestCheckpointHeight.Get(ctx)
	if err != nil {
		return types.Checkpoint{}, err
	}

	return k.Checkpoints.Get(ctx, height)
}

func (k Keeper) GetCheckpointByHeight(ctx context.Context, height uint64) (types.Checkpoint, error) {
	return k.Checkpoints.Get(ctx, height)
}

func (k Keeper) setGenesisCheckpoints(ctx context.Context, checkpoints []types.Checkpoint) error {
	var latestHeight uint64

	for _, checkpoint := range checkpoints {
		if err := k.Checkpoints.Set(ctx, checkpoint.Height, checkpoint); err != nil {
			return err
		}

		if checkpoint.Height > latestHeight {
			latestHeight = checkpoint.Height
		}
	}

	if latestHeight > 0 {
		return k.LatestCheckpointHeight.Set(ctx, latestHeight)
	}

	return nil
}

func (k Keeper) exportCheckpoints(ctx context.Context) ([]types.Checkpoint, error) {
	checkpoints := make([]types.Checkpoint, 0)

	err := k.Checkpoints.Walk(ctx, nil, func(_ uint64, value types.Checkpoint) (bool, error) {
		checkpoints = append(checkpoints, value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return checkpoints, nil
}
