package cmd

import (
	"context"
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
			outs := command.NewOutputs()

			execGroup := &sync.WaitGroup{}
			semaphore := make(chan struct{}, concurrency)
			for _, p := range profiles {
				semaphore <- struct{}{}
				ch, err := outs.Allocate(p, output, stripPrefix)
				if err != nil {
					return err
				}
				c := command.NewCommand(p, args)
				execGroup.Add(1)
				go func() {
					defer execGroup.Done()
					err = c.Exec(ch)
					if err != nil {
						ch <- err.Error()
					}
					<-semaphore
				}()
			}

			end := make(chan struct{})
			go func() {
				execGroup.Wait()
				outs.Close()
				close(end)
			}()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			select {
			case <-ctx.Done():
				return nil
			case <-end:
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
