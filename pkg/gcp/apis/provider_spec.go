/*
Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

const (
	// GCPServiceAccountJSON is a constant for a key name that is part of the GCP cloud credentials.
	GCPServiceAccountJSON string = "serviceAccountJSON"
)

// GCPProviderSpec contains the fields of
// provider spec that the plugin expects
type GCPProviderSpec struct {
	CanIPForward       bool                   `json:"canIpForward"`
	DeletionProtection bool                   `json:"deletionProtection"`
	Description        *string                `json:"description,omitempty"`
	Disks              []*GCPDisk             `json:"disks,omitempty"`
	Labels             map[string]string      `json:"labels,omitempty"`
	MachineType        string                 `json:"machineType"`
	Metadata           []*GCPMetadata         `json:"metadata,omitempty"`
	NetworkInterfaces  []*GCPNetworkInterface `json:"networkInterfaces,omitempty"`
	Scheduling         GCPScheduling          `json:"scheduling"`
	ServiceAccounts    []GCPServiceAccount    `json:"serviceAccounts"`
	Region             string                 `json:"region"`
	Zone               string                 `json:"zone"`

	// SSHKeys is an optional array of ssh public keys to deploy to VM (may already be included in UserData)
	// +optional
	SSHKeys []string `json:"sshKeys,omitempty"`
	// Tags to be placed on the VM
	// +optional
	Tags []string `json:"tags,omitempty"`
	// TODO: Add the raw extension struct expected while recieving machine operating requests
	// Some dummy examples are mentioned below
	// below ones are the ones created by the webide
	// MachineImageName contains the image name from which machine is to be spawned
	//MachineImageName string
}

// GCPDisk describes disks for GCP.
type GCPDisk struct {
	AutoDelete *bool             `json:"autoDelete"`
	Boot       bool              `json:"boot"`
	SizeGb     int64             `json:"sizeGb"`
	Type       string            `json:"type"`
	Interface  string            `json:"interface"`
	Image      string            `json:"image"`
	Labels     map[string]string `json:"labels"`
}

// GCPMetadata describes metadata for GCP.
type GCPMetadata struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
}

// GCPNetworkInterface describes network interfaces for GCP
type GCPNetworkInterface struct {
	DisableExternalIP bool   `json:"disableExternalIP,omitempty"`
	Network           string `json:"network,omitempty"`
	Subnetwork        string `json:"subnetwork,omitempty"`
}

// GCPScheduling describes scheduling configuration for GCP.
type GCPScheduling struct {
	AutomaticRestart  bool   `json:"automaticRestart"`
	OnHostMaintenance string `json:"onHostMaintenance"`
	Preemptible       bool   `json:"preemptible"`
}

// GCPServiceAccount describes service accounts for GCP.
type GCPServiceAccount struct {
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}
