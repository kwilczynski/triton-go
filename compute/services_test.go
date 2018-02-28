//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package compute_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/testutils"
	"github.com/pkg/errors"
)

var listServicesErrorType = errors.New("unable to list services")

func TestAccServicesList(t *testing.T) {
	const stateKey = "services"

	testutils.AccTest(t, testutils.TestCase{
		Steps: []testutils.Step{

			&testutils.StepClient{
				StateBagKey: stateKey,
				CallFunc: func(config *triton.ClientConfig) (interface{}, error) {
					return compute.NewClient(config)
				},
			},

			&testutils.StepAPICall{
				StateBagKey: stateKey,
				CallFunc: func(client interface{}) (interface{}, error) {
					c := client.(*compute.ComputeClient)
					ctx := context.Background()
					input := &compute.ListServicesInput{}
					return c.Services().List(ctx, input)
				},
			},

			&testutils.StepAssertFunc{
				AssertFunc: func(state testutils.TritonStateBag) error {
					services, ok := state.GetOk(stateKey)
					if !ok {
						return fmt.Errorf("state key %q not found", stateKey)
					}

					toFind := []string{"docker"}
					for _, serviceName := range toFind {
						found := false
						for _, service := range services.([]*compute.Service) {
							if service.Name == serviceName {
								found = true
								if service.Endpoint == "" {
									return fmt.Errorf("%q has no endpoint", service.Name)
								}
							}
						}
						if !found {
							return fmt.Errorf("could not find service %q", serviceName)
						}
					}

					return nil
				},
			},
		},
	})
}

func TestListServices(t *testing.T) {
	computeClient := MockComputeClient()

	do := func(ctx context.Context, cc *compute.ComputeClient) ([]*compute.Service, error) {
		defer testutils.DeactivateClient()

		services, err := cc.Services().List(ctx, &compute.ListServicesInput{})
		if err != nil {
			return nil, err
		}
		return services, nil
	}

	t.Run("successful", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "services"), listServicesSuccess)

		resp, err := do(context.Background(), computeClient)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if resp == nil {
			t.Fatal("expected response, got nil")
		}
	})

	t.Run("eof", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "services"), listServicesEmpty)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error to contain EOF, got %v", err)
		}
	})

	t.Run("bad_decode", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "services"), listServicesBadDecode)

		_, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Errorf("expected decode to fail, got %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		testutils.RegisterResponder("GET", path.Join("/", accountURL, "services"), listServicesError)

		resp, err := do(context.Background(), computeClient)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		if resp != nil {
			t.Errorf("expected no response, got %v", resp)
		}

		if !strings.Contains(err.Error(), "unable to list services") {
			t.Errorf("expected error to equal test error, got %v", err)
		}
	})
}

func listServicesSuccess(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`{
  "cloudapi": "https://us-west-1.api.example.com",
  "docker": "tcp://us-west-1.docker.example.com",
  "manta": "https://us-west.manta.example.com"
}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listServicesEmpty(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       http.NoBody,
	}, nil
}

func listServicesBadDecode(req *http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Content-Type", "application/json")

	body := strings.NewReader(`[{
  "cloudapi": "https://us-west-1.api.example.com",
  "docker": "tcp://us-west-1.docker.example.com",
  "manta": "https://us-west.manta.example.com",
}]`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     header,
		Body:       ioutil.NopCloser(body),
	}, nil
}

func listServicesError(req *http.Request) (*http.Response, error) {
	return nil, listServicesErrorType
}
