from __future__ import annotations

from dataclasses import dataclass, field
from hashlib import sha256
from typing import Iterable


def _digest(parts: Iterable[str]) -> str:
    material = "|".join(parts).encode("utf-8")
    return sha256(material).hexdigest()


class ConsensusFailure(RuntimeError):
    """Raised when a block cannot reach Tendermint-style supermajority."""


@dataclass(frozen=True)
class Validator:
    name: str
    stake: int


@dataclass(frozen=True)
class L2Block:
    height: int
    previous_hash: str
    transactions: tuple[str, ...]
    proposer: str
    signed_by: tuple[str, ...]
    signed_stake: int
    total_stake: int
    block_hash: str

    @property
    def has_supermajority(self) -> bool:
        return self.signed_stake * 3 > self.total_stake * 2


@dataclass(frozen=True)
class AnchorTransaction:
    txid: str
    l2_height: int
    l2_block_hash: str


@dataclass(frozen=True)
class BitcoinBlock:
    height: int
    previous_hash: str
    txids: tuple[str, ...]
    block_hash: str


class TendermintL2Chain:
    """
    Minimal Tendermint-style PoS chain model.

    Blocks commit only when signed stake exceeds 2/3 of total stake.
    """

    def __init__(self, validators: list[Validator], chain_id: str = "pow-pos-demo") -> None:
        if not validators:
            raise ValueError("at least one validator is required")
        if any(v.stake <= 0 for v in validators):
            raise ValueError("validator stake must be positive")

        self.chain_id = chain_id
        self.validators = validators
        self._validator_stake = {v.name: v.stake for v in validators}
        self.total_stake = sum(v.stake for v in validators)
        self.pending_transactions: list[str] = []

        genesis_hash = _digest([chain_id, "genesis"])
        self.blocks: list[L2Block] = [
            L2Block(
                height=0,
                previous_hash="0" * 64,
                transactions=(),
                proposer="genesis",
                signed_by=tuple(v.name for v in validators),
                signed_stake=self.total_stake,
                total_stake=self.total_stake,
                block_hash=genesis_hash,
            )
        ]

    @property
    def latest_block(self) -> L2Block:
        return self.blocks[-1]

    def queue_transaction(self, tx: str) -> None:
        self.pending_transactions.append(tx)

    def propose_block(
        self,
        signing_validators: set[str] | None = None,
        max_txs: int = 100,
    ) -> L2Block:
        if max_txs <= 0:
            raise ValueError("max_txs must be positive")

        height = self.latest_block.height + 1
        proposer = self.validators[(height - 1) % len(self.validators)].name
        txs = tuple(self.pending_transactions[:max_txs])
        del self.pending_transactions[:max_txs]

        if signing_validators is None:
            signers = tuple(v.name for v in self.validators)
        else:
            unknown = signing_validators.difference(self._validator_stake)
            if unknown:
                raise ValueError(f"unknown validators in signing set: {sorted(unknown)}")
            signers = tuple(v.name for v in self.validators if v.name in signing_validators)

        signed_stake = sum(self._validator_stake[name] for name in signers)
        if signed_stake * 3 <= self.total_stake * 2:
            raise ConsensusFailure(
                f"block {height} rejected: signed stake {signed_stake}/{self.total_stake} "
                "is not greater than 2/3"
            )

        block_hash = _digest(
            [
                self.chain_id,
                str(height),
                self.latest_block.block_hash,
                proposer,
                ",".join(txs),
                ",".join(signers),
            ]
        )
        block = L2Block(
            height=height,
            previous_hash=self.latest_block.block_hash,
            transactions=txs,
            proposer=proposer,
            signed_by=signers,
            signed_stake=signed_stake,
            total_stake=self.total_stake,
            block_hash=block_hash,
        )
        self.blocks.append(block)
        return block


class BitcoinPoWChain:
    """Small Bitcoin-like chain model with mempool and confirmations."""

    def __init__(self, chain_id: str = "bitcoin-regtest-demo") -> None:
        self.chain_id = chain_id
        self.mempool: list[AnchorTransaction] = []
        self._anchor_by_txid: dict[str, AnchorTransaction] = {}

        genesis_hash = _digest([chain_id, "genesis"])
        self.blocks: list[BitcoinBlock] = [
            BitcoinBlock(height=0, previous_hash="0" * 64, txids=(), block_hash=genesis_hash)
        ]
        self._next_anchor_nonce = 1

    @property
    def tip_height(self) -> int:
        return self.blocks[-1].height

    def broadcast_anchor(self, l2_height: int, l2_block_hash: str) -> AnchorTransaction:
        txid = _digest(
            [
                self.chain_id,
                "anchor",
                str(self._next_anchor_nonce),
                str(l2_height),
                l2_block_hash,
            ]
        )
        self._next_anchor_nonce += 1
        anchor = AnchorTransaction(txid=txid, l2_height=l2_height, l2_block_hash=l2_block_hash)
        self.mempool.append(anchor)
        self._anchor_by_txid[txid] = anchor
        return anchor

    def mine_block(self) -> BitcoinBlock:
        txids = tuple(tx.txid for tx in self.mempool)
        previous = self.blocks[-1]
        height = previous.height + 1
        block_hash = _digest([self.chain_id, str(height), previous.block_hash, ",".join(txids)])
        block = BitcoinBlock(height=height, previous_hash=previous.block_hash, txids=txids, block_hash=block_hash)
        self.blocks.append(block)
        self.mempool.clear()
        return block

    def confirmations(self, txid: str) -> int:
        for block in self.blocks:
            if txid in block.txids:
                return self.tip_height - block.height + 1
        if txid in self._anchor_by_txid:
            return 0
        return 0


@dataclass
class CheckpointBridge:
    checkpoint_interval: int = 3
    required_confirmations: int = 2
    anchor_txid_by_height: dict[int, str] = field(default_factory=dict)

    def __post_init__(self) -> None:
        if self.checkpoint_interval <= 0:
            raise ValueError("checkpoint_interval must be positive")
        if self.required_confirmations <= 0:
            raise ValueError("required_confirmations must be positive")

    def maybe_anchor_latest_block(
        self,
        l2_chain: TendermintL2Chain,
        bitcoin_chain: BitcoinPoWChain,
    ) -> AnchorTransaction | None:
        latest = l2_chain.latest_block
        if latest.height == 0:
            return None
        if latest.height % self.checkpoint_interval != 0:
            return None
        if latest.height in self.anchor_txid_by_height:
            return None

        anchor = bitcoin_chain.broadcast_anchor(l2_height=latest.height, l2_block_hash=latest.block_hash)
        self.anchor_txid_by_height[latest.height] = anchor.txid
        return anchor

    def pow_confirmations_for_height(self, l2_height: int, bitcoin_chain: BitcoinPoWChain) -> int:
        txid = self.anchor_txid_by_height.get(l2_height)
        if txid is None:
            return 0
        return bitcoin_chain.confirmations(txid)

    def is_pow_final(self, l2_height: int, bitcoin_chain: BitcoinPoWChain) -> bool:
        return self.pow_confirmations_for_height(l2_height, bitcoin_chain) >= self.required_confirmations


def short_hash(value: str, width: int = 10) -> str:
    if width <= 0:
        raise ValueError("width must be positive")
    return value[:width]
