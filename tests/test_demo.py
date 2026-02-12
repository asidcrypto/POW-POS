from __future__ import annotations

import unittest

from pow_pos_demo.core import CheckpointBridge, ConsensusFailure, TendermintL2Chain, Validator
from pow_pos_demo.simulation import build_security_report, run_demo


class TendermintConsensusTests(unittest.TestCase):
    def test_block_commits_with_supermajority_stake(self) -> None:
        chain = TendermintL2Chain(
            validators=[
                Validator("alice", 40),
                Validator("bob", 35),
                Validator("carol", 25),
            ]
        )
        chain.queue_transaction("tx-1")
        block = chain.propose_block(signing_validators={"alice", "bob"})
        self.assertTrue(block.has_supermajority)
        self.assertEqual(block.height, 1)

    def test_block_rejected_without_supermajority(self) -> None:
        chain = TendermintL2Chain(
            validators=[
                Validator("alice", 40),
                Validator("bob", 35),
                Validator("carol", 25),
            ]
        )
        chain.queue_transaction("tx-1")
        with self.assertRaises(ConsensusFailure):
            chain.propose_block(signing_validators={"alice", "carol"})


class AnchorFlowTests(unittest.TestCase):
    def test_checkpoint_finality_after_confirmations(self) -> None:
        l2, btc, bridge, _events = run_demo(
            l2_blocks=6,
            checkpoint_interval=3,
            required_pow_confirmations=2,
            mine_btc_every_l2_blocks=1,
        )
        self.assertIn(3, bridge.anchor_txid_by_height)
        self.assertIn(6, bridge.anchor_txid_by_height)
        self.assertGreaterEqual(bridge.pow_confirmations_for_height(3, btc), 2)
        self.assertGreaterEqual(bridge.pow_confirmations_for_height(6, btc), 2)
        self.assertTrue(bridge.is_pow_final(3, btc))
        self.assertTrue(bridge.is_pow_final(6, btc))
        self.assertFalse(bridge.is_pow_final(5, btc))
        self.assertEqual(bridge.pow_confirmations_for_height(5, btc), 0)

        report = build_security_report(l2, btc, bridge)
        row_6 = next(row for row in report if row["l2_height"] == 6)
        self.assertTrue(row_6["anchored_to_pow"])
        self.assertTrue(row_6["pow_final"])

    def test_bridge_validates_inputs(self) -> None:
        with self.assertRaises(ValueError):
            CheckpointBridge(checkpoint_interval=0, required_confirmations=2)
        with self.assertRaises(ValueError):
            CheckpointBridge(checkpoint_interval=3, required_confirmations=0)


if __name__ == "__main__":
    unittest.main()
