package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

var envSharedCredentialsFile = "AWS_SHARED_CREDENTIALS_FILE"
var envAWSConfigFile = "AWS_CONFIG_FILE"

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
	// TODO ignore not-exist file
	sharedCredentialsPath := envOrDefault(envSharedCredentialsFile, filepath.Join(home, ".aws", "credentials"))
	configFilePath := envOrDefault(envAWSConfigFile, filepath.Join(home, ".aws", "config"))
	f, err := ini.Load(sharedCredentialsPath, configFilePath)
	if err != nil {
		return nil, err
	}
	return profiles(f), nil
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
