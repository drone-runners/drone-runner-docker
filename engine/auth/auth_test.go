// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package auth

import (
	"bytes"
	"encoding/base64"
	"os"
	"testing"

	"github.com/drone-runners/drone-runner-docker/engine"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	got, err := ParseString(sample)
	if err != nil {
		t.Error(err)
		return
	}
	want := []*engine.Auth{
		{
			Address:  "index.docker.io",
			Username: "octocat",
			Password: "correct-horse-battery-staple",
		},
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

func TestParseGCR(t *testing.T) {
	got, err := ParseFile("testdata/config_gcr.json")
	if err != nil {
		t.Error(err)
		return
	}
	want := []*engine.Auth{
		{
			Address:  "gcr.io",
			Username: "_json_key",
			Password: "xxx:bar\n",
		},
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

func TestParseErr(t *testing.T) {
	_, err := ParseString("")
	if err == nil {
		t.Errorf("Expect unmarshal error")
	}
}

func TestParseFile(t *testing.T) {
	got, err := ParseFile("./testdata/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	want := []*engine.Auth{
		{
			Address:  "index.docker.io",
			Username: "octocat",
			Password: "correct-horse-battery-staple",
		},
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf(diff)
	}
}

func TestParseFileErr(t *testing.T) {
	_, err := ParseFile("./testdata/x.json")
	if _, ok := err.(*os.PathError); !ok {
		t.Errorf("Expect error when file does not exist")
	}
}

func Test_encodeDecode(t *testing.T) {
	username := "octocat"
	password := "correct-horse-battery-staple"

	encoded := encode(username, password)
	decodedUsername, decodedPassword := decode(encoded)
	if got, want := decodedUsername, username; got != want {
		t.Errorf("Want decoded username %s, got %s", want, got)
	}
	if got, want := decodedPassword, password; got != want {
		t.Errorf("Want decoded password %s, got %s", want, got)
	}
}

func Test_decodeInvalid(t *testing.T) {
	username, password := decode("b2N0b2NhdDp==")
	if username != "" || password != "" {
		t.Errorf("Expect decoding error")
	}
}

func TestEncode(t *testing.T) {
	username := "octocat"
	password := "correct-horse-battery-staple"
	result := Encode(username, password)
	got, err := base64.URLEncoding.DecodeString(result)
	if err != nil {
		t.Error(err)
		return
	}
	want := []byte(`{"username":"octocat","password":"correct-horse-battery-staple"}`)
	if bytes.Equal(got, want) == false {
		t.Errorf("Could not decode credential header")
	}
}

func TestMarshal(t *testing.T) {
	auths := []*engine.Auth{
		{
			Address:  "index.docker.io",
			Username: "octocat",
			Password: "correct-horse-battery-staple",
		},
	}
	got, _ := Marshal(auths)
	want := []byte(`{"auths":{"index.docker.io":{"auth":"b2N0b2NhdDpjb3JyZWN0LWhvcnNlLWJhdHRlcnktc3RhcGxl"}}}`)
	if bytes.Equal(got, want) == false {
		t.Errorf("Could not decode credential header")
	}
}

var sample = `{
	"auths": {
		"https://index.docker.io/v1/": {
			"auth": "b2N0b2NhdDpjb3JyZWN0LWhvcnNlLWJhdHRlcnktc3RhcGxl"
		}
	}
}`
