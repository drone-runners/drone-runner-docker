// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package runtime

import (
	"bytes"
	"io"
	"testing"

	"github.com/drone-runners/drone-runner-docker/engine/experimental/engine"
)

func TestReplace(t *testing.T) {
	secrets := []*engine.Secret{
		{Name: "DOCKER_USERNAME", Data: []byte("octocat"), Mask: false},
		{Name: "DOCKER_PASSWORD", Data: []byte("correct-horse-batter-staple"), Mask: true},
		{Name: "DOCKER_EMAIL", Data: []byte(""), Mask: true},
	}

	buf := new(bytes.Buffer)
	w := newReplacer(&nopCloser{buf}, secrets)
	w.Write([]byte("username octocat password correct-horse-batter-staple"))
	w.Close()

	if got, want := buf.String(), "username octocat password ******"; got != want {
		t.Errorf("Want masked string %s, got %s", want, got)
	}
}

func TestReplaceMultiline(t *testing.T) {
	key := `
-----BEGIN PRIVATE KEY-----
MIIBVQIBADANBgkqhkiG9w0BAQEFAASCAT8wggE7AgEAAkEA0SC5BIYpanOv6wSm
dHVVMRa+6iw/0aJpT9/LKcZ0XYQ43P9Vwn8c46MDvFJ+Uy41FwbxT+QpXBoLlp8D
sJY/dQIDAQABAkAesoL2GwtxSNIF2YTli2OZ9RDJJv2nNAPpaZxU4YCrST1AXGPB
tFm0LjYDDlGJ448syKRpdypAyCR2LidwrVRxAiEA+YU5Zv7bOwODCsmtQtIfBfhu
6SMBGMDijK7OYfTtjQsCIQDWjvly6b6doVMdNjqqTsnA8J1ShjSb8bFXkMels941
fwIhAL4Rr7I3PMRtXmrfSa325U7k+Yd59KHofCpyFiAkNLgVAiB8JdR+wnOSQAOY
loVRgC9LXa6aTp9oUGxeD58F6VK9PwIhAIDhSxkrIatXw+dxelt8DY0bEdDbYzky
r9nicR5wDy2W
-----END PRIVATE KEY-----`

	line := `> MIIBVQIBADANBgkqhkiG9w0BAQEFAASCAT8wggE7AgEAAkEA0SC5BIYpanOv6wSm`

	secrets := []*engine.Secret{
		{Name: "SSH_KEY", Data: []byte(key), Mask: true},
	}

	buf := new(bytes.Buffer)
	w := newReplacer(&nopCloser{buf}, secrets)
	w.Write([]byte(line))
	w.Close()

	if got, want := buf.String(), "> ******"; got != want {
		t.Errorf("Want masked string %s, got %s", want, got)
	}
}

func TestReplaceMultilineJson(t *testing.T) {
	key := `{
  "token":"MIIBVQIBADANBgkqhkiG9w0BAQEFAASCAT8wggE7AgEAAkEA0SC5BIYpanOv6wSm"
}`

	line := `{
  "token":"MIIBVQIBADANBgkqhkiG9w0BAQEFAASCAT8wggE7AgEAAkEA0SC5BIYpanOv6wSm"
}`

	secrets := []*engine.Secret{
		{Name: "JSON_KEY", Data: []byte(key), Mask: true},
	}

	buf := new(bytes.Buffer)
	w := newReplacer(&nopCloser{buf}, secrets)
	w.Write([]byte(line))
	w.Close()

	if got, want := buf.String(), "{\n  ******\n}"; got != want {
		t.Errorf("Want masked string %s, got %s", want, got)
	}
}

// this test verifies that if there are no secrets to scan and
// mask, the io.WriteCloser is returned as-is.
func TestReplaceNone(t *testing.T) {
	secrets := []*engine.Secret{
		{Name: "DOCKER_USERNAME", Data: []byte("octocat"), Mask: false},
		{Name: "DOCKER_PASSWORD", Data: []byte("correct-horse-batter-staple"), Mask: false},
	}

	buf := new(bytes.Buffer)
	w := &nopCloser{buf}
	r := newReplacer(w, secrets)
	if w != r {
		t.Errorf("Expect buffer returned with no replacer")
	}
}

type nopCloser struct {
	io.Writer
}

func (*nopCloser) Close() error {
	return nil
}
