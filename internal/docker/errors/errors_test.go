// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package errors

import (
	"errors"
	"testing"
)

func TestTrimExtraInfo(t *testing.T) {
	const (
		before = `Error response from daemon: container encountered an error during CreateProcess: failure in a Windows system call: The system cannot find the file specified. (0x2) extra info: { "User":"ContainerUser" }`
		after  = `Error response from daemon: container encountered an error during CreateProcess: failure in a Windows system call: The system cannot find the file specified.`
	)
	errBefore := errors.New(before)
	errAfter := TrimExtraInfo(errBefore)
	if errAfter.Error() != after {
		t.Errorf("Expect trimmed image")
	}
}
