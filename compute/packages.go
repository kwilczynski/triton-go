//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package compute

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/joyent/triton-go/client"
	"github.com/pkg/errors"
)

type PackagesClient struct {
	client *client.Client
}

type Package struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Memory      int64  `json:"memory"`
	Disk        int64  `json:"disk"`
	Swap        int64  `json:"swap"`
	LWPs        int64  `json:"lwps"`
	VCPUs       int64  `json:"vcpus"`
	Version     string `json:"version"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
}

type ListPackagesInput struct {
	Name    string `json:"name"`
	Memory  int64  `json:"memory"`
	Disk    int64  `json:"disk"`
	Swap    int64  `json:"swap"`
	LWPs    int64  `json:"lwps"`
	VCPUs   int64  `json:"vcpus"`
	Version string `json:"version"`
	Group   string `json:"group"`
}

func (c *PackagesClient) List(ctx context.Context, input *ListPackagesInput) ([]*Package, error) {
	fullPath := path.Join("/", c.client.AccountName, "packages")

	query := &url.Values{}
	if input.Name != "" {
		query.Set("name", input.Name)
	}
	if input.Memory != 0 {
		query.Set("memory", strconv.Itoa(int(input.Memory)))
	}
	if input.Disk != 0 {
		query.Set("disk", strconv.Itoa(int(input.Disk)))
	}
	if input.Swap != 0 {
		query.Set("swap", strconv.Itoa(int(input.Swap)))
	}
	if input.LWPs != 0 {
		query.Set("lwps", strconv.Itoa(int(input.LWPs)))
	}
	if input.VCPUs != 0 {
		query.Set("vcpus", strconv.Itoa(int(input.VCPUs)))
	}
	if input.Version != "" {
		query.Set("version", input.Version)
	}
	if input.Group != "" {
		query.Set("group", input.Group)
	}

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
		Query:  query,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list packages")
	}
	defer respReader.Close()

	var result []*Package
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list packages response")
	}

	return result, nil
}

type GetPackageInput struct {
	ID string
}

func (c *PackagesClient) Get(ctx context.Context, input *GetPackageInput) (*Package, error) {
	fullPath := path.Join("/", c.client.AccountName, "packages", input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get package")
	}
	defer respReader.Close()

	var result *Package
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get package response")
	}

	return result, nil
}
