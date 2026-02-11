package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"powpos/x/checkpoint/types"
)

func (q queryServer) QueryLatestCheckpoint(
	ctx context.Context,
	req *types.QueryLatestCheckpointRequest,
) (*types.QueryLatestCheckpointResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	checkpoint, err := q.k.GetLatestCheckpoint(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "latest checkpoint not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryLatestCheckpointResponse{
		Checkpoint: checkpoint,
	}, nil
}

func (q queryServer) QueryCheckpointByHeight(
	ctx context.Context,
	req *types.QueryCheckpointByHeightRequest,
) (*types.QueryCheckpointByHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	checkpoint, err := q.k.GetCheckpointByHeight(ctx, req.Height)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "checkpoint not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryCheckpointByHeightResponse{
		Checkpoint: checkpoint,
	}, nil
}
