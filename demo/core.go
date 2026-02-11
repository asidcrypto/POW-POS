package demo

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

func digest(parts ...string) string {
	material := strings.Join(parts, "|")
	sum := sha256.Sum256([]byte(material))
	return hex.EncodeToString(sum[:])
}

type Validator struct {
	Name  string
	Stake int64
}

type L2Block struct {
	Height       int64
	PreviousHash string
	Txs          []string
	Proposer     string
	SignedBy     []string
	SignedStake  int64
	TotalStake   int64
	BlockHash    string
}

func (b L2Block) HasSupermajority() bool {
	return b.SignedStake*3 > b.TotalStake*2
}

type AnchorTransaction struct {
	TxID        string
	L2Height    int64
	L2BlockHash string
}

type BitcoinBlock struct {
	Height       int64
	PreviousHash string
	TxIDs        []string
	BlockHash    string
}

var ErrConsensusFailure = errors.New("block rejected: signed stake is not greater than 2/3")

type TendermintL2Chain struct {
	chainID        string
	validators     []Validator
	validatorStake map[string]int64
	totalStake     int64
	pendingTxs     []string
	blocks         []L2Block
}

func NewTendermintL2Chain(chainID string, validators []Validator) (*TendermintL2Chain, error) {
	if chainID == "" {
		chainID = "pow-pos-demo"
	}
	if len(validators) == 0 {
		return nil, errors.New("at least one validator is required")
	}

	validatorStake := make(map[string]int64, len(validators))
	var totalStake int64
	for _, v := range validators {
		if v.Name == "" {
			return nil, errors.New("validator name cannot be empty")
		}
		if v.Stake <= 0 {
			return nil, errors.New("validator stake must be positive")
		}
		if _, exists := validatorStake[v.Name]; exists {
			return nil, fmt.Errorf("duplicate validator name: %s", v.Name)
		}
		validatorStake[v.Name] = v.Stake
		totalStake += v.Stake
	}

	genesisHash := digest(chainID, "genesis")
	genesis := L2Block{
		Height:       0,
		PreviousHash: strings.Repeat("0", 64),
		Txs:          nil,
		Proposer:     "genesis",
		SignedBy:     validatorNames(validators),
		SignedStake:  totalStake,
		TotalStake:   totalStake,
		BlockHash:    genesisHash,
	}

	return &TendermintL2Chain{
		chainID:        chainID,
		validators:     validators,
		validatorStake: validatorStake,
		totalStake:     totalStake,
		pendingTxs:     []string{},
		blocks:         []L2Block{genesis},
	}, nil
}

func validatorNames(validators []Validator) []string {
	names := make([]string, 0, len(validators))
	for _, v := range validators {
		names = append(names, v.Name)
	}
	return names
}

func (c *TendermintL2Chain) LatestBlock() L2Block {
	return c.blocks[len(c.blocks)-1]
}

func (c *TendermintL2Chain) Blocks() []L2Block {
	blocks := make([]L2Block, len(c.blocks))
	copy(blocks, c.blocks)
	return blocks
}

func (c *TendermintL2Chain) TotalStake() int64 {
	return c.totalStake
}

func (c *TendermintL2Chain) QueueTransaction(tx string) {
	c.pendingTxs = append(c.pendingTxs, tx)
}

func (c *TendermintL2Chain) ProposeBlock(signingValidators []string, maxTxs int) (L2Block, error) {
	if maxTxs <= 0 {
		return L2Block{}, errors.New("maxTxs must be positive")
	}

	latest := c.LatestBlock()
	height := latest.Height + 1
	proposer := c.validators[(height-1)%int64(len(c.validators))].Name

	take := maxTxs
	if take > len(c.pendingTxs) {
		take = len(c.pendingTxs)
	}
	txs := make([]string, take)
	copy(txs, c.pendingTxs[:take])
	c.pendingTxs = c.pendingTxs[take:]

	signers, signedStake, err := c.resolveSigners(signingValidators)
	if err != nil {
		return L2Block{}, err
	}

	if signedStake*3 <= c.totalStake*2 {
		return L2Block{}, fmt.Errorf("%w (%d/%d)", ErrConsensusFailure, signedStake, c.totalStake)
	}

	blockHash := digest(
		c.chainID,
		fmt.Sprintf("%d", height),
		latest.BlockHash,
		proposer,
		strings.Join(txs, ","),
		strings.Join(signers, ","),
	)

	block := L2Block{
		Height:       height,
		PreviousHash: latest.BlockHash,
		Txs:          txs,
		Proposer:     proposer,
		SignedBy:     signers,
		SignedStake:  signedStake,
		TotalStake:   c.totalStake,
		BlockHash:    blockHash,
	}
	c.blocks = append(c.blocks, block)
	return block, nil
}

func (c *TendermintL2Chain) resolveSigners(signingValidators []string) ([]string, int64, error) {
	if len(signingValidators) == 0 {
		signers := validatorNames(c.validators)
		return signers, c.totalStake, nil
	}

	included := make(map[string]bool, len(signingValidators))
	for _, name := range signingValidators {
		if _, ok := c.validatorStake[name]; !ok {
			return nil, 0, fmt.Errorf("unknown validator in signing set: %s", name)
		}
		included[name] = true
	}

	signers := make([]string, 0, len(c.validators))
	var signedStake int64
	for _, v := range c.validators {
		if included[v.Name] {
			signers = append(signers, v.Name)
			signedStake += v.Stake
		}
	}
	return signers, signedStake, nil
}

