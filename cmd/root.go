package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wurelicke/awswrap/command"
	"github.com/wurelicke/awswrap/profile"
)

// Getting from https://docs.aws.amazon.com/general/latest/gr/rande.html
// copy($$("#w421aac11b7b9c13 > tbody > tr > td:nth-child(2)").map(t => t.innerHTML).map(e => `${e.replace(/-(.)/g, (f, v) => v.toUpperCase())} = "${e}"`).join("\n")).
var (
	usEast2      = "us-east-2"
	usEast1      = "us-east-1"
	usWest1      = "us-west-1"
	usWest2      = "us-west-2"
	afSouth1     = "af-south-1"
	apEast1      = "ap-east-1"
	apSouth1     = "ap-south-1"
	apNortheast3 = "ap-northeast-3"
	apNortheast2 = "ap-northeast-2"
	apSoutheast1 = "ap-southeast-1"
	apSoutheast2 = "ap-southeast-2"
	apNortheast1 = "ap-northeast-1"
	caCentral1   = "ca-central-1"
	cnNorth1     = "cn-north-1"
	cnNorthwest1 = "cn-northwest-1"
	euCentral1   = "eu-central-1"
	euWest1      = "eu-west-1"
	euWest2      = "eu-west-2"
	euSouth1     = "eu-south-1"
	euWest3      = "eu-west-3"
	euNorth1     = "eu-north-1"
	meSouth1     = "me-south-1"
	saEast1      = "sa-east-1"
)

// all regions
// copy("var regions = []string{\n" + $$("#w421aac11b7b9c13 > tbody > tr > td:nth-child(2)").map(t => t.innerHTML).map(e => `${e.replace(/-(.)/g, (f, v) => v.toUpperCase())}"`).join("\n") "\n}").
var allRegions = []string{
	usEast2,
	usEast1,
	usWest1,
	usWest2,
	afSouth1,
	apEast1,
	apSouth1,
	apNortheast3,
	apNortheast2,
	apSoutheast1,
	apSoutheast2,
	apNortheast1,
	caCentral1,
	cnNorth1,
	cnNorthwest1,
	euCentral1,
	euWest1,
	euWest2,
	euSouth1,
	euWest3,
	euNorth1,
	meSouth1,
	saEast1,
}

func makeDecorator(stripPrefix bool, profile string, region string) func(string) string {
	if stripPrefix && region == "" {
		return func(str string) string {
			return str
		}
	}

	if region == "" {
		prefix := fmt.Sprintf("[%s]\t", profile)
		return func(s string) string {
			return prefix + s
		}
	}
	prefix := fmt.Sprintf("[%s][%s]\t", profile, region)
	return func(s string) string {
		return prefix + s
	}
}

func NewRootCommand() *cobra.Command { // nolint:cyclop
	var output string
	var stripPrefix bool
	var concurrency int
	var global bool
	var regions []string
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
			if global {
				regions = []string{}
				for _, r := range allRegions {
					if !strings.HasPrefix(r, "cn") {
						regions = append(regions, r)
					}
				}
			}

			for _, p := range profiles {
				for _, region := range regions {
					semaphore <- struct{}{}
					p := p.WithRegion(region)
					decorator := makeDecorator(stripPrefix, p.Name, p.Region)
					ch, err := outs.Allocate(p, output, decorator)
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
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", runtime.NumCPU()*4, "concurrency for command. default is cpu count * 4")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "execute command to global")
	cmd.Flags().StringSliceVarP(&regions, "region", "r", []string{""}, "region")

	return cmd
}
