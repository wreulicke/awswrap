package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"sync"
)

type Flusher interface {
	Flush() error
}

type FlushableWriteCloser struct {
	io.Writer
	io.Closer
	Flusher
}

type Outputs struct {
	outs     map[string]Output
	channels []chan string
}

type Output struct {
	ch chan<- string
}

func NewOutputs(len int) *Outputs {
	return &Outputs{
		outs: map[string]Output{},
	}
}

type NopFlusher struct {
}

func (NopFlusher) Flush() error {
	return nil
}

func (o *Outputs) add(ch chan string) {
	o.channels = append(o.channels, ch)
}

func (o *Outputs) Allocate(p Profile, output string, stripPrefix bool, group *sync.WaitGroup) (chan<- string, error) {
	decorator := makeDecorator(stripPrefix, p.Name)
	if output == "" {
		stdch := make(chan string)
		outputChannel(decorator, FlushableWriteCloser{
			Writer:  os.Stdout,
			Flusher: &NopFlusher{},
			Closer:  io.NopCloser(nil),
		}, stdch, group)
		o.add(stdch)
		return stdch, nil
	}
	path, err := makeOutputPath(output, p)
	if err != nil {
		return nil, err
	}
	if o, found := o.outs[path]; found {
		return o.ch, nil
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return nil, err
	}
	ch := make(chan string)
	o.add(ch)
	w := bufio.NewWriter(f)
	outputChannel(decorator, FlushableWriteCloser{
		Writer:  w,
		Flusher: w,
		Closer:  f,
	}, ch, group)
	o.outs[path] = Output{ch}
	return ch, nil
}

func (o *Outputs) Close() {
	for _, ch := range o.channels {
		close(ch)
	}
}

func makeOutputPath(src string, p Profile) (string, error) {
	t := template.New(p.Name)
	temp, err := t.Parse(src)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = temp.Execute(buf, p)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func outputChannel(decorator func(string) string, w FlushableWriteCloser, ch <-chan string, group *sync.WaitGroup) {
	group.Add(1)
	go func() {
		for {
			s, more := <-ch
			if more {
				fmt.Fprintln(w, decorator(s))
			}
			if !more {
				w.Flush()
				w.Close()
				group.Done()
				return
			}
		}
	}()
}

func makeDecorator(stripPrefix bool, p string) func(string) string {
	if stripPrefix {
		return func(str string) string {
			return str
		}
	}
	return func(str string) string {
		return fmt.Sprintf("[%s] \t%s", p, str)
	}
}
