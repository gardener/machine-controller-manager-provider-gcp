// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package gcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/utils/ptr"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
	errors2 "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/errors"
	"github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/validation"
)

const (
	// ProviderPrefix is the prefix used by the GCP provider
	ProviderPrefix = "gce://"
)

// CreateMachineUtil method is used to create a GCP machine
func (ms *MachinePlugin) CreateMachineUtil(_ context.Context, machineName string, providerSpec *api.GCPProviderSpec, secret *corev1.Secret) (string, error) {
	ctx, computeService, err := ms.SPI.NewComputeService(secret)
	if err != nil {
		return "", err
	}

	project, err := ExtractProject(secret.Data)
	if err != nil {
		return "", err
	}
	var (
		zone = providerSpec.Zone

		instance = &compute.Instance{
			CanIpForward:       providerSpec.CanIPForward,
			DeletionProtection: providerSpec.DeletionProtection,
			Labels:             providerSpec.Labels,
			MachineType:        fmt.Sprintf("zones/%s/machineTypes/%s", zone, providerSpec.MachineType),
			MinCpuPlatform:     providerSpec.MinCPUPlatform,
			Name:               machineName,
			Scheduling: &compute.Scheduling{
				AutomaticRestart:  &providerSpec.Scheduling.AutomaticRestart,
				OnHostMaintenance: providerSpec.Scheduling.OnHostMaintenance,
				Preemptible:       providerSpec.Scheduling.Preemptible,
			},
			Tags: &compute.Tags{
				Items: providerSpec.Tags,
			},
		}
	)

	if providerSpec.Gpu != nil {
		instance.GuestAccelerators = []*compute.AcceleratorConfig{
			{
				AcceleratorType:  fmt.Sprintf("projects/%s/zones/%s/acceleratorTypes/%s", project, zone, providerSpec.Gpu.AcceleratorType),
				AcceleratorCount: providerSpec.Gpu.Count,
			},
		}
	}

	if providerSpec.Description != nil {
		instance.Description = *providerSpec.Description
	}

	instance.Disks = createAttachedDisks(providerSpec.Disks, zone, machineName)

	var metadataItems []*compute.MetadataItems
	metadataItems = append(metadataItems, getUserData(string(secret.Data["userData"])))

	for _, metadata := range providerSpec.Metadata {
		metadataItems = append(metadataItems, &compute.MetadataItems{
			Key:   metadata.Key,
			Value: metadata.Value,
		})
	}
	instance.Metadata = &compute.Metadata{
		Items: metadataItems,
	}

	var networkInterfaces []*compute.NetworkInterface
	for _, nic := range providerSpec.NetworkInterfaces {
		computeNIC := &compute.NetworkInterface{}

		if !nic.DisableExternalIP {
			// When DisableExternalIP is false, implies Attach an external IP to VM
			computeNIC.AccessConfigs = []*compute.AccessConfig{{}}
		}
		if len(nic.Network) != 0 {
			computeNIC.Network = fmt.Sprintf("projects/%s/global/networks/%s", project, nic.Network)
		}
		if len(nic.Subnetwork) != 0 {
			computeNIC.Subnetwork = fmt.Sprintf("regions/%s/subnetworks/%s", providerSpec.Region, nic.Subnetwork)
		}
		networkInterfaces = append(networkInterfaces, computeNIC)
	}
	instance.NetworkInterfaces = networkInterfaces

	var serviceAccounts []*compute.ServiceAccount
	for _, sa := range providerSpec.ServiceAccounts {
		serviceAccounts = append(serviceAccounts, &compute.ServiceAccount{
			Email:  sa.Email,
			Scopes: sa.Scopes,
		})
	}
	instance.ServiceAccounts = serviceAccounts
	operation, err := computeService.Instances.Insert(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return "", classifyIfResourceExhaustedError(err)
	}

	if err := WaitUntilOperationCompleted(computeService, project, zone, operation.Name); err != nil {
		return "", err
	}

	return encodeMachineID(project, zone, machineName), nil
}

