package provider

import (
	"fmt"

	gcp "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	IntegrationTestTag = "mcm-integration-test-true"
)

type ResourcesTrackerImpl struct {
	InitialVolumes   []string
	InitialInstances []string
	InitialMachines  []string
	MachineClass     *v1alpha1.MachineClass
	SecretData       map[string][]byte
	ClusterName      string
}

func (r *ResourcesTrackerImpl) InitializeResourcesTracker(machineClass *v1alpha1.MachineClass, secretData map[string][]byte, clusterName string) error {

	clusterTag := "kubernetes-io-cluster-" + clusterName
	clusterTagValue := "1"
	r.MachineClass = machineClass
	r.SecretData = secretData
	r.ClusterName = clusterName

	ms := gcp.NewGCPPlugin(&gcp.PluginSPIImpl{})
	ctx, svc, err := ms.SPI.NewComputeService(&corev1.Secret{Data: secretData})
	if err != nil {
		return err
	}

	instances, err := getInstancesWithTag(ctx, svc, IntegrationTestTag, machineClass, secretData)

	if err == nil {
		r.InitialInstances = instances
		volumes, err := getAvailableVolumes(ctx, svc, clusterTag, clusterTagValue, machineClass, secretData)
		if err == nil {
			r.InitialVolumes = volumes
			r.InitialMachines, err = getMachines(machineClass, secretData)
			return err
		} else {
			return err
		}
	}
	return err
}

// probeResources will look for resources currently available and returns them
func (r *ResourcesTrackerImpl) probeResources() ([]string, []string, []string, error) {
	// Check for VM instances with matching tags/labels
	// Describe volumes attached to VM instance & delete the volumes
	// Finally delete the VM instance

	clusterTag := "kubernetes-io-cluster-" + r.ClusterName
	clusterTagValue := "1"

	ms := gcp.NewGCPPlugin(&gcp.PluginSPIImpl{})
	ctx, svc, err := ms.SPI.NewComputeService(&corev1.Secret{Data: r.SecretData})
	if err != nil {
		return nil, nil, nil, err
	}

	instances, err := getInstancesWithTag(ctx, svc, IntegrationTestTag, r.MachineClass, r.SecretData)
	if err != nil {
		return instances, nil, nil, err
	}

	// Check for available volumes in cloud provider with tag/label [Status:available]
	availVols, err := getAvailableVolumes(ctx, svc, clusterTag, clusterTagValue, r.MachineClass, r.SecretData)
	if err != nil {
		return instances, availVols, nil, err
	}

	availMachines, _ := getMachines(r.MachineClass, r.SecretData)
	// Check for available vpc and network interfaces in cloud provider with tag
	err = additionalResourcesCheck(ctx, svc, r.ClusterName, r.MachineClass, r.SecretData)

	return instances, availVols, availMachines, err

}

// differenceOrphanedResources checks for difference in the found orphaned resource before test execution with the list after test execution
func differenceOrphanedResources(beforeTestExecution []string, afterTestExecution []string) []string {
	var diff []string

	// Loop two times, first to find beforeTestExecution strings not in afterTestExecution,
	// second loop to find afterTestExecution strings not in beforeTestExecution
	for i := 0; i < 2; i++ {
		for _, b1 := range beforeTestExecution {
			found := false
			for _, a2 := range afterTestExecution {
				if b1 == a2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, b1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			beforeTestExecution, afterTestExecution = afterTestExecution, beforeTestExecution
		}
	}

	return diff
}

/* IsOrphanedResourcesAvailable checks whether there are any orphaned resources left.
If yes, then prints them and returns true. If not, then returns false
*/
func (r *ResourcesTrackerImpl) IsOrphanedResourcesAvailable() bool {
	afterTestExecutionInstances, afterTestExecutionAvailVols, afterTestExecutionAvailmachines, err := r.probeResources()
	//Check there is no error occured
	if err == nil {
		orphanedResourceInstances := differenceOrphanedResources(r.InitialInstances, afterTestExecutionInstances)
		if orphanedResourceInstances != nil {
			fmt.Println("orphaned instances are:", orphanedResourceInstances)
			return true
		}
		orphanedResourceAvailVols := differenceOrphanedResources(r.InitialVolumes, afterTestExecutionAvailVols)
		if orphanedResourceAvailVols != nil {
			fmt.Println("orphaned volumes are:", orphanedResourceAvailVols)
			return true
		}
		orphanedResourceAvailMachines := differenceOrphanedResources(r.InitialMachines, afterTestExecutionAvailmachines)
		if orphanedResourceAvailMachines != nil {
			fmt.Println("orphaned volumes are:", orphanedResourceAvailMachines)
			return true
		}
		return false
	}
	//assuming there are orphaned resources as probe can not be done
	return true
}
