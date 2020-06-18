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

Modifications Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved.
*/

package gcp

import (
	"context"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/apis"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
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
func (spi *PluginSPIImpl) NewComputeService(secrets *corev1.Secret) (context.Context, *compute.Service, error) {
	ctx := context.Background()
	jwt, err := google.JWTConfigFromJSON((secrets.Data[api.GCPServiceAccountJSON]), compute.CloudPlatformScope)
	if err != nil {
		return nil, nil, err
	}

	clientOption := option.WithTokenSource(jwt.TokenSource(ctx))
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
