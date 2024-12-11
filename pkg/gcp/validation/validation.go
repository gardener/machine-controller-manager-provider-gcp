// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
)

// ValidateProviderSpec validates gcp provider spec
func ValidateProviderSpec(spec *api.GCPProviderSpec) []error {
	fldPath := field.NewPath("spec")
	var allErrs []error

	allErrs = append(allErrs, validateGCPDisks(spec.Disks, fldPath.Child("disks"))...)

	if "" == spec.MachineType {
		allErrs = append(allErrs, field.Required(fldPath.Child("machineType"), "machineType is required"))
	}
	if "" == spec.Region {
		allErrs = append(allErrs, field.Required(fldPath.Child("region"), "region is required"))
	}
	if "" == spec.Zone {
		allErrs = append(allErrs, field.Required(fldPath.Child("zone"), "zone is required"))
	}

	allErrs = append(allErrs, validateGCPNetworkInterfaces(spec.NetworkInterfaces, fldPath.Child("networkInterfaces"))...)
	allErrs = append(allErrs, validateGCPMetadata(spec.Metadata, fldPath.Child("networkInterfaces"))...)
	allErrs = append(allErrs, validateGCPGpu(spec.Gpu, fldPath.Child("gpu"))...)
	allErrs = append(allErrs, validateGCPScheduling(spec.Scheduling, spec.Gpu, fldPath.Child("scheduling"))...)

	return allErrs
}

// ValidateZone validates the zone in the providerSpec
func ValidateZone(zone string) error {
	if zone == "" {
		return fmt.Errorf("zone cannot be empty")
	}
	return nil
}

// ValidateSecret validates the machine class secret
func ValidateSecret(secret *corev1.Secret) []error {
	var allErrs []error

	if secret == nil {
		allErrs = append(allErrs, fmt.Errorf("secret object that has been passed by the MCM is nil"))
	} else {
		_, serviceAccountJSONExists := secret.Data[api.GCPServiceAccountJSON]
		_, serviceAccountJSONAlternativeExists := secret.Data[api.GCPAlternativeServiceAccountJSON]
		_, credentialsConfigExists := secret.Data[api.GCPCredentialsConfig]
		_, userDataExists := secret.Data["userData"]

		if !serviceAccountJSONExists && !serviceAccountJSONAlternativeExists && !credentialsConfigExists {
			allErrs = append(allErrs, fmt.Errorf("secret %s, %s or %s is required field", api.GCPServiceAccountJSON, api.GCPAlternativeServiceAccountJSON, api.GCPCredentialsConfig))
		}
		if !userDataExists {
			allErrs = append(allErrs, fmt.Errorf("secret userData is required field"))
		}
	}

	return allErrs
}

func validateGCPDisks(disks []*api.GCPDisk, fldPath *field.Path) []error {
	var allErrs []error

	if 0 == len(disks) {
		allErrs = append(allErrs, field.Required(fldPath, "at least one disk is required"))
	}

	for i, disk := range disks {
		idxPath := fldPath.Index(i)
		if disk.Type == api.GCPDiskTypeScratch && (disk.Interface != api.GCPDiskInterfaceNVME && disk.Interface != api.GCPDiskInterfaceSCSI) {
			allErrs = append(allErrs, field.NotSupported(idxPath.Child("interface"), disk.Interface, []string{api.GCPDiskInterfaceNVME, api.GCPDiskInterfaceSCSI}))
		}
		if disk.Boot && "" == disk.Image {
			allErrs = append(allErrs, field.Required(idxPath.Child("image"), "image is required for boot disk"))
		}
		if disk.Encryption != nil {
			var kmsKeyName = strings.TrimSpace(disk.Encryption.KmsKeyName)
			var kmsKeyServiceAccount = strings.TrimSpace(disk.Encryption.KmsKeyServiceAccount)
			if kmsKeyName == "" {
				allErrs = append(allErrs, field.Required(idxPath.Child("kmsKeyName"), "kmsKeyName is required to be specified"))
			}
			// to deal with situation where  just spaces have been specified for `kmsKeyServiceAccount`
			if disk.Encryption.KmsKeyServiceAccount != "" && kmsKeyServiceAccount == "" {
				allErrs = append(allErrs, field.Required(idxPath.Child("kmsKeyServiceAccount"), "kmsKeyServiceAccount should either be explicitly specified without spaces or left un-specified to default to the Compute Service Agent"))
			}
		}
	}

	return allErrs
}

