package provider

/**
	Orphaned Resources
	- VMs:
		Describe instances with specified tag name:<cluster-name>
		Report/Print out instances found
		Describe volumes attached to the instance (using instance id)
		Report/Print out volumes found
		Delete attached volumes found
		Terminate instances found
	- Disks:
		Describe volumes with tag status:available
		Report/Print out volumes found
		Delete identified volumes
**/

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	v1 "k8s.io/api/core/v1"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
	providerDriver "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	compute "google.golang.org/api/compute/v1"
)

func getMachines(machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	var machines []string
	var spi providerDriver.PluginSPIImpl
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

// getInstancesWithTag describes the instance with the specified tag and terminates it
func getInstancesWithTag(ctx context.Context, svc *compute.Service, searchTagName string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	var instancesID []string
	project, err := providerDriver.ExtractProject(secretData)
	if err != nil {
		return nil, err
	}

	var providerSpec *api.GCPProviderSpec

	err = json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {

		providerSpec = nil
		log.Printf("Error occured while performing unmarshal %s", err.Error())
		return instancesID, err
	}

	zone := providerSpec.Zone
	req := svc.Instances.List(project, zone)
	if err := req.Pages(ctx, func(page *compute.InstanceList) error {
		for _, server := range page.Items {
			//in gcp the tags are just string, not key value pair
			for _, tag := range server.Tags.Items {
				if tag == searchTagName {
					instancesID = append(instancesID, server.Name)

					TerminateInstance(svc, project, zone, server.Name)

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

//TerminateInstance terminates a specified instance
func TerminateInstance(svc *compute.Service, project, zone, instanceName string) {
	operation, err := svc.Instances.Delete(project, zone, instanceName).Context(context.Background()).Do()
	if err != nil {
		fmt.Printf("can't terminate the instance %s, %s\n", instanceName, err.Error())
	}

	err = providerDriver.WaitUntilOperationCompleted(svc, project, zone, operation.Name)
	if err != nil {
		fmt.Printf("Deletion of volume %s failed with error: %s\n", instanceName, err.Error())
	}
}

// getAvailableVolumes gets list of possible orphaned volumes ,verifies if they still exist and if yes then deletes them
func getAvailableVolumes(ctx context.Context, svc *compute.Service, orphanDisks []string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
	var availVolID []string
	project, err := providerDriver.ExtractProject(secretData)
	if err != nil {
		return nil, err
	}

	var providerSpec *api.GCPProviderSpec

	err = json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {
		providerSpec = nil
		log.Printf("Error occured while performing unmarshal %s", err.Error())
		return availVolID, err
	}

	zone := providerSpec.Zone
	// req := svc.Disks.List(project, zone).Filter("status = READY")

	// //check whether these tags are present on the disks or not
	// if err := req.Pages(ctx, func(page *compute.DiskList) error {
	// 	for _, disk := range page.Items {
	// 		for key, value := range disk.Labels {

	// 			// #the disks didn't have any tag, check after getting access
	// 			if strings.Contains(key, tagName) && strings.Contains(value, tagValue) {
	// 				fmt.Printf("%s", disk.Name)
	// 				availVolID = append(availVolID, disk.Name)

	// 				//delete the disk
	// 				//deleteVolume(ctx, svc, project, zone, disk.Name)
	// 				break
	// 			}
	// 		}
	// 	}
	// 	return nil
	// }); err != nil {
	// 	return availVolID, err
	// }
	for _, diskName := range orphanDisks {
		_, err := svc.Disks.Get(project, zone, diskName).Context(ctx).Do()
		if err == nil {
			availVolID = append(availVolID, diskName)

			deleteVolume(ctx, svc, project, zone, diskName)
		}
	}

	return availVolID, nil
}

// deleteVolume deletes the specified volume
func deleteVolume(ctx context.Context, svc *compute.Service, project, zone, diskName string) {
	operation, err := svc.Disks.Delete(project, zone, diskName).Context(ctx).Do()
	if err != nil {
		fmt.Printf("Deletion of volume %s failed with error: %s\n", diskName, err.Error())
	}

	err = providerDriver.WaitUntilOperationCompleted(svc, project, zone, operation.Name)
	if err != nil {
		fmt.Printf("Deletion of volume %s failed with error: %s\n", diskName, err.Error())
	}
}