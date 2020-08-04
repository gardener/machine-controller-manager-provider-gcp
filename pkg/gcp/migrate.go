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

// Package gcp contains the cloud provider specific implementations to manage machines
package gcp

import (
	"encoding/json"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/apis"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// ProviderGCP string const to identify GCP provider
	ProviderGCP = "GCP"
)

// fillUpMachineClass copies over the fields from ProviderMachineClass to MachineClass
func fillUpMachineClass(gcpMachineClass *v1alpha1.GCPMachineClass, machineClass *v1alpha1.MachineClass) error {

	disks := []*api.GCPDisk{}
	for _, gcpDisk := range gcpMachineClass.Spec.Disks {
		disk := &api.GCPDisk{
			AutoDelete: gcpDisk.AutoDelete,
			Boot:       gcpDisk.Boot,
			Image:      gcpDisk.Image,
			Interface:  gcpDisk.Interface,
			Labels:     gcpDisk.Labels,
			SizeGb:     gcpDisk.SizeGb,
			Type:       gcpDisk.Type,
		}
		disks = append(disks, disk)
	}

	metaDataList := []*api.GCPMetadata{}
	for _, gcpMetaData := range gcpMachineClass.Spec.Metadata {
		metaData := &api.GCPMetadata{
			Key:   gcpMetaData.Key,
			Value: gcpMetaData.Value,
		}
		metaDataList = append(metaDataList, metaData)
	}

	networkInterfaces := []*api.GCPNetworkInterface{}
	for _, gcpNetworkInterface := range gcpMachineClass.Spec.NetworkInterfaces {
		networkInterface := &api.GCPNetworkInterface{
			DisableExternalIP: gcpNetworkInterface.DisableExternalIP,
			Network:           gcpNetworkInterface.Network,
			Subnetwork:        gcpNetworkInterface.Subnetwork,
		}
		networkInterfaces = append(networkInterfaces, networkInterface)
	}

	scheduling := api.GCPScheduling{
		AutomaticRestart:  gcpMachineClass.Spec.Scheduling.AutomaticRestart,
		OnHostMaintenance: gcpMachineClass.Spec.Scheduling.OnHostMaintenance,
		Preemptible:       gcpMachineClass.Spec.Scheduling.Preemptible,
	}

	serviceAccounts := []api.GCPServiceAccount{}
	for _, gcpServiceAccount := range gcpMachineClass.Spec.ServiceAccounts {
		serviceAccount := api.GCPServiceAccount{
			Email:  gcpServiceAccount.Email,
			Scopes: gcpServiceAccount.Scopes,
		}
		serviceAccounts = append(serviceAccounts, serviceAccount)
	}

	providerSpec := &api.GCPProviderSpec{
		APIVersion:         api.V1alpha1,
		CanIPForward:       gcpMachineClass.Spec.CanIpForward,
		DeletionProtection: gcpMachineClass.Spec.DeletionProtection,
		Description:        gcpMachineClass.Spec.Description,
		Disks:              disks,
		Labels:             gcpMachineClass.Spec.Labels,
		MachineType:        gcpMachineClass.Spec.MachineType,
		Metadata:           metaDataList,
		NetworkInterfaces:  networkInterfaces,
		Region:             gcpMachineClass.Spec.Region,
		Scheduling:         scheduling,
		ServiceAccounts:    serviceAccounts,
		Tags:               gcpMachineClass.Spec.Tags,
		Zone:               gcpMachineClass.Spec.Zone,
	}

	// Marshal providerSpec into Raw Bytes
	providerSpecMarshal, err := json.Marshal(providerSpec)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Migrate finalizers, labels, annotations
	machineClass.Name = gcpMachineClass.Name
	machineClass.Labels = gcpMachineClass.Labels
	machineClass.Annotations = gcpMachineClass.Annotations
	machineClass.Finalizers = gcpMachineClass.Finalizers
	machineClass.ProviderSpec = runtime.RawExtension{
		Raw: providerSpecMarshal,
	}
	machineClass.SecretRef = gcpMachineClass.Spec.SecretRef
	machineClass.Provider = ProviderGCP

	return nil
}