func createAttachedDisks(disks []*api.GCPDisk, zone, machineName string) []*compute.AttachedDisk {
	attachedDisks := make([]*compute.AttachedDisk, 0, len(disks))
	for _, disk := range disks {
		var attachedDisk compute.AttachedDisk
		switch disk.Type {
		case api.GCPDiskTypeScratch:
			attachedDisk = compute.AttachedDisk{
				Type:       api.GCPDiskTypeScratch,
				Boot:       false,
				AutoDelete: ptr.Deref(disk.AutoDelete, true),
				Interface:  disk.Interface,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskType: fmt.Sprintf("zones/%s/diskTypes/%s", zone, "local-ssd"),
				},
			}
		default:
			attachedDisk = compute.AttachedDisk{
				Type:       api.GCPDiskTypePersistent,
				Boot:       disk.Boot,
				AutoDelete: ptr.Deref(disk.AutoDelete, true),
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskSizeGb:            disk.SizeGb,
					DiskType:              fmt.Sprintf("zones/%s/diskTypes/%s", zone, disk.Type),
					Labels:                disk.Labels,
					SourceImage:           disk.Image,
					ProvisionedIops:       disk.ProvisionedIops,
					ProvisionedThroughput: disk.ProvisionedThroughput,
				},
			}
		}

		if disk.Encryption != nil {
			attachedDisk.DiskEncryptionKey = &compute.CustomerEncryptionKey{
				KmsKeyName:           strings.TrimSpace(disk.Encryption.KmsKeyName),
				KmsKeyServiceAccount: strings.TrimSpace(disk.Encryption.KmsKeyServiceAccount),
			}
			klog.V(3).Infof("(CreateMachineUtil) For machineName: %q, diskLabel: %q, DiskEncryptionKey.KmsKeyName: %q, "+
				"DiskEncryptionKey.KmsKeyServiceAccount: %q",
				machineName,
				disk.Labels["name"],
				attachedDisk.DiskEncryptionKey.KmsKeyName,
				attachedDisk.DiskEncryptionKey.KmsKeyServiceAccount)
		}
		attachedDisks = append(attachedDisks, &attachedDisk)
	}
	return attachedDisks
}

func encodeMachineID(project, zone, name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s/%s", ProviderPrefix, project, zone, name)
}

// DeleteMachineUtil deletes a VM by name
func (ms *MachinePlugin) DeleteMachineUtil(_ context.Context, machineName string, _ string, providerSpec *api.GCPProviderSpec, secret *corev1.Secret) (string, error) {
	ctx, computeService, err := ms.SPI.NewComputeService(secret)
	if err != nil {
		return "", err
	}

	project, err := ExtractProject(secret.Data)
	if err != nil {
		return "", err
	}

	zone := providerSpec.Zone

	result, err := getVMs(ctx, machineName, providerSpec, secret, project, computeService)
	if err != nil {
		return "", err
	} else if len(result) == 0 {
		return "", &errors2.MachineNotFoundError{Name: machineName}
	}

	operation, err := computeService.Instances.Delete(project, zone, machineName).Context(ctx).Do()
	if err != nil {
		if ae, ok := err.(*googleapi.Error); ok && ae.Code == http.StatusNotFound {
			return "", nil
		}
		return "", err
	}

	return encodeMachineID(project, zone, machineName), WaitUntilOperationCompleted(computeService, project, zone, operation.Name)
}

// GetMachineStatusUtil checks for existence of VM by name
func (ms *MachinePlugin) GetMachineStatusUtil(_ context.Context, machineName string, _ string, providerSpec *api.GCPProviderSpec, secret *corev1.Secret) (string, error) {
	ctx, computeService, err := ms.SPI.NewComputeService(secret)
	if err != nil {
		return "", err
	}

	project, err := ExtractProject(secret.Data)
	if err != nil {
		return "", err
	}
	zone := providerSpec.Zone

	result, err := getVMs(ctx, machineName, providerSpec, secret, project, computeService)
	if err != nil {
		return "", err
	} else if len(result) == 0 {
		// No running instance exists with the given machine-ID
		return "", &errors2.MachineNotFoundError{Name: machineName}
	}

	return encodeMachineID(project, zone, machineName), nil
}

