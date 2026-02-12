#!/usr/bin/env python3
from __future__ import annotations

import argparse
import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parents[1]
if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))

from pow_pos_demo.simulation import build_security_report, run_demo


def yes_no(value: bool) -> str:
    return "yes" if value else "no"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Tendermint-style PoS L2 + Bitcoin-style PoW anchoring demo. "
            "This is a deterministic simulation for architecture demonstration."
        )
    )
    parser.add_argument("--l2-blocks", type=int, default=12, help="Number of L2 blocks to produce")
    parser.add_argument(
        "--checkpoint-interval",
        type=int,
        default=3,
        help="Anchor every N L2 blocks into Bitcoin",
    )
    parser.add_argument(
        "--pow-confirmations",
        type=int,
        default=2,
        help="Bitcoin confirmations required before a checkpoint is considered PoW-final",
    )
    parser.add_argument(
        "--mine-btc-every",
        type=int,
        default=2,
        help="Mine one Bitcoin block every N L2 blocks during the main loop",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    l2, btc, bridge, events = run_demo(
        l2_blocks=args.l2_blocks,
        checkpoint_interval=args.checkpoint_interval,
        required_pow_confirmations=args.pow_confirmations,
        mine_btc_every_l2_blocks=args.mine_btc_every,
    )

    print("\n=== PoS-on-PoW Demo (Tendermint-style + Bitcoin-style) ===\n")
    for event in events:
        print(f"[{event.step:02d}] {event.message}")

    report = build_security_report(l2, btc, bridge)
    print("\nSecurity report:")
    print(
        "  "
        + "height".ljust(8)
        + "hash".ljust(14)
        + "pos_final".ljust(12)
        + "anchored".ljust(11)
        + "pow_conf".ljust(10)
        + "pow_final"
    )
    print("  " + "-" * 67)
    for row in report:
        print(
            "  "
            + str(row["l2_height"]).ljust(8)
            + str(row["l2_hash"]).ljust(14)
            + yes_no(bool(row["pos_final"])).ljust(12)
            + yes_no(bool(row["anchored_to_pow"])).ljust(11)
            + str(row["pow_confirmations"]).ljust(10)
            + yes_no(bool(row["pow_final"]))
        )

    pow_final_count = sum(1 for row in report if row["pow_final"])
    print(
        f"\nSummary: {len(report)} L2 blocks committed by PoS, "
        f"{pow_final_count} blocks additionally secured by PoW checkpoints."
    )
    print(
        "Note: This demo models the architecture; production systems need real CometBFT/Cosmos "
        "consensus, bridge proofs, and adversarial testing."
    )


if __name__ == "__main__":
    main()
