package types_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"powpos/x/checkpoint/types"
)

func TestFormatBitcoinCommitment(t *testing.T) {
	appHash := bytes.Repeat([]byte{0xaa}, 100)

	commitment := types.FormatBitcoinCommitment(appHash)

	require.Less(t, len(commitment), 80)
	require.True(t, bytes.HasPrefix(commitment, []byte{0x43, 0x48, 0x4b, 0x50}))
	require.Equal(t, appHash[:75], commitment[4:])
}
