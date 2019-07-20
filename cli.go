package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"

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
	defer outs.Close()

	concurrency := c.Int("concurrency")

	endCh := make(chan struct{}, concurrency)
	go func() {
		for _, p := range profiles {
			endCh <- struct{}{}
			ch, err := outs.Allocate(p, c.String("output"), c.Bool("strip-prefix"))
			if err != nil {
				fmt.Println(err)
				return
			}
			c := NewCommand(p, c.Args())
			err = c.Exec(endCh, ch)
			if err != nil {
				fmt.Println(err)
			}
		}
		close(end)
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	for {
		select {
		case <-interrupt:
			os.Exit(0) // TODO more handling
		case <-end:
			return nil
		}
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
