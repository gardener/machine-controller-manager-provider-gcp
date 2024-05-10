// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

const (
	// GCPServiceAccountJSON is a constant for a key name that is part of the GCP cloud credentials.
	GCPServiceAccountJSON = "serviceAccountJSON"
	// GCPAlternativeServiceAccountJSON is a constant for a key name of a secret containing the GCP credentials (service
	// account json).
	GCPAlternativeServiceAccountJSON = "serviceaccount.json"
)

// +genclient

// GCPProviderSpec contains the fields of
// provider spec that the plugin expects
type GCPProviderSpec struct {
	// APIVersion refers to the APIVersion of the object
	APIVersion string

	// CanIpForward: Allows this instance to send and receive packets with
	// non-matching destination or source IPs. This is required if you plan
	// to use this instance to forward routes. For more information, see
	// Enabling IP Forwarding.
	CanIPForward bool `json:"canIpForward"`

	// DeletionProtection: Whether the resource should be protected against
	// deletion.
	DeletionProtection bool `json:"deletionProtection"`

	// Description: An optional description of this resource. Provide this
	// property when you create the resource.
	Description *string `json:"description,omitempty"`

	// Disks: Array of disks associated with this instance. Persistent disks
	// must be created before you can assign them.
	Disks []*GCPDisk `json:"disks,omitempty"`

	// Gpu: Configurations related to GPU which would be attached to the instance. Enough
	// Quota of the particular GPU should be available.
	Gpu *GCPGpu `json:"gpu,omitempty"`

	// Labels: Labels to apply to this instance.
	Labels map[string]string `json:"labels,omitempty"`

	// MachineType: Full or partial URL of the machine type resource to use
	// for this instance, in the format:
	// zones/zone/machineTypes/machine-type. This is provided by the client
	// when the instance is created. For example, the following is a valid
	// partial url to a predefined machine
	// type:
	// zones/us-central1-f/machineTypes/n1-standard-1
	//
	//
	// To create a custom machine type, provide a URL to a machine type in
	// the following format, where CPUS is 1 or an even number up to 32 (2,
	// 4, 6, ... 24, etc), and MEMORY is the total memory for this instance.
	// Memory must be a multiple of 256 MB and must be supplied in MB (e.g.
	// 5 GB of memory is 5120
	// MB):
	// zones/zone/machineTypes/custom-CPUS-MEMORY
	//
	//
	// For example: zones/us-central1-f/machineTypes/custom-4-5120
	//
	// For a full list of restrictions, read the Specifications for custom
	// machine types.
	MachineType string `json:"machineType"`

	// Metadata: The metadata key/value pairs assigned to this instance.
	// This includes custom metadata and predefined keys.
	Metadata []*GCPMetadata `json:"metadata,omitempty"`

	// MinCpuPlatform: The name of the minimum CPU platform that is requested
	// for this instance.
	MinCPUPlatform string `json:"minCpuPlatform,omitempty"`

	// NetworkInterfaces: An array of network configurations for this
	// instance. These specify how interfaces are configured to interact
	// with other network services, such as connecting to the internet.
	// Multiple interfaces are supported per instance.
	NetworkInterfaces []*GCPNetworkInterface `json:"networkInterfaces,omitempty"`

	// Region: in which instance is to be deployed
	Region string `json:"region"`

	// Scheduling: Sets the scheduling options for this instance.
	Scheduling GCPScheduling `json:"scheduling"`

	// ServiceAccounts: A list of service accounts, with their specified
	// scopes, authorized for this instance. Only one service account per VM
	// instance is supported.
	//
	// Service accounts generate access tokens that can be accessed through
	// the metadata server and used to authenticate applications on the
	// instance. See Service Accounts for more information.
	ServiceAccounts []GCPServiceAccount `json:"serviceAccounts"`

	// Tags: to be placed on the VM
	// +optional
	Tags []string `json:"tags,omitempty"`

	// Zone: in which instance is to be deployed
	Zone string `json:"zone"`
}

