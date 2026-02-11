# PoS L2 Anchored to PoW (Tendermint-style Demo)

This repository is a **demo** of a Layer 2 architecture where:

- L2 blocks are produced with **Tendermint-style Proof of Stake finality** (2/3+ stake signatures).
- The L2 periodically writes **checkpoints** to a Bitcoin-like **Proof of Work** chain.
- A block can be considered:
  - `PoS-final` right after supermajority commit.
  - `PoW-final` after its checkpoint receives enough Bitcoin confirmations.

This is designed to show *how the flow can work* for presentations and prototyping.

## What this demo is (and is not)

### It demonstrates
- Fast L2 finality with validator stake voting.
- Periodic checkpoint anchoring into a PoW chain.
- Confirmation-based security escalation from PoS-final to PoW-final.

### It does not claim
- Full trustless inheritance of Bitcoin security by itself.
- Production bridge security (fraud proofs, light client proofs, reorg handling, etc.).
- A full Cosmos SDK chain implementation.

---

## Quickstart (macOS)

### 1) Prerequisites
- macOS with `python3` available.
  - Verify: `python3 --version`

### 2) Clone and enter repo
```bash
git clone <your-repo-url>
cd POW-POS
```

### 3) Run tests
```bash
python3 -m unittest discover -s tests -v
```

### 4) Run the demo
```bash
python3 scripts/run_demo.py
```

Optional flags:
```bash
python3 scripts/run_demo.py \
  --l2-blocks 15 \
  --checkpoint-interval 3 \
  --pow-confirmations 2 \
  --mine-btc-every 2
```

You will see:
- L2 blocks committing with stake signatures.
- Checkpoints submitted at interval heights.
- Bitcoin blocks mined and confirmations increasing.
- A final table showing which L2 blocks are PoS-final and PoW-final.

---

## How it maps to your real goal

For a production-grade version on top of Bitcoin:

1. Replace the L2 simulator with a real **CometBFT (Tendermint)** app chain (usually Cosmos SDK + custom modules).
2. Run Bitcoin `regtest`/`testnet` and submit checkpoint transactions from a relayer.
3. Add verifier logic:
   - checkpoint inclusion proofs,
   - required confirmation policy,
   - reorg and liveness handling.
4. Add slashing/dispute mechanisms for invalid checkpoints or validator faults.

This demo gives you a concrete narrative: **fast PoS execution + delayed PoW anchoring for stronger settlement assurances**.

---

## Project layout

```text
pow_pos_demo/
  core.py          # Tendermint-style PoS, Bitcoin-style PoW, bridge logic
  simulation.py    # Deterministic demo scenario + security report builder
scripts/
  run_demo.py      # CLI entrypoint
tests/
  test_demo.py     # Unit tests
```
