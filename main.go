package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/imega/snake-game/ai"
	"github.com/imega/snake-game/snake"
	"github.com/imega/snake-game/state"
)

func main() {
	p := state.Parameters{}

	flag.IntVar(&p.Speed, "s", 100, "snake speed limit")
	flag.IntVar(&p.MaxInstance, "i", 1000, "max number of instances in epoch")
	flag.Float64Var(&p.MutationRate, "r", 0.1, "mutation rate on the weights of synapses")
	flag.Float64Var(&p.MutationRange, "n", 0.5, "interval of the mutation changes on the synapse weight")
	flag.IntVar(&p.MaxSnakeSteps, "t", 200, "max snake stept without eat")
	flag.IntVar(&p.MinScoreEpoch, "m", 0, "min score in epoch")
	flag.StringVar(&p.PrefixFilename, "p", "", "prefix filename with brain")
	flag.BoolVar(&p.Silent, "q", false, "start in silent mode")
	flag.BoolVar(&p.Human, "h", false, "start in human mode")
	flag.BoolVar(&p.CreateBrain, "c", false, "create empty brain")
	flag.Parse()

	if p.CreateBrain {
		if err := ai.CreateBrain(p); err != nil {
			fmt.Printf("failed to create brain, %s\n", err)
			usage()
			os.Exit(1)
		}

		prefix := ""
		if p.PrefixFilename != "" {
			prefix = p.PrefixFilename + "-"
		}

		fmt.Printf("brain created: %sbrain-0.json\n", prefix)
		os.Exit(0)
	}

	ch := make(chan state.SnakeGame)
	statCh := make(chan state.Stat)

	if !p.Human {
		args := flag.Args()
		if len(args) == 0 {
			fmt.Printf("empty args filename\n")
			usage()
			os.Exit(1)
		}

		p.BrainFilename = args[0]

		go func() {
			if err := ai.New(p, ch, snake.KeyboardEventsChan, statCh); err != nil {
				fmt.Printf("failed to start, %s\n", err)
				usage()
				os.Exit(1)
			}
		}()
	}

	snake.NewGame().Start(p, ch, statCh)
}

func usage() {
	fmt.Fprintf(
		flag.CommandLine.Output(),
		"\nUsage: %s [-chiqmnprst] [<prefix>brain-<score>.json]\n",
		os.Args[0],
	)
	flag.PrintDefaults()
}
