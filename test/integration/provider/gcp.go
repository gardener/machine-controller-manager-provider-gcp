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
	"strings"

	v1 "k8s.io/api/core/v1"

	"net/http"

	api "github.com/gardener/machine-controller-manager-provider-gcp/pkg/api/v1alpha1"
	providerDriver "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
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

// getInstancesWithTag describes the instance with the specified tag
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

// getAvailableVolumes describes volumes with the specified tag
func getAvailableVolumes(ctx context.Context, svc *compute.Service, tagName string, tagValue string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) ([]string, error) {
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
	req := svc.Disks.List(project, zone)

	//check whether these tags are present on the disks or not
	if err := req.Pages(ctx, func(page *compute.DiskList) error {
		for _, disk := range page.Items {
			for key, value := range disk.Labels {

				// #the disks didn't have any tag, check after getting access
				if strings.Contains(key, tagName) && strings.Contains(value, tagValue) {
					fmt.Printf("%s", disk.Name)
					availVolID = append(availVolID, disk.Name)

					//delete the disk
					deleteVolume(ctx, svc, project, zone, disk.Name)
					break
				}
			}
		}
		return nil
	}); err != nil {
		return availVolID, err
	}

	return availVolID, nil
}

// deleteVolume deletes the specified volume
func deleteVolume(ctx context.Context, svc *compute.Service, project, zone, diskName string) error {
	operation, err := svc.Disks.Delete(project, zone, diskName).Context(ctx).Do()
	if err != nil {
		if ae, ok := err.(*googleapi.Error); ok && ae.Code == http.StatusNotFound {
			return nil
		}
		return err
	}

	return providerDriver.WaitUntilOperationCompleted(svc, project, zone, operation.Name)
}

// additionalResourcesCheck describes VPCs and network interfaces
func additionalResourcesCheck(ctx context.Context, svc *compute.Service, clusterName string, machineClass *v1alpha1.MachineClass, secretData map[string][]byte) error {
	// Checks for Network interfaces and VPCs
	// If the command succeeds, no output is returned.
	project, err := providerDriver.ExtractProject(secretData)
	if err != nil {
		return err
	}

	var providerSpec *api.GCPProviderSpec

	err = json.Unmarshal([]byte(machineClass.ProviderSpec.Raw), &providerSpec)
	if err != nil {
		providerSpec = nil
		log.Printf("Error occured while performing unmarshal %s", err.Error())
		return err
	}

	zone := providerSpec.Zone

	networkFilter := "name=" + clusterName
	resultNetwork := svc.Networks.List(project).Filter(networkFilter)

	if err := resultNetwork.Pages(ctx, func(page *compute.NetworkList) error {
		for _, network := range page.Items {
			fmt.Println(network.Id)
			//list all the instances
			req := svc.Instances.List(project, zone)
			//for every nic of each instance see if nic.network contains network.name
			//if yes , print it.
			if err := req.Pages(ctx, func(page *compute.InstanceList) error {
				for _, instance := range page.Items {
					for _, networkInterface := range instance.NetworkInterfaces {
						if strings.Contains(networkInterface.Network, network.Name) {
							fmt.Printf("nic: %s\n", networkInterface.Name)
						}
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
