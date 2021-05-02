package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"

	"github.com/urfave/cli"
)

func Action(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("args is not found")
	}
	profiles, err := LoadProfile()
	if err != nil {
		return err
	}
	end := make(chan bool)
	outs := NewOutputs(len(profiles))

	concurrency := c.Int("concurrency")
	group := &sync.WaitGroup{}
	execGroup := &sync.WaitGroup{}
	go func() {
		endCh := make(chan struct{}, concurrency)
		for _, p := range profiles {
			endCh <- struct{}{}
			ch, err := outs.Allocate(p, c.String("output"), c.Bool("strip-prefix"), group)
			if err != nil {
				fmt.Println(err)
				return
			}
			c := NewCommand(p, c.Args())
			err = c.Exec(endCh, ch, execGroup)
			if err != nil {
				fmt.Println(err)
			}
		}
		close(end)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	select {
	case <-ctx.Done():
		return nil
	case <-end:
		execGroup.Wait()
		outs.Close()
		group.Wait()
		return nil
	}
}

func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = "awswrap"
	app.Usage = "awswrap"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "output, o",
			Usage: "output file template.",
		},
		cli.BoolFlag{
			Name:  "strip-prefix, s",
			Usage: "strip-prefix",
		},
		cli.IntFlag{
			Name:  "concurrency, c",
			Usage: "concurrency for command. default is cpu count",
			Value: runtime.NumCPU(),
		},
	}
	app.Action = Action
	return app
}
