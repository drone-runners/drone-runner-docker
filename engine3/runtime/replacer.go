// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package runtime

import (
	"io"
	"strings"

	"github.com/drone-runners/drone-runner-docker/engine3/engine"
)

// replacer is an io.Writer that finds and masks sensitive data.
type replacer struct {
	w io.WriteCloser
	r *strings.Replacer
}

// newReplacer returns a replacer that wraps io.Writer w.
func newReplacer(w io.WriteCloser, secrets []*engine.Secret) io.WriteCloser {
	var oldnew []string
	for _, secret := range secrets {
		v := string(secret.Data)
		if len(v) == 0 || secret.Mask == false {
			continue
		}

		for _, part := range strings.Split(v, "\n") {
			part = strings.TrimSpace(part)

			// avoid masking empty or single character
			// strings.
			if len(part) < 2 {
				continue
			}

			masked := "******"
			oldnew = append(oldnew, part)
			oldnew = append(oldnew, masked)
		}
	}
	if len(oldnew) == 0 {
		return w
	}
	return &replacer{
		w: w,
		r: strings.NewReplacer(oldnew...),
	}
}

// Write writes p to the base writer. The method scans for any
// sensitive data in p and masks before writing.
func (r *replacer) Write(p []byte) (n int, err error) {
	_, err = r.w.Write([]byte(r.r.Replace(string(p))))
	return len(p), err
}

// Close closes the base writer.
func (r *replacer) Close() error {
	return r.w.Close()
}
