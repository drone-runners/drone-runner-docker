// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package jsonmessage

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *jsonError) Error() string {
	return e.Message
}

type jsonMessage struct {
	ID       string        `json:"id"`
	Status   string        `json:"status"`
	Error    *jsonError    `json:"errorDetail"`
	Progress *jsonProgress `json:"progressDetail"`
}

type jsonProgress struct {
}

// Copy copies a json message string to the io.Writer.
func Copy(in io.Reader, out io.Writer) error {
	dec := json.NewDecoder(in)
	for {
		var jm jsonMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if jm.Error != nil {
			if jm.Error.Code == 401 {
				return fmt.Errorf("authentication is required")
			}
			return jm.Error
		}

		if jm.Progress != nil {
			continue
		}
		if jm.ID == "" {
			fmt.Fprintf(out, "%s\n", jm.Status)
		} else {
			fmt.Fprintf(out, "%s: %s\n", jm.ID, jm.Status)
		}
	}
	return nil
}
