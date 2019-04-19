package main

import (
	"errors"
	"log"
	"os"
	"os/signal"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "awswrap"
	app.Usage = "awswrap"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "output, o",
			Usage: "output",
		},
		cli.BoolFlag{
			Name:  "strip-prefix, s",
			Usage: "strip-prefix",
		},
	}
	app.Action = func(c *cli.Context) error {
		args := c.Args()
		if len(args) == 0 {
			return errors.New("args is not found")
		}
		profiles, err := LoadProfile()
		if err != nil {
			return err
		}
		end := make(chan bool)
		numProfiles := len(profiles)

		outs := NewOutputs()
		defer outs.Close()
		for _, p := range profiles {
			ch, err := outs.Allocate(p, c.String("output"), c.Bool("strip-prefix"))
			if err != nil {
				return err
			}
			c := NewCommand(p, args)
			err = c.Exec(end, ch)
			if err != nil {
				return err
			}
		}

		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		for {
			select {
			case _ = <-interrupt:
				os.Exit(0) // TODO more handling
			case _ = <-end:
				numProfiles--
				if numProfiles == 0 {
					return nil
				}
			}
		}

	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