// ListMachinesUtil lists all VMs in the DC or folder
func (ms *MachinePlugin) ListMachinesUtil(_ context.Context, providerSpec *api.GCPProviderSpec, secret *corev1.Secret) (map[string]string, error) {
	ctx, computeService, err := ms.SPI.NewComputeService(secret)
	if err != nil {
		return nil, err
	}

	project, err := ExtractProject(secret.Data)
	if err != nil {
		return nil, err
	}

	result, err := getVMs(ctx, "", providerSpec, secret, project, computeService)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getVMs(ctx context.Context, machineID string, providerSpec *api.GCPProviderSpec, _ *corev1.Secret, project string, computeService *compute.Service) (map[string]string, error) {
	listOfVMs := make(map[string]string)

	searchClusterName := ""
	searchNodeRole := ""

	for _, key := range providerSpec.Tags {
		if strings.Contains(key, "kubernetes-io-cluster-") {
			searchClusterName = key
		} else if strings.Contains(key, "kubernetes-io-role-") {
			searchNodeRole = key
		}
	}

	if searchClusterName == "" || searchNodeRole == "" {
		return listOfVMs, nil
	}

	zone := providerSpec.Zone

	req := computeService.Instances.List(project, zone)
	if err := req.Pages(ctx, func(page *compute.InstanceList) error {
		for _, server := range page.Items {
			clusterName := ""
			nodeRole := ""

			for _, key := range server.Tags.Items {
				if strings.Contains(key, "kubernetes-io-cluster-") {
					clusterName = key
				} else if strings.Contains(key, "kubernetes-io-role-") {
					nodeRole = key
				}
			}

			if clusterName == searchClusterName && nodeRole == searchNodeRole {
				instanceID := server.Name

				if machineID == "" {
					listOfVMs[encodeMachineID(project, zone, instanceID)] = server.Name
				} else if machineID == instanceID {
					listOfVMs[encodeMachineID(project, zone, instanceID)] = server.Name
					klog.V(3).Infof("Found machine with name: %q", server.Name)
					break
				}
			}
		}
		return nil
	}); err != nil {
		return listOfVMs, err
	}

	return listOfVMs, nil
}

// decodeProviderSpec converts request parameters to api.ProviderSpec
func decodeProviderSpec(machineClass *v1alpha1.MachineClass) (*api.GCPProviderSpec, error) {
	var providerSpec *api.GCPProviderSpec
	// If machineClass is nil
	if machineClass == nil {
		return nil, status.Error(codes.Internal, "MachineClass ProviderSpec is nil")
	}
	// Extract providerSpec
	err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return providerSpec, nil
}

func validateProviderSpec(providerSpec *api.GCPProviderSpec) error {
	if validationErr := validation.ValidateProviderSpec(providerSpec); validationErr != nil {
		err := fmt.Errorf("error while validating ProviderSpec %v", validationErr)
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func validateSecret(secret *corev1.Secret) error {
	validationErr := validation.ValidateSecret(secret)
	if validationErr != nil {
		err := fmt.Errorf("error while validating Secret %v", validationErr)
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func validateZone(zone string) error {
	if err := validation.ValidateZone(zone); err != nil {
		err = fmt.Errorf("error while validating Zone %v", err)
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

func prepareErrorf(err error, format string, args ...interface{}) error {
	var (
		code    codes.Code
		wrapped error
	)
	switch err.(type) {
	case *errors2.MachineNotFoundError:
		code = codes.NotFound
		wrapped = err
	case *errors2.MachineResourceExhaustedError:
		code = codes.ResourceExhausted
		wrapped = errors.Wrap(err, fmt.Sprintf(format, args...))
	default:
		code = codes.Internal
		wrapped = errors.Wrap(err, fmt.Sprintf(format, args...))
	}
	klog.V(2).Infof(wrapped.Error())
	return status.Error(code, wrapped.Error())
}

// ExtractProject returns the name of the project which is extracted from the secret
func ExtractProject(credentialsData map[string][]byte) (string, error) {
	serviceAccountJSON := extractCredentialsFromData(credentialsData, api.GCPServiceAccountJSON, api.GCPAlternativeServiceAccountJSON)

	var j struct {
		Project string `json:"project_id"`
	}
	if err := json.Unmarshal([]byte(serviceAccountJSON), &j); err != nil {
		return "Error", err
	}
	return j.Project, nil
}

// WaitUntilOperationCompleted waits for the specified operation to be completed and returns true if it does else returns false
func WaitUntilOperationCompleted(computeService *compute.Service, project, zone, operationName string) error {
	return wait.Poll(5*time.Second, 300*time.Second, func() (bool, error) {
		op, err := computeService.ZoneOperations.Get(project, zone, operationName).Do()
		if err != nil {
			return false, err
		}
		klog.V(3).Infof("Waiting for operation to be completed... (status: %s)", op.Status)
		if op.Status == "DONE" {
			if op.Error == nil {
				return true, nil
			}
			var (
				errorMessages []string
				latestOpErr   *compute.OperationErrorErrors
			)
			for _, opErr := range op.Error.Errors {
				latestOpErr = opErr
				errorMessages = append(errorMessages, opErr.Message)
			}

			return false, checkIfResourceExhaustedError(latestOpErr, errorMessages)
		}
		return false, nil
	})
}

func getUserData(userData string) *compute.MetadataItems {
	if strings.HasPrefix(userData, "#cloud-config") {
		return &compute.MetadataItems{
			Key:   "user-data",
			Value: &userData,
		}
	}

	return &compute.MetadataItems{
		Key:   "startup-script",
		Value: &userData,
	}
}

func classifyIfResourceExhaustedError(err error) error {
	gerr, ok := err.(*googleapi.Error)
	// https://cloud.google.com/compute/docs/troubleshooting/troubleshooting-vm-creation#zone_availability also depends on error message, that's why adopted this approach
	if ok && strings.Contains(gerr.Message, "does not exist in zone") {
		return &errors2.MachineResourceExhaustedError{Msg: err.Error()}
	}
	return err
}

func checkIfResourceExhaustedError(opErr *compute.OperationErrorErrors, errorMessages []string) error {
	combinedErrMsg := strings.Join(errorMessages, "; ")
	if opErr.Code == "RESOURCE_POOL_EXHAUSTED" || opErr.Code == "ZONE_RESOURCE_POOL_EXHAUSTED" || opErr.Code == "ZONE_RESOURCE_POOL_EXHAUSTED_WITH_DETAILS" || strings.Contains(opErr.Code, "QUOTA") {
		return &errors2.MachineResourceExhaustedError{Msg: combinedErrMsg}
	}
	return fmt.Errorf(combinedErrMsg)
}
