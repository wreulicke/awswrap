package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wurelicke/awswrap/command"
	"github.com/wurelicke/awswrap/profile"
)

func NewRootCommand() *cobra.Command {
	var output string
	var stripPrefix bool
	var concurrency int
	cmd := &cobra.Command{
		Use:   "awswrap",
		Short: "awswrap is wrapper command to execute command to all aws environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := profile.LoadProfile()
			if err != nil {
				return err
			}
			end := make(chan bool)
			outs := command.NewOutputs(len(profiles))

			group := &sync.WaitGroup{}
			execGroup := &sync.WaitGroup{}
			go func() {
				semaphore := make(chan struct{}, concurrency)
				for _, p := range profiles {
					semaphore <- struct{}{}
					ch, err := outs.Allocate(p, output, stripPrefix, group)
					if err != nil {
						fmt.Println(err)
						return
					}
					c := command.NewCommand(p, args)
					err = c.Exec(semaphore, ch, execGroup)
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
		},
		Args: cobra.MinimumNArgs(1),
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file template")
	cmd.Flags().BoolVarP(&stripPrefix, "strip-prefix", "s", false, "strip prefix")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", runtime.NumCPU(), "concurrency for command. defualt is cpu count")

	return cmd
}
