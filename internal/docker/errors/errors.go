// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package errors

import (
	"errors"
	"strings"
)

// TrimExtraInfo is a helper function that trims extra information
// from a Docker error. Specifically, on Windows, this can expose
// environment variables and other sensitive data.
func TrimExtraInfo(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	i := strings.Index(s, "extra info:")
	if i > 0 {
		s = s[:i]
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, "(0x2)")
		s = strings.TrimSpace(s)
		return errors.New(s)
	}
	return err
}
