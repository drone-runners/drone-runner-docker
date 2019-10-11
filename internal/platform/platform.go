// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

// Package platform contains code to provision and destroy server
// instances on the Digital Ocean cloud platform.
package platform

import (
	"context"
	"time"

	"github.com/drone/runner-go/logger"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type (
	// RegisterArgs provides arguments to register the SSH
	// public key with the account.
	RegisterArgs struct {
		Fingerprint string
		Name        string
		Data        string
		Token       string
	}

	// DestroyArgs provides arguments to destroy the server
	// instance.
	DestroyArgs struct {
		ID    int
		IP    string
		Token string
	}

	// ProvisionArgs provides arguments to provision instances.
	ProvisionArgs struct {
		Key    string
		Image  string
		Name   string
		Region string
		Size   string
		Token  string
	}

	// Instance represents a provisioned server instance.
	Instance struct {
		ID int
		IP string
	}
)

// Provision provisions the server instance.
func Provision(ctx context.Context, args ProvisionArgs) (Instance, error) {
	res := Instance{}
	req := &godo.DropletCreateRequest{
		Name:   args.Name,
		Region: args.Region,
		Size:   args.Size,
		Tags:   []string{"drone"},
		IPv6:   false,
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: args.Key},
		},
		Image: godo.DropletCreateImage{
			Slug: args.Image,
		},
	}

	logger := logger.FromContext(ctx).
		WithField("region", req.Region).
		WithField("image", req.Image.Slug).
		WithField("size", req.Size).
		WithField("name", req.Name)

	logger.Debug("instance create")

	client := newClient(ctx, args.Token)
	droplet, _, err := client.Droplets.Create(ctx, req)
	if err != nil {
		logger.WithError(err).Error("cannot create instance")
		return res, err
	}

	// record the droplet ID
	res.ID = droplet.ID

	logger.WithField("name", req.Name).
		Info("instance created")

	// poll the digitalocean endpoint for server updates
	// and exit when a network address is allocated.
	interval := time.Duration(0)
poller:
	for {
		select {
		case <-ctx.Done():
			logger.WithField("name", req.Name).
				Debug("cannot ascertain network")

			return res, ctx.Err()
		case <-time.After(interval):
			interval = time.Second * 30

			logger.WithField("name", req.Name).
				Debug("find instance network")

			droplet, _, err = client.Droplets.Get(ctx, res.ID)
			if err != nil {
				logger.WithError(err).
					Error("cannot find instance")
				return res, err
			}

			for _, network := range droplet.Networks.V4 {
				if network.Type == "public" {
					res.IP = network.IPAddress
				}
			}

			if res.IP != "" {
				break poller
			}
		}
	}

	logger.WithField("name", req.Name).
		WithField("ip", res.IP).
		WithField("id", res.ID).
		Debug("instance network ready")

	return res, nil
}

// Destroy destroys the server instance.
func Destroy(ctx context.Context, args DestroyArgs) error {
	client := newClient(ctx, args.Token)
	_, err := client.Droplets.Delete(ctx, args.ID)
	if err != nil {
		logger.FromContext(ctx).
			WithError(err).
			WithField("id", args.ID).
			WithField("ip", args.IP).
			Error("cannot terminate server")
	}
	return err
}

// RegisterKey registers the ssh public key with the account if
// it is not already registered.
func RegisterKey(ctx context.Context, args RegisterArgs) error {
	client := newClient(ctx, args.Token)
	_, _, err := client.Keys.GetByFingerprint(ctx, args.Fingerprint)
	if err == nil {
		return nil
	}

	// if the ssh key does not exists we attempt to register
	// with the digital ocean account.
	_, _, err = client.Keys.Create(ctx, &godo.KeyCreateRequest{
		Name:      args.Name,
		PublicKey: args.Data,
	})
	return err
}

// helper function returns a new docker client.
func newClient(ctx context.Context, token string) *godo.Client {
	return godo.NewClient(
		oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
			},
		)),
	)
}