// GCPDisk describes disks for GCP.
type GCPDisk struct {
	// AutoDelete: Specifies whether the disk will be auto-deleted when the
	// instance is deleted (but not when the disk is detached from the
	// instance).
	AutoDelete *bool `json:"autoDelete"`

	// Boot: Indicates that this is a boot disk. The virtual machine will
	// use the first partition of the disk for its root filesystem.
	Boot bool `json:"boot"`

	// SizeGb: Specifies the size of the disk in base-2 GB.
	SizeGb int64 `json:"sizeGb"`

	// Type: Specifies the disk type to use to create the instance. If
	// not specified, the default is pd-standard, specified using the full
	// URL. For
	// example:
	// https://www.googleapis.com/compute/v1/projects/project/zones/
	// zone/diskTypes/pd-standard
	//
	//
	// Other values include pd-ssd and local-ssd. If you define this field,
	// you can provide either the full or partial URL. For example, the
	// following are valid values:
	// - https://www.googleapis.com/compute/v1/projects/project/zones/zone/diskTypes/diskType
	// - projects/project/zones/zone/diskTypes/diskType
	// - zones/zone/diskTypes/diskType  Note that for InstanceTemplate, this
	// is the name of the disk type, not URL.
	// If you use "SCRATCH" as the value, it is defaulted to local-ssd
	Type string `json:"type"`

	// Interface: Specifies the disk interface to use for attaching this
	// disk, which is either SCSI or NVME. The default is SCSI. Persistent
	// disks must always use SCSI and the request will fail if you attempt
	// to attach a persistent disk in any other format than SCSI. Local SSDs
	// can use either NVME or SCSI. For performance characteristics of SCSI
	// over NVMe, see Local SSD performance.
	//
	// Possible values:
	//   "NVME"
	//   "SCSI"
	// This is only applied when the disk type is "SCRATCH" currently
	Interface string `json:"interface"`

	// Image: The source image to create this disk. When creating a
	// new instance, one of initializeParams.sourceImage or disks.source is
	// required except for local SSD.
	//
	// To create a disk with one of the public operating system images,
	// specify the image by its family name. For example, specify
	// family/debian-9 to use the latest Debian 9
	// image:
	// projects/debian-cloud/global/images/family/debian-9
	//
	//
	// Alternatively, use a specific version of a public operating system
	// image:
	// projects/debian-cloud/global/images/debian-9-stretch-vYYYYMMDD
	//
	//
	//
	// To create a disk with a custom image that you created, specify the
	// image name in the following
	// format:
	// global/images/my-custom-image
	//
	//
	// You can also specify a custom image by its image family, which
	// returns the latest version of the image in that family. Replace the
	// image name with
	// family/family-name:
	// global/images/family/my-image-family
	//
	//
	// If the source image is deleted later, this field will not be set.
	Image string `json:"image"`

	// Encryption: Encryption details for this disk
	Encryption *GCPDiskEncryption `json:"encryption"`

	// Labels: Labels to apply to this disk. These can be later modified by
	// the disks.setLabels method. This field is only applicable for
	// persistent disks.
	Labels map[string]string `json:"labels"`

	// ProvisionedIops of disk to create.
	// Only for use with disks of type pd-extreme and hyperdisk-extreme.
	// The IOPS must be specified within defined limits
	ProvisionedIops int64 `json:"provisionedIops"`

	// ProvisionedThroughput of disk to create.
	// Only for hyperdisk-balanced or hyperdisk-throughput volumes,
	// measured in MiB per second, that the disk can handle.
	// The throughput must be specified within defined limits
	ProvisionedThroughput int64 `json:"provisionedThroughput"`
}

// GCPDiskEncryption holds references to encryption data
type GCPDiskEncryption struct {
	// KmsKeyName: key name of the cloud kms disk encryption key. Not optional
	KmsKeyName string `json:"kmsKeyName"`

	// KmsKeyServiceAccount: The service account granted the `roles/cloudkms.cryptoKeyEncrypterDecrypter` for the key name.
	// If empty, then the role should be given to the Compute Engine Service Agent Account. This usually has the format
	// service-PROJECT_NUMBER@compute-system.iam.gserviceaccount.com. See: https://cloud.google.com/iam/docs/service-agents#compute-engine-service-agent
	// One can add IAM roles using the gcloud CLI:
	//  gcloud projects add-iam-policy-binding projectId --member
	//	serviceAccount:name@projectIdgserviceaccount.com --role roles/cloudkms.cryptoKeyEncrypterDecrypter
	KmsKeyServiceAccount string `json:"kmsKeyServiceAccount"`
}

