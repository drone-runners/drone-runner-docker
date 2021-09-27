// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

// Package client provides a client for using the runner API.
package livelog

import (
	"context"
	"io"
)

// Error represents a json-encoded API error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// A Client manages communication with the runner.
type Client interface {
	// Batch batch writes logs to the build logs.
	Batch(ctx context.Context, key string, lines []*Line) error

	// Upload uploads the full logs to the server.
	Upload(ctx context.Context, key string, r io.Reader) error

	// Open opens the stream to write logs
	Open(ctx context.Context, key string) error

	// Close closes the data stream
	Close(ctx context.Context, key string) error
}