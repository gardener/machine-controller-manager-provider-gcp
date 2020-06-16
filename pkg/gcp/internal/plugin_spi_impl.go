/*
 * Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"context"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/apis"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	option "google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
)

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
