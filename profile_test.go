package main

import (
	"os"
	"testing"
)

func TestLoadProfile(t *testing.T) {
	os.Setenv(envSharedCredentialsFile, "testdata/data")
	os.Setenv(envAWSConfigFile, "testdata/not_exists")
	p, err := LoadProfile()
	if err != nil {
		t.Error("Unexpected error", err)
		return
	}
	if p["test"].Region != "test" {
		t.Errorf("profile `test` contains region. actual: %s", p["test"].Region)
	}
	if p["test"].Name != "test" {
		t.Errorf("profile `test` is not Name is `test`. actual: %s", p["test"].Name)
	}
}

func TestLoadProfileWithBoth(t *testing.T) {
	os.Setenv(envSharedCredentialsFile, "testdata/data")
	os.Setenv(envAWSConfigFile, "testdata/data")
	p, err := LoadProfile()
	if err != nil {
		t.Errorf("Unexpected error. %+v", err)
		return
	}
	if p["test"].Region != "test" {
		t.Errorf("profile `test` contains region. actual: %s", p["test"].Region)
	}
	if p["test"].Name != "test" {
		t.Errorf("profile `test` is not Name is `test`. actual: %s", p["test"].Name)
	}
}

func TestLoadProfileWhenCredentialDoesNotExists(t *testing.T) {
	os.Setenv(envSharedCredentialsFile, "testdata/not_exists")
	os.Setenv(envAWSConfigFile, "testdata/data")
	_, err := LoadProfile()
	if os.IsNotExist(err) == false {
		if err != nil {
			t.Error("Unexpected error", err)
		} else {
			t.Error("There is unexpectedly no error", err)
		}
		return
	}
}
func TestLoadProfileWhenConfigFileIsNotExists(t *testing.T) {
	os.Setenv(envSharedCredentialsFile, "testdata/data")
	os.Setenv(envAWSConfigFile, "testdata/not_exists")
	_, err := LoadProfile()
	if err != nil {
		t.Error("Unexpected error", err)
	}
}

func TestLoadProfileWhenConfigAndCredentialsAreNotFound(t *testing.T) {
	os.Setenv(envSharedCredentialsFile, "testdata/not_exists")
	os.Setenv(envAWSConfigFile, "testdata/not_exists")
	_, err := LoadProfile()
	if err == nil {
		t.Error("There is unexpectedly no error")
	}
}
