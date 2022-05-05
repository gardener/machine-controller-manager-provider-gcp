/*
 * Copyright 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
 *
 */

package validation

import (
	"fmt"
	"regexp"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// DiskTypeStandard is the standard disk type
	DiskTypeStandard = "pd-standard"
	// DiskTypeSSD is the SSD disk type
	DiskTypeSSD = "pd-ssd"
	// DiskTypeScratch is the SCRATCH disk type
	DiskTypeScratch = "SCRATCH"
	// DiskTypePDBalanced is the balanced disk type
	DiskTypePDBalanced = "pd-balanced"

	// DiskInterfaceNVME is the NVME disk interface
	DiskInterfaceNVME = "NVME"
	// DiskInterfaceSCSI is the SCSI disk interface
	DiskInterfaceSCSI = "SCSI"
)

// ValidateGCPProviderSpec validates gcp provider spec
func ValidateGCPProviderSpec(spec *api.GCPProviderSpec, secrets *corev1.Secret) []error {
	allErrs := validateGCPMachineClassSpec(spec, field.NewPath("spec"))
	allErrs = append(allErrs, validateSecrets(secrets)...)
	return allErrs
}

func validateSecrets(secret *corev1.Secret) []error {
	var allErrs []error

	if secret == nil {
		allErrs = append(allErrs, fmt.Errorf("secret object that has been passed by the MCM is nil"))
	} else {
		_, serviceAccountJSONExists := secret.Data[api.GCPServiceAccountJSON]
		_, serviceAccountJSONAlternativeExists := secret.Data[api.GCPAlternativeServiceAccountJSON]
		_, userDataExists := secret.Data["userData"]

		if !serviceAccountJSONExists && !serviceAccountJSONAlternativeExists {
			allErrs = append(allErrs, fmt.Errorf("secret %s or %s is required field", api.GCPServiceAccountJSON, api.GCPAlternativeServiceAccountJSON))
		}
		if !userDataExists {
			allErrs = append(allErrs, fmt.Errorf("secret userData is required field"))
		}
	}

	return allErrs
}

func validateGCPMachineClassSpec(spec *api.GCPProviderSpec, fldPath *field.Path) []error {
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
	allErrs = append(allErrs, validateGCPServiceAccounts(spec.ServiceAccounts, fldPath.Child("serviceAccounts"))...)
	allErrs = append(allErrs, validateGCPGpu(spec.Gpu, fldPath.Child("gpu"))...)
	allErrs = append(allErrs, validateGCPScheduling(spec.Scheduling, spec.Gpu, fldPath.Child("scheduling"))...)

	return allErrs
}

func validateGCPDisks(disks []*api.GCPDisk, fldPath *field.Path) []error {
	var allErrs []error

	if 0 == len(disks) {
		allErrs = append(allErrs, field.Required(fldPath, "at least one disk is required"))
	}

	for i, disk := range disks {
		idxPath := fldPath.Index(i)
		if disk.SizeGb < 20 {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("sizeGb"), disk.SizeGb, "disk size must be at least 20 GB"))
		}
		if disk.Type != DiskTypeStandard && disk.Type != DiskTypeSSD && disk.Type != DiskTypeScratch && disk.Type != DiskTypePDBalanced {
			allErrs = append(allErrs, field.NotSupported(idxPath.Child("type"), disk.Type, []string{DiskTypeStandard, DiskTypeSSD, DiskTypeScratch, DiskTypePDBalanced}))
		}
		if disk.Type == DiskTypeScratch && (disk.Interface != DiskInterfaceNVME && disk.Interface != DiskInterfaceSCSI) {
			allErrs = append(allErrs, field.NotSupported(idxPath.Child("interface"), disk.Interface, []string{DiskInterfaceNVME, DiskInterfaceSCSI}))
		}
		if disk.Boot && "" == disk.Image {
			allErrs = append(allErrs, field.Required(idxPath.Child("image"), "image is required for boot disk"))
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
		if "" == nic.Network && "" == nic.Subnetwork {
			allErrs = append(allErrs, field.Required(idxPath, "either network or subnetwork or both is required"))
		}
	}

	return allErrs
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

func validateGCPServiceAccounts(serviceAccounts []api.GCPServiceAccount, fldPath *field.Path) []error {
	var allErrs []error

	if 0 == len(serviceAccounts) {
		allErrs = append(allErrs, field.Required(fldPath, "at least one service account is required"))
	}

	for i, account := range serviceAccounts {
		idxPath := fldPath.Index(i)
		if match, _ := regexp.MatchString(`^[^@]+@(?:[a-zA-Z-0-9]+\.)+[a-zA-Z]{2,}$`, account.Email); !match {
			allErrs = append(allErrs, field.Invalid(idxPath.Child("email"), account.Email, "email address is of invalid format"))
		}
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
