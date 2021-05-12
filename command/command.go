package command

import (
	"bufio"
	"os"
	"os/exec"

	"github.com/wurelicke/awswrap/profile"
)

type Command struct {
	*exec.Cmd
}

func NewCommand(p profile.Profile, args []string) *Command {
	/* #nosec */
	c := exec.Command(args[0], args[1:]...)
	c.Env = append(os.Environ(), "AWS_PROFILE="+p.Name)
	if p.Region != "" {
		c.Env = append(c.Env, "AWS_REGION="+p.Region)
	}
	return &Command{c}
}

func (c *Command) Exec(ch chan<- string) error {
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

	scout := bufio.NewScanner(stdout)
	go func() {
		for scout.Scan() {
			ch <- scout.Text()
		}
		stdout.Close()
	}()

	scerr := bufio.NewScanner(stderr)
	go func() {
		for scerr.Scan() {
			ch <- scerr.Text()
		}
		stderr.Close()
	}()

	return c.Wait()
}
