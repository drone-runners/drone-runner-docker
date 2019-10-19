// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import "testing"

func TestParseImage(t *testing.T) {
	tests := []struct {
		image     string
		canonical string
		domain    string
		latest    bool
		err       bool
	}{
		{
			image:     "golang",
			canonical: "docker.io/library/golang:latest",
			domain:    "docker.io",
			latest:    true,
		},
		{
			image:     "golang:1.11",
			canonical: "docker.io/library/golang:1.11",
			domain:    "docker.io",
			latest:    false,
		},
		{
			image: "",
			err:   true,
		},
	}

	for _, test := range tests {
		canonical, domain, latest, err := parseImage(test.image)
		if test.err {
			if err == nil {
				t.Errorf("Expect error parsing image %s", test.image)
			}
			continue
		}
		if err != nil {
			t.Error(err)
		}
		if got, want := canonical, test.canonical; got != want {
			t.Errorf("Want image %s, got %s", want, got)
		}
		if got, want := domain, test.domain; got != want {
			t.Errorf("Want image domain %s, got %s", want, got)
		}
		if got, want := latest, test.latest; got != want {
			t.Errorf("Want image latest %v, got %v", want, got)
		}
	}
}
