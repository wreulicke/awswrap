package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	"gopkg.in/ini.v1"
)

var envSharedCredentialsFile = "AWS_SHARED_CREDENTIALS_FILE"
var envAWSConfigFile = "AWS_CONFIG_FILE"

type fallback func() (string, error)

func envOrDefault(key string, d string) string {
	v := os.Getenv(key)
	if v == "" {
		return d
	}
	return v
}

func LoadProfile() (*ini.File, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	// TODO ignore not-exist file
	sharedCredentialsPath := envOrDefault(envSharedCredentialsFile, filepath.Join(home, ".aws", "credentials"))
	configFilePath := envOrDefault(envAWSConfigFile, filepath.Join(home, ".aws", "config"))
	return ini.Load(sharedCredentialsPath, configFilePath)
}

type Profile struct {
	Region string
	Name   string
}

func Profiles(i *ini.File) map[string]Profile {
	profiles := map[string]Profile{}
	for _, s := range i.Sections() {
		name := s.Name()
		if strings.HasPrefix(name, "profile ") {
			name = strings.TrimPrefix(name, "profile ")
		} else if name == "DEFAULT" {
			name = "default"
		}
		if p, found := profiles[name]; found {
			if s.HasKey("region") {
				k, _ := s.GetKey("region")
				p.Region = k.String()
			}
		} else {
			p := Profile{
				Name: name,
			}
			if s.HasKey("region") {
				k, _ := s.GetKey("region")
				p.Region = k.String()
			}
			profiles[name] = p
		}
	}

	return profiles
}

func start(p string, args []string, end chan bool, ch chan<- string, decorator func(string) string) error {
	c := exec.Command(args[0], args[1:]...)
	c.Env = append(os.Environ(), "AWS_PROFILE="+p)
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
			ch <- decorator(scout.Text())
		}
		stdout.Close()
		end <- true
	}()

	scerr := bufio.NewScanner(stderr)
	go func() {
		for scerr.Scan() {
			ch <- decorator(scerr.Text())
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

func outputChannel(writer io.Writer, closer io.Closer) chan<- string {
	ch := make(chan string)
	go func() {
		for {
			s, more := <-ch
			fmt.Fprintln(writer, s)
			if !more {
				closer.Close()
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
		f, err := LoadProfile()
		if err != nil {
			return err
		}
		profiles := Profiles(f)
		end := make(chan bool)
		numProfiles := len(profiles)
		openWriter := map[string]chan<- string{}

		defer func() {
			for _, ch := range openWriter {
				close(ch)
			}
		}()

		stdch := make(chan string)
		go func() {
			for {
				s, more := <-stdch
				fmt.Println(s)
				if !more {
					return
				}
			}
		}()
		for _, p := range profiles {
			var ch chan<- string = stdch
			decorator := makeDecorator(c.Bool("strip-prefix"), p.Name)
			if c.IsSet("output") {
				output := c.String("output")
				t := template.New(p.Name)
				temp, err := t.Parse(output)
				if err != nil {
					return err
				}
				buf := &bytes.Buffer{}
				err = temp.Execute(buf, p)
				if err != nil {
					return err
				}
				path := buf.String()
				if o, found := openWriter[path]; found {
					ch = o
				} else {
					writer, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
					if err != nil {
						return err
					}

					ch = outputChannel(writer, writer)
					openWriter[path] = ch
				}
			}
			err := start(p.Name, args, end, ch, decorator)
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