// GCPMetadata describes metadata for GCP.
type GCPMetadata struct {
	// Key: Key for the metadata entry. Keys must conform to the following
	// regexp: [a-zA-Z0-9-_]+, and be less than 128 bytes in length. This is
	// reflected as part of a URL in the metadata server. Additionally, to
	// avoid ambiguity, keys must not conflict with any other metadata keys
	// for the project.
	Key string `json:"key"`

	// Value: Value for the metadata entry. These are free-form strings, and
	// only have meaning as interpreted by the image running in the
	// instance. The only restriction placed on values is that their size
	// must be less than or equal to 262144 bytes (256 KiB).
	Value *string `json:"value"`
}

// GCPNetworkInterface describes network interfaces for GCP
type GCPNetworkInterface struct {
	// DisableExternalIP: is false, implies Attach an external IP to VM
	DisableExternalIP bool `json:"disableExternalIP,omitempty"`

	// Network: URL of the network resource for this instance. When creating
	// an instance, if neither the network nor the subnetwork is specified,
	// the default network global/networks/default is used; if the network
	// is not specified but the subnetwork is specified, the network is
	// inferred.
	//
	// This field is optional when creating a firewall rule. If not
	// specified when creating a firewall rule, the default network
	// global/networks/default is used.
	//
	// If you specify this property, you can specify the network as a full
	// or partial URL. For example, the following are all valid URLs:
	// - https://www.googleapis.com/compute/v1/projects/project/global/networks/network
	// - projects/project/global/networks/network
	// - global/networks/default
	Network string `json:"network,omitempty"`

	// Subnetwork: The URL of the Subnetwork resource for this instance. If
	// the network resource is in legacy mode, do not provide this property.
	// If the network is in auto subnet mode, providing the subnetwork is
	// optional. If the network is in custom subnet mode, then this field
	// should be specified. If you specify this property, you can specify
	// the subnetwork as a full or partial URL. For example, the following
	// are all valid URLs:
	// - https://www.googleapis.com/compute/v1/projects/project/regions/region/subnetworks/subnetwork
	// - regions/region/subnetworks/subnetwork
	Subnetwork string `json:"subnetwork,omitempty"`
}

// GCPScheduling describes scheduling configuration for GCP.
type GCPScheduling struct {
	// AutomaticRestart: Specifies whether the instance should be
	// automatically restarted if it is terminated by Compute Engine (not
	// terminated by a user). You can only set the automatic restart option
	// for standard instances. Preemptible instances cannot be automatically
	// restarted.
	//
	// By default, this is set to true so an instance is automatically
	// restarted if it is terminated by Compute Engine.
	AutomaticRestart bool `json:"automaticRestart"`

	// OnHostMaintenance: Defines the maintenance behavior for this
	// instance. For standard instances, the default behavior is MIGRATE.
	// For preemptible instances, the default and only possible behavior is
	// TERMINATE. For more information, see Setting Instance Scheduling
	// Options.
	//
	// Possible values:
	//   "MIGRATE"
	//   "TERMINATE"
	OnHostMaintenance string `json:"onHostMaintenance"`

	// Preemptible: Defines whether the instance is preemptible. This can
	// only be set during instance creation, it cannot be set or changed
	// after the instance has been created.
	Preemptible bool `json:"preemptible"`
}

// GCPServiceAccount describes service accounts for GCP.
type GCPServiceAccount struct {
	// Email: Email address of the service account.
	Email string `json:"email"`

	// Scopes: The list of scopes to be made available for this service
	// account.
	Scopes []string `json:"scopes"`
}

// GCPGpu describes gpu configurations for GCP
type GCPGpu struct {
	AcceleratorType string `json:"acceleratorType"`
	Count           int64  `json:"count"`
}
