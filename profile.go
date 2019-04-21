package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

const envSharedCredentialsFile = "AWS_SHARED_CREDENTIALS_FILE"

const envAWSConfigFile = "AWS_CONFIG_FILE"

func envOrDefault(key string, d string) string {
	v := os.Getenv(key)
	if v == "" {
		return d
	}
	return v
}

func LoadProfile() (map[string]Profile, error) { // TODO use struct
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	sharedCredentialsPath := envOrDefault(envSharedCredentialsFile, filepath.Join(home, ".aws", "credentials"))
	configFilePath := envOrDefault(envAWSConfigFile, filepath.Join(home, ".aws", "config"))
	files := []io.ReadCloser{}
	f, err := os.Open(sharedCredentialsPath)
	if err != nil {
		// credentials is required.
		return nil, err
	}
	files = append(files, f)

	f, err = os.Open(configFilePath)
	// config file is not required.
	if err == nil {
		files = append(files, f)
	}

	if len(files) == 1 {
		iniFile, err := ini.Load(files[0])
		if err != nil {
			return nil, err
		}
		return profiles(iniFile), nil
	}

	iniFile, err := ini.Load(files[0], files[1])
	if err != nil {
		return nil, err
	}
	return profiles(iniFile), nil
}

type Profile struct {
	Region string
	Name   string
}

func profiles(i *ini.File) map[string]Profile { // TODO use struct
	profiles := map[string]Profile{}
	for _, s := range i.Sections() {
		name := s.Name()
		if strings.HasPrefix(name, "profile ") {
			name = strings.TrimPrefix(name, "profile ")
		} else if name == "DEFAULT" {
			// ignore default
			continue
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
