package main

import (
	"os"
	"os/exec"
	"sync"
)

type Command struct {
	*exec.Cmd
}

func NewCommand(p Profile, args []string) *Command {
	c := exec.Command(args[0], args[1:]...)
	c.Env = append(os.Environ(), "AWS_PROFILE="+p.Name)
	c.Stdout = os.Stdout
	c.Stderr = os.Stdout
	return &Command{Cmd: c}
}

func (c *Command) Exec(end <-chan struct{}, ch chan<- string, group *sync.WaitGroup) error {
	if err := c.Start(); err != nil {
		return err
	}

	return c.Wait()
}