func validateGCPNetworkInterfaces(interfaces []*api.GCPNetworkInterface, fldPath *field.Path) []error {
	var allErrs []error

	if 0 == len(interfaces) {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkInterfaces"), "at least one network interface is required"))
	}

	for i, nic := range interfaces {
		idxPath := fldPath.Index(i)

		// Validate network and subnetwork
		if "" == nic.Network && "" == nic.Subnetwork {
			allErrs = append(allErrs, field.Required(idxPath, "either network or subnetwork or both is required"))
		}

		// Validate StackType
		if nic.StackType != "" && (nic.StackType != "IPV4_IPV6" && nic.StackType != "IPV4_ONLY") {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("stackType"), nic.Ipv6AccessType, "StackType must be either IPV4_IPV6 or IPV4_ONLY"))
		}

		// Validate IPv6 Access Type
		if nic.Ipv6AccessType != "" && (nic.Ipv6AccessType != "EXTERNAL" && nic.Ipv6AccessType != "INTERNAL") {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("ipv6AccessType"), nic.Ipv6AccessType, "IPv6 AccessType must be either EXTERNAL or INTERNAL"))
		}

		// Validate IP CIDR RANGE
		if nic.IpCidrRange != "" {
			if err := validateIpCidrRange(nic.IpCidrRange); err != nil {
				allErrs = append(allErrs, field.Invalid(idxPath.Child("ipCidrRange"), nic.IpCidrRange, err.Error()))
			}
		}
	}

	return allErrs
}

// Function to validate CIDR range (e.g., "/xx" where xx is between 0 and 32)
func validateIpCidrRange(cidrRange string) error {
	// Check if the string starts with "/"
	if !strings.HasPrefix(cidrRange, "/") {
		return fmt.Errorf("CIDR range must start with '/'")
	}

	// Extract the mask size
	maskStr := cidrRange[1:]
	maskSize, err := strconv.Atoi(maskStr)
	if err != nil {
		return fmt.Errorf("invalid CIDR mask size: %s", maskStr)
	}

	// Validate mask size (0 to 32 for IPv4)
	if maskSize < 0 || maskSize > 32 {
		return fmt.Errorf("CIDR mask size must be between 0 and 32")
	}

	return nil
}

func validateGCPMetadata(metadata []*api.GCPMetadata, fldPath *field.Path) []error {
	var allErrs []error

	for i, item := range metadata {
		idxPath := fldPath.Index(i)
		if item.Key == "user-data" {
			allErrs = append(allErrs, field.Forbidden(idxPath.Child("key"), "user-data key is forbidden in metadata"))
		}
	}

	return allErrs
}

func validateGCPScheduling(scheduling api.GCPScheduling, gpu *api.GCPGpu, fldPath *field.Path) []error {
	var allErrs []error

	if "MIGRATE" != scheduling.OnHostMaintenance && "TERMINATE" != scheduling.OnHostMaintenance {
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("onHostMaintenance"), scheduling.OnHostMaintenance, []string{"MIGRATE", "TERMINATE"}))
	}

	if gpu != nil && scheduling.OnHostMaintenance != "TERMINATE" {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("onHostMaintenance"), "liveMigration is not allowed for VMs with gpu attached, use \"TERMINATE\" instead"))
	}

	return allErrs
}

func validateGCPGpu(gpu *api.GCPGpu, fldPath *field.Path) []error {
	var allErrs []error

	if gpu != nil {
		if gpu.AcceleratorType == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("acceleratorType"), "accelerator type needs to be provided"))
		}

		if gpu.Count == 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("count"), "gpu count must be > 0"))
		}
	}

	return allErrs
}