type BitcoinPoWChain struct {
	chainID         string
	mempool         []AnchorTransaction
	anchorByTxID    map[string]AnchorTransaction
	blocks          []BitcoinBlock
	nextAnchorNonce int64
}

func NewBitcoinPoWChain(chainID string) *BitcoinPoWChain {
	if chainID == "" {
		chainID = "bitcoin-regtest-demo"
	}
	genesisHash := digest(chainID, "genesis")
	genesis := BitcoinBlock{
		Height:       0,
		PreviousHash: strings.Repeat("0", 64),
		TxIDs:        nil,
		BlockHash:    genesisHash,
	}
	return &BitcoinPoWChain{
		chainID:         chainID,
		mempool:         []AnchorTransaction{},
		anchorByTxID:    map[string]AnchorTransaction{},
		blocks:          []BitcoinBlock{genesis},
		nextAnchorNonce: 1,
	}
}

func (c *BitcoinPoWChain) TipHeight() int64 {
	return c.blocks[len(c.blocks)-1].Height
}

func (c *BitcoinPoWChain) Blocks() []BitcoinBlock {
	blocks := make([]BitcoinBlock, len(c.blocks))
	copy(blocks, c.blocks)
	return blocks
}

func (c *BitcoinPoWChain) BroadcastAnchor(l2Height int64, l2BlockHash string) AnchorTransaction {
	txID := digest(
		c.chainID,
		"anchor",
		fmt.Sprintf("%d", c.nextAnchorNonce),
		fmt.Sprintf("%d", l2Height),
		l2BlockHash,
	)
	c.nextAnchorNonce++

	anchor := AnchorTransaction{
		TxID:        txID,
		L2Height:    l2Height,
		L2BlockHash: l2BlockHash,
	}
	c.mempool = append(c.mempool, anchor)
	c.anchorByTxID[txID] = anchor
	return anchor
}

func (c *BitcoinPoWChain) MineBlock() BitcoinBlock {
	txIDs := make([]string, 0, len(c.mempool))
	for _, tx := range c.mempool {
		txIDs = append(txIDs, tx.TxID)
	}

	previous := c.blocks[len(c.blocks)-1]
	height := previous.Height + 1
	blockHash := digest(c.chainID, fmt.Sprintf("%d", height), previous.BlockHash, strings.Join(txIDs, ","))
	block := BitcoinBlock{
		Height:       height,
		PreviousHash: previous.BlockHash,
		TxIDs:        txIDs,
		BlockHash:    blockHash,
	}
	c.blocks = append(c.blocks, block)
	c.mempool = c.mempool[:0]
	return block
}

func (c *BitcoinPoWChain) Confirmations(txID string) int64 {
	for _, b := range c.blocks {
		for _, id := range b.TxIDs {
			if id == txID {
				return c.TipHeight() - b.Height + 1
			}
		}
	}
	if _, exists := c.anchorByTxID[txID]; exists {
		return 0
	}
	return 0
}

type CheckpointBridge struct {
	CheckpointInterval int64
	RequiredConfirms   int64
	AnchorTxIDByHeight map[int64]string
}

func NewCheckpointBridge(checkpointInterval, requiredConfirms int64) (*CheckpointBridge, error) {
	if checkpointInterval <= 0 {
		return nil, errors.New("checkpointInterval must be positive")
	}
	if requiredConfirms <= 0 {
		return nil, errors.New("requiredConfirms must be positive")
	}
	return &CheckpointBridge{
		CheckpointInterval: checkpointInterval,
		RequiredConfirms:   requiredConfirms,
		AnchorTxIDByHeight: map[int64]string{},
	}, nil
}

func (b *CheckpointBridge) MaybeAnchorLatestBlock(l2 *TendermintL2Chain, btc *BitcoinPoWChain) *AnchorTransaction {
	latest := l2.LatestBlock()
	if latest.Height == 0 {
		return nil
	}
	if latest.Height%b.CheckpointInterval != 0 {
		return nil
	}
	if _, exists := b.AnchorTxIDByHeight[latest.Height]; exists {
		return nil
	}

	anchor := btc.BroadcastAnchor(latest.Height, latest.BlockHash)
	b.AnchorTxIDByHeight[latest.Height] = anchor.TxID
	return &anchor
}

func (b *CheckpointBridge) PoWConfirmationsForHeight(l2Height int64, btc *BitcoinPoWChain) int64 {
	txID, ok := b.AnchorTxIDByHeight[l2Height]
	if !ok {
		return 0
	}
	return btc.Confirmations(txID)
}

func (b *CheckpointBridge) IsPoWFinal(l2Height int64, btc *BitcoinPoWChain) bool {
	return b.PoWConfirmationsForHeight(l2Height, btc) >= b.RequiredConfirms
}

func ShortHash(value string, width int) string {
	if width <= 0 {
		width = 10
	}
	if len(value) < width {
		return value
	}
	return value[:width]
}
