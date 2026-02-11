package types

var bitcoinCommitmentMagic = [4]byte{0x43, 0x48, 0x4b, 0x50} // "CHKP"

const maxBitcoinOPReturnPayloadSize = 79

// FormatBitcoinCommitment creates a compact payload suitable for Bitcoin OP_RETURN.
// It prepends a fixed 4-byte magic prefix and truncates appHash if needed so the
// total payload stays below 80 bytes.
func FormatBitcoinCommitment(appHash []byte) []byte {
	maxAppHashLen := maxBitcoinOPReturnPayloadSize - len(bitcoinCommitmentMagic)
	if maxAppHashLen < 0 {
		return nil
	}

	if len(appHash) > maxAppHashLen {
		appHash = appHash[:maxAppHashLen]
	}

	payload := make([]byte, 0, len(bitcoinCommitmentMagic)+len(appHash))
	payload = append(payload, bitcoinCommitmentMagic[:]...)
	payload = append(payload, appHash...)
	return payload
}
