package provider

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"

	providerDriver "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	compute "google.golang.org/api/compute/v1"
)

// deleteVolume deletes the specified volume
func deleteVolume(ctx context.Context, svc *compute.Service, project, zone, diskName string) error {
	operation, err := svc.Disks.Delete(project, zone, diskName).Context(ctx).Do()
	if err != nil {
		fmt.Printf("Deletion of volume %s failed with error: %s\n", diskName, err.Error())
		return err
	}

	err = providerDriver.WaitUntilOperationCompleted(svc, project, zone, operation.Name)
	if err != nil {
		fmt.Printf("Deletion of volume %s failed with error: %s\n", diskName, err.Error())
		return err
	}

	fmt.Printf("Deleted an orphan disk %s\n,", diskName)

	return nil
}

func getMachines(machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	var (
		machines []string
		spi      providerDriver.PluginSPIImpl
	)
	driverprovider := providerDriver.NewGCPPlugin(&spi)
	machineList, err := driverprovider.ListMachines(context.TODO(), &driver.ListMachinesRequest{
		MachineClass: machineClass,
		Secret:       &v1.Secret{Data: secretData},
	})
	if err != nil {
		return nil, err
	} else if len(machineList.MachineList) != 0 {
		fmt.Printf("\nAvailable Machines: ")
		for _, machine := range machineList.MachineList {
			fmt.Printf("%s\n", machine)
			machines = append(machines, machine)
		}
	}
	return machines, nil
}

// getOrphanedVMs returns a list of VMs with Integration Test tag which couldn't be deleted
func getOrphanedVMs(ctx context.Context, svc *compute.Service, project, zone string) ([]string, error) {
	var (
		instancesID []string
	)

	req := svc.Instances.List(project, zone)
	if err := req.Pages(ctx, func(page *compute.InstanceList) error {
		for _, server := range page.Items {
			//in gcp the tags are just string, not key value pair
			for _, tag := range server.Tags.Items {
				if tag == IntegrationTestTag {
					if err := terminateInstance(svc, project, zone, server.Name); err != nil {
						instancesID = append(instancesID, server.Name)
					}
					break
				}
			}
		}
		return nil
	}); err != nil {
		return instancesID, err
	}

	return instancesID, nil
}

// getOrphanedVolumes returns a list of disks which couldn't be deleted
func getOrphanedVolumes(ctx context.Context, svc *compute.Service, project string, zone string, orphanDisks []string) ([]string, error) {
	var availVolID []string

	for _, diskName := range orphanDisks {
		_, err := svc.Disks.Get(project, zone, diskName).Context(ctx).Do()
		if err == nil {
			if err = deleteVolume(ctx, svc, project, zone, diskName); err != nil {
				availVolID = append(availVolID, diskName)
			}
		}
	}

	return availVolID, nil
}

//terminateInstance terminates a specified instance
func terminateInstance(svc *compute.Service, project, zone, instanceName string) error {
	operation, err := svc.Instances.Delete(project, zone, instanceName).Context(context.Background()).Do()
	if err != nil {
		fmt.Printf("can't terminate the instance %s, %s\n", instanceName, err.Error())
		return err
	}

	err = providerDriver.WaitUntilOperationCompleted(svc, project, zone, operation.Name)
	if err != nil {
		fmt.Printf("Deletion of instance %s failed with error: %s\n", instanceName, err.Error())
		return err
	}

	fmt.Printf("Deleted an orphan VM %s\n,", instanceName)
	return nil
}
