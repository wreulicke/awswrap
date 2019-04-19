package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
)

type Outputs struct {
	outs map[string]Output
}

type Output struct {
	ch   chan<- string
	done <-chan bool
}

func NewOutputs() *Outputs {
	return &Outputs{outs: map[string]Output{}}
}

func (o *Outputs) Allocate(p Profile, output string, stripPrefix bool) (chan<- string, error) {
	decorator := makeDecorator(stripPrefix, p.Name)
	if output == "" {
		stdch := make(chan string)
		go func() {
			for {
				s, more := <-stdch
				if s != "" {
					fmt.Println(decorator(s))
				}
				if !more {
					return
				}
			}
		}()
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
	done := make(chan bool)
	ch := outputChannel(decorator, f, done)
	o.outs[path] = Output{ch, done}
	return ch, nil
}

func (o *Outputs) Close() {
	for _, out := range o.outs {
		close(out.ch)
	}
	for _, out := range o.outs {
		<-out.done
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

func outputChannel(decorator func(string) string, writer io.WriteCloser, done chan<- bool) chan<- string {
	ch := make(chan string)
	w := bufio.NewWriter(writer)
	go func() {
		for {
			s, more := <-ch
			fmt.Fprintln(w, decorator(s))
			if !more {
				w.Flush()
				writer.Close()
				done <- true
				return
			}
		}
	}()
	return ch
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
