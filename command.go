package main

import (
	"bufio"
	"os"
	"os/exec"
)

type Command struct {
	*exec.Cmd
}

func NewCommand(p Profile, args []string) *Command {
	c := exec.Command(args[0], args[1:]...)
	c.Env = append(os.Environ(), "AWS_PROFILE="+p.Name)
	return &Command{c}
}

func (c *Command) Exec(end chan<- bool, ch chan<- string) error {
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}

	scout := bufio.NewScanner(stdout)
	go func() {
		for scout.Scan() {
			ch <- scout.Text()
		}
		stdout.Close()
		end <- true
	}()

	scerr := bufio.NewScanner(stderr)
	go func() {
		for scerr.Scan() {
			ch <- scerr.Text()
		}
		stderr.Close()
	}()

	if err := c.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return err
	}
	return nil
}
