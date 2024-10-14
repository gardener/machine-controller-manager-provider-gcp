/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This file was copied and modified from the kubernetes-csi/drivers project
https://github.com/kubernetes-csi/drivers/blob/release-1.0/pkg/nfs/plugin.go

Modifications Copyright SAP SE or an SAP affiliate company and Gardener contributors
*/

package gcp

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
)

// PluginSPI provides an interface to deal with cloud provider session
// You can optionally enhance this interface to add interface methods here
// You can use it to mock cloud provider calls
type PluginSPI interface {
	NewComputeService(secrets *corev1.Secret) (context.Context, *compute.Service, error)
}

// MachinePlugin implements the driver.Driver
// It also implements the PluginSPI interface
type MachinePlugin struct {
	SPI PluginSPI
}

// PluginSPIImpl is the real implementation of PluginSPI interface
// that makes the calls to the provider SDK
type PluginSPIImpl struct{}

// NewComputeService returns an instance of the compute service
func (spi *PluginSPIImpl) NewComputeService(secret *corev1.Secret) (context.Context, *compute.Service, error) {
	ctx := context.Background()
	credentialsConfigJSON := extractCredentialsFromData(secret.Data, api.GCPServiceAccountJSON, api.GCPAlternativeServiceAccountJSON, api.GCPCredentialsConfig)

	creds, err := google.CredentialsFromJSONWithParams(ctx, []byte(credentialsConfigJSON), google.CredentialsParams{
		Scopes: []string{compute.CloudPlatformScope},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse serviceAccountJSON secret value: %w", err)
	}

	clientOption := option.WithTokenSource(creds.TokenSource)
	computeService, err := compute.NewService(ctx, clientOption)
	if err != nil {
		return nil, nil, err
	}
	return ctx, computeService, nil

}

// NewGCPPlugin returns a new Gcp plugin
func NewGCPPlugin(pluginSPI PluginSPI) *MachinePlugin {
	return &MachinePlugin{
		SPI: pluginSPI,
	}
}

// extractCredentialsFromData extracts and trims a value from the given data map. The first key that exists is being
// returned, otherwise, the next key is tried, etc. If no key exists then an empty string is returned.
func extractCredentialsFromData(data map[string][]byte, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			return strings.TrimSpace(string(val))
		}
	}
	return ""
}
