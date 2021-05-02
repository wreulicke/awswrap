package command

import (
	"bufio"
	"os"
	"os/exec"
	"sync"

	"github.com/wurelicke/awswrap/profile"
)

type Command struct {
	*exec.Cmd
}

func NewCommand(p profile.Profile, args []string) *Command {
	c := exec.Command(args[0], args[1:]...)
	c.Env = append(os.Environ(), "AWS_PROFILE="+p.Name)
	return &Command{c}
}

func (c *Command) Exec(end <-chan struct{}, ch chan<- string, group *sync.WaitGroup) error {
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}

	if err := c.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return err
	}

	group.Add(2)
	scout := bufio.NewScanner(stdout)
	go func() {
		for scout.Scan() {
			ch <- scout.Text()
		}
		stdout.Close()
		group.Done()
		<-end
	}()

	scerr := bufio.NewScanner(stderr)
	go func() {
		for scerr.Scan() {
			ch <- scerr.Text()
		}
		stderr.Close()
		group.Done()
	}()

	return nil
}