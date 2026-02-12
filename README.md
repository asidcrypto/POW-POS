# PoS L2 Anchored to PoW (Go Demo)

This repo contains a **Go-only architecture demo** for:

- Tendermint-style **PoS L2 finality** (block commit requires >2/3 stake signing).
- Periodic L2 **checkpoint anchoring** to a Bitcoin-like PoW chain.
- Security progression from `PoS-final` to `PoW-final` after required confirmations.

It is a deterministic simulator for explaining design and tradeoffs.

## What this proves

- You can combine fast PoS execution with slower PoW settlement anchoring.
- Not every L2 block needs direct PoW settlement if you checkpoint at intervals.
- A clear policy can define when an anchored state is considered final.

## What this does not prove

- Full production trustlessness by itself.
- Real Bitcoin or CometBFT networking.
- Fraud/validity proof systems and adversarial bridge hardening.

---

## Quickstart (this repository)

```bash
go test ./...
go run ./cmd/powposdemo
```

Example with custom settings:

```bash
go run ./cmd/powposdemo \
  --l2-blocks 15 \
  --checkpoint-interval 3 \
  --pow-confirmations 2 \
  --mine-btc-every 2
```

---

## Zero-to-demo on macOS (copy/paste)

If you want to start from an empty folder:

1) Install Go (if needed):

```bash
brew install go
go version
```

2) Create project and initialize module:

```bash
mkdir -p pow-pos-go-demo
cd pow-pos-go-demo
go mod init pow-pos-go-demo
mkdir -p demo cmd/powposdemo
```

3) Create files (copy from this repo):
- `demo/core.go`
- `demo/simulation.go`
- `demo/demo_test.go`
- `cmd/powposdemo/main.go`

4) Run:

```bash
go test ./...
go run ./cmd/powposdemo
```

---

## Project layout

```text
cmd/powposdemo/main.go  # CLI runner
demo/core.go            # PoS chain, PoW chain, checkpoint bridge
demo/simulation.go      # Scenario runner + report builder
demo/demo_test.go       # Unit tests
```

---

## Mapping this to a real chain

For a production version aligned to your idea:

1. Build a real app chain with Cosmos SDK + CometBFT.
2. Submit checkpoints to Bitcoin regtest/testnet/mainnet using a relayer.
3. Add checkpoint verification logic and reorg handling.
4. Add validator slashing/dispute paths for invalid behavior.
