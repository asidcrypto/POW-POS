package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/asidcrypto/POW-POS/demo"
)

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func main() {
	cfg := demo.DefaultRunConfig()

	flag.Int64Var(&cfg.L2Blocks, "l2-blocks", cfg.L2Blocks, "Number of L2 blocks to produce")
	flag.Int64Var(&cfg.CheckpointInterval, "checkpoint-interval", cfg.CheckpointInterval, "Anchor every N L2 blocks to Bitcoin")
	flag.Int64Var(&cfg.RequiredPoWConfirms, "pow-confirmations", cfg.RequiredPoWConfirms, "Bitcoin confirmations required before checkpoint is PoW-final")
	flag.Int64Var(&cfg.MineBtcEveryL2Blocks, "mine-btc-every", cfg.MineBtcEveryL2Blocks, "Mine one Bitcoin block every N L2 blocks")
	flag.Parse()

	l2, btc, bridge, events, err := demo.RunDemo(cfg)
	if err != nil {
		log.Fatalf("run demo: %v", err)
	}

	fmt.Println("\n=== PoS-on-PoW Demo (Tendermint-style + Bitcoin-style) ===\n")
	for _, event := range events {
		fmt.Printf("[%02d] %s\n", event.Step, event.Message)
	}

	report := demo.BuildSecurityReport(l2, btc, bridge)
	fmt.Println("\nSecurity report:")

	w := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "height\thash\tpos_final\tanchored\tpow_conf\tpow_final")
	fmt.Fprintln(w, "------\t----\t---------\t--------\t--------\t---------")

	var powFinalCount int64
	for _, row := range report {
		if row.PoWFinal {
			powFinalCount++
		}
		fmt.Fprintf(
			w,
			"%d\t%s\t%s\t%s\t%d\t%s\n",
			row.L2Height,
			row.L2Hash,
			yesNo(row.PoSFinal),
			yesNo(row.AnchoredToPoW),
			row.PoWConfirmations,
			yesNo(row.PoWFinal),
		)
	}
	w.Flush()

	fmt.Printf(
		"\nSummary: %d L2 blocks committed by PoS, %d blocks additionally secured by PoW checkpoints.\n",
		len(report),
		powFinalCount,
	)
	fmt.Println(
		"Note: This is an architecture demo; production systems need real CometBFT/Cosmos consensus, bridge proofs, and adversarial testing.",
	)
}
