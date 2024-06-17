package main

import (
	"context"
	"flag"
	"log"
	"sync"

	stegobot "github.com/Kurichi/stego-bot"
)

func main() {
	n := *flag.Int("n", 1, "Number of bots to run")

	ctx, cancel := context.WithCancel(context.Background())
	cfg := stegobot.NewConfig()

	wg := &sync.WaitGroup{}
	for range n {
		b, err := stegobot.New(ctx, cfg)
		if err != nil {
			log.Fatalf("%+v\n", err)
		}

		wg.Add(1)
		go func() {
			if err := b.Run(); err != nil {
				log.Fatalf("%+v\n", err)
				cancel()
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
