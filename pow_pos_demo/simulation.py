from __future__ import annotations

from dataclasses import dataclass

from .core import BitcoinPoWChain, CheckpointBridge, TendermintL2Chain, Validator, short_hash


@dataclass(frozen=True)
class DemoEvent:
    step: int
    message: str


def default_validators() -> list[Validator]:
    return [
        Validator(name="alice", stake=40),
        Validator(name="bob", stake=30),
        Validator(name="carol", stake=20),
        Validator(name="dave", stake=10),
    ]


def run_demo(
    l2_blocks: int = 12,
    checkpoint_interval: int = 3,
    required_pow_confirmations: int = 2,
    mine_btc_every_l2_blocks: int = 2,
) -> tuple[TendermintL2Chain, BitcoinPoWChain, CheckpointBridge, list[DemoEvent]]:
    if l2_blocks <= 0:
        raise ValueError("l2_blocks must be positive")
    if mine_btc_every_l2_blocks <= 0:
        raise ValueError("mine_btc_every_l2_blocks must be positive")

    l2 = TendermintL2Chain(default_validators())
    btc = BitcoinPoWChain()
    bridge = CheckpointBridge(
        checkpoint_interval=checkpoint_interval,
        required_confirmations=required_pow_confirmations,
    )

    events: list[DemoEvent] = []
    step = 1

    for block_index in range(1, l2_blocks + 1):
        l2.queue_transaction(f"user_tx_{block_index:03d}_a")
        l2.queue_transaction(f"user_tx_{block_index:03d}_b")
        block = l2.propose_block()
        events.append(
            DemoEvent(
                step=step,
                message=(
                    f"L2 block {block.height} committed "
                    f"({block.signed_stake}/{block.total_stake} stake, hash={short_hash(block.block_hash)})"
                ),
            )
        )
        step += 1

        anchor = bridge.maybe_anchor_latest_block(l2, btc)
        if anchor is not None:
            events.append(
                DemoEvent(
                    step=step,
                    message=(
                        f"Checkpoint submitted for L2 block {anchor.l2_height} "
                        f"(txid={short_hash(anchor.txid)})"
                    ),
                )
            )
            step += 1

        if block_index % mine_btc_every_l2_blocks == 0:
            mined = btc.mine_block()
            tx_count = len(mined.txids)
            events.append(
                DemoEvent(
                    step=step,
                    message=f"Bitcoin block {mined.height} mined ({tx_count} txs)",
                )
            )
            step += 1

    # Mine extra Bitcoin blocks so anchors can pass confirmation threshold.
    for _ in range(required_pow_confirmations + 1):
        mined = btc.mine_block()
        events.append(
            DemoEvent(
                step=step,
                message=f"Bitcoin block {mined.height} mined (confirmation extension)",
            )
        )
        step += 1

    return l2, btc, bridge, events


def build_security_report(
    l2: TendermintL2Chain,
    btc: BitcoinPoWChain,
    bridge: CheckpointBridge,
) -> list[dict[str, str | int | bool]]:
    report: list[dict[str, str | int | bool]] = []
    for block in l2.blocks[1:]:
        confirmations = bridge.pow_confirmations_for_height(block.height, btc)
        anchored = block.height in bridge.anchor_txid_by_height
        report.append(
            {
                "l2_height": block.height,
                "l2_hash": short_hash(block.block_hash),
                "pos_final": block.has_supermajority,
                "anchored_to_pow": anchored,
                "pow_confirmations": confirmations,
                "pow_final": bridge.is_pow_final(block.height, btc),
            }
        )
    return report
