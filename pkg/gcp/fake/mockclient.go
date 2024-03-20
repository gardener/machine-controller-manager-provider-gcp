// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"fmt"
	"net/http"

	compute "google.golang.org/api/compute/v1"
	option "google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
)

// PluginSPIImpl is the mock implementation of PluginSPIImpl
type PluginSPIImpl struct {
	Client *http.Client
}

// NewComputeService creates a compute service instance using the mock
func (ms *PluginSPIImpl) NewComputeService(secrets *corev1.Secret) (context.Context, *compute.Service, error) {
	ctx := context.Background()

	_, serviceAccountJSON := secrets.Data[api.GCPServiceAccountJSON]
	_, serviceAccountJSONAlternative := secrets.Data[api.GCPAlternativeServiceAccountJSON]
	if !serviceAccountJSON && !serviceAccountJSONAlternative {
		return nil, nil, fmt.Errorf("Missing secrets to connect to compute service")
	}

	// create a compute service using a mockclient work
	client := option.WithHTTPClient(ms.Client)
	endpoint := option.WithEndpoint("http://127.0.0.1:6666")

	computeService, err := compute.NewService(ctx, client, endpoint)
	if err != nil {
		return nil, nil, err
	}

	return ctx, computeService, nil
}
