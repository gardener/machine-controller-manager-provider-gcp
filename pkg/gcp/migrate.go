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
	"context"
	"encoding/json"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/apis"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machineutils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

const (
	// MachineControllerManagerProviderGCP is the string constant to identify machine controller GCP
	MachineControllerManagerProviderGCP = "machine-controller-manager-provider-gcp"
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

	return nil
}

// createMachineClass creates the generic machineClass corresponding to the providerMachineClass
func (ms *MachinePlugin) createMachineClass(ctx context.Context, req *driver.MigrateMachineClassRequest) error {
	gcpMachineClass, err := req.MachineClient.GCPMachineClasses(req.Namespace).Get(req.ClassSpec.Name, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	machineClass, err := req.MachineClient.MachineClasses(req.Namespace).Get(req.ClassSpec.Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		// MachineClass doesn't exist, create a fresh one
		machineClass = &v1alpha1.MachineClass{}
		fillUpMachineClass(gcpMachineClass, machineClass)

		_, err = req.MachineClient.MachineClasses(req.Namespace).Create(machineClass)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

	} else if err == nil {
		// MachineClass exists, and needs to be updated
		machineClass = machineClass.DeepCopy()
		fillUpMachineClass(gcpMachineClass, machineClass)

		_, err = req.MachineClient.MachineClasses(req.Namespace).Update(machineClass)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		return status.Error(codes.Internal, err.Error())
	}

	klog.V(2).Info("Create/Apply successful for MachineClass %s/%s", req.ClassSpec.Kind, req.ClassSpec.Name)

	return nil
}

// updateClassReferences updates all machine objects to refer to the new MachineClass.
func (ms *MachinePlugin) updateClassReferences(ctx context.Context, req *driver.MigrateMachineClassRequest) error {

	var (
		className = req.ClassSpec.Name
		classKind = req.ClassSpec.Kind
		namespace = req.Namespace
	)

	// Update Machines
	machineList, err := req.MachineClient.Machines(namespace).List(metav1.ListOptions{})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for _, machine := range machineList.Items {
		if machine.Spec.Class.Name == className && machine.Spec.Class.Kind == classKind {
			clone := machine.DeepCopy()
			clone.Spec.Class.Kind = MachineClassKind

			_, err := req.MachineClient.Machines(namespace).Update(clone)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			klog.V(1).Infof("Updated class reference for machine %s/%s", namespace, machine.Name)
		}
	}

	// Update MachineSets
	machineSetList, err := req.MachineClient.MachineSets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for _, machineSet := range machineSetList.Items {
		if machineSet.Spec.Template.Spec.Class.Name == className &&
			machineSet.Spec.Template.Spec.Class.Kind == classKind {

			clone := machineSet.DeepCopy()
			clone.Spec.Template.Spec.Class.Kind = MachineClassKind

			_, err := req.MachineClient.MachineSets(namespace).Update(clone)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			klog.V(1).Infof("Updated class reference for machineSet %s/%s", namespace, machineSet.Name)
		}
	}

	// Update MachineDeployments
	machineDeploymentList, err := req.MachineClient.MachineDeployments(req.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for _, machineDeployment := range machineDeploymentList.Items {
		if machineDeployment.Spec.Template.Spec.Class.Name == className &&
			machineDeployment.Spec.Template.Spec.Class.Kind == classKind {

			clone := machineDeployment.DeepCopy()
			clone.Spec.Template.Spec.Class.Kind = MachineClassKind

			_, err := req.MachineClient.MachineDeployments(req.Namespace).Update(clone)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			klog.V(1).Infof("Updated class reference for machineDeployment %s/%s", namespace, machineDeployment.Name)
		}
	}

	return nil
}

// addMigratedAnnotationForProviderMachineClass adds ignore provider MachineClass annotation
func (ms *MachinePlugin) addMigratedAnnotationForProviderMachineClass(ctx context.Context, req *driver.MigrateMachineClassRequest) error {

	gcpMachineClass, err := req.MachineClient.GCPMachineClasses(req.Namespace).Get(req.ClassSpec.Name, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	clone := gcpMachineClass.DeepCopy()
	if clone.Annotations == nil {
		clone.Annotations = make(map[string]string)
	}
	clone.Annotations[machineutils.MigratedMachineClass] = MachineControllerManagerProviderGCP

	_, err = req.MachineClient.GCPMachineClasses(req.Namespace).Update(clone)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	klog.V(1).Infof("Set migrated annotation for GCPMachineClass %s/%s", req.Namespace, req.ClassSpec.Name)

	return nil
}
