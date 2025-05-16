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
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"

	"github.com/gardener/gardener-extension-provider-gcp/pkg/gcp"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
)

var (
	allowedTokenURL                             = "https://sts.googleapis.com/v1/token" // #nosec G101 (CWE-798) -- Constant value, not subject to change
	allowedServiceAccountImpersonationURLRegExp = regexp.MustCompile(`^https://iamcredentials\.googleapis\.com/v1/projects/-/serviceAccounts/.+:generateAccessToken$`)
	allowedSubjectTokenType                     = "urn:ietf:params:oauth:token-type:jwt"                    // #nosec G101 (CWE-798) -- Constant value, not subject to change
	allowedCredSourceFilePath                   = "/var/run/secrets/gardener.cloud/workload-identity/token" // #nosec G101 (CWE-798) -- Constant value, not subject to change
)

var serviceAccountAllowedFields = map[string]struct{}{
	"type":                        {},
	"project_id":                  {},
	"client_email":                {},
	"universe_domain":             {},
	"auth_uri":                    {},
	"auth_provider_x509_cert_url": {},
	"client_x509_cert_url":        {},
	"client_id":                   {},
	"private_key_id":              {},
	"private_key":                 {},
	"token_uri":                   {},
}

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
	credentialsConfigJSON, credentialKey := extractCredentialsFromData(secret.Data, api.GCPServiceAccountJSON, api.GCPAlternativeServiceAccountJSON, api.GCPCredentialsConfig)

	sa, err := gcp.GetCredentialsConfigFromJSON([]byte(credentialsConfigJSON))
	if err != nil {
		return ctx, nil, fmt.Errorf("could not get service account from %q field: %w", credentialKey, err)
	}

	if sa.Type == gcp.ServiceAccountCredentialType {

		fields := map[string]string{}
		if err := json.Unmarshal([]byte(credentialsConfigJSON), &fields); err != nil {
			return ctx, nil, fmt.Errorf("failed to unmarshal '%q' field: %w", credentialKey, err)
		}

		for f := range fields {
			if _, ok := serviceAccountAllowedFields[f]; !ok {
				return ctx, nil, fmt.Errorf("forbidden fields are present. Allowed fields are %s", strings.Join(slices.Collect(maps.Keys(serviceAccountAllowedFields)), ", "))
			}
		}
		jwt, err := google.JWTConfigFromJSON([]byte(credentialsConfigJSON), compute.CloudPlatformScope)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot parse serviceAccountJSON secret value: %w", err)
		}
		clientOption := option.WithTokenSource(jwt.TokenSource(ctx))
		computeService, err := compute.NewService(ctx, clientOption)
		if err != nil {
			return nil, nil, err
		}
		return ctx, computeService, nil

	} else if sa.Type == gcp.ExternalAccountCredentialType {
		err := validateExtAccountFields(sa)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid secret. Err: %w", err)
		}
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

	} else {
		return ctx, nil, fmt.Errorf("forbidden credential type %q used. Only %q or %q is allowed", sa.Type, gcp.ServiceAccountCredentialType, gcp.ExternalAccountCredentialType)
	}
}

func validateExtAccountFields(sa *gcp.CredentialsConfig) error {
	if strings.TrimSpace(sa.TokenURL) != allowedTokenURL {
		return fmt.Errorf("invalid token URL found")
	}

	if !allowedServiceAccountImpersonationURLRegExp.MatchString(sa.ServiceAccountImpersonationURL) {
		return fmt.Errorf("invalid service_account_impersonation_url found in secret")
	}

	if strings.TrimSpace(sa.SubjectTokenType) != allowedSubjectTokenType {
		return fmt.Errorf("invalid subject_token_type found in secret")
	}

	if strings.TrimSpace(sa.TokenFilePath) != allowedCredSourceFilePath {
		return fmt.Errorf("invalid credential_source file path present")
	}

	return nil
}

// NewGCPPlugin returns a new Gcp plugin
func NewGCPPlugin(pluginSPI PluginSPI) *MachinePlugin {
	return &MachinePlugin{
		SPI: pluginSPI,
	}
}

// extractCredentialsFromData extracts and trims a value from the given data map. The first key that exists is being
// returned, otherwise, the next key is tried, etc. If no key exists then an empty string is returned.
func extractCredentialsFromData(data map[string][]byte, keys ...string) (string, string) {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			return strings.TrimSpace(string(val)), key
		}
	}
	return "", ""
}
