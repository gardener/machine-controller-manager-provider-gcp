package provider

import (
	"fmt"

	gcp "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// IntegrationTestTag is specifically used for integration test
	// in case the test is run against non seed cluster then the supplied MachineClass
	// is expected to have these tags set so that the machines created from this suite
	// won't be orphan collected.
	IntegrationTestTag = "kubernetes-io-role-integration-test"
)

type ResourcesTrackerImpl struct {
	MachineClass *v1alpha1.MachineClass
	SecretData   map[string][]byte
	ClusterName  string
}

func (r *ResourcesTrackerImpl) InitializeResourcesTracker(machineClass *v1alpha1.MachineClass, secretData map[string][]byte, clusterName string) error {

	r.MachineClass = machineClass
	r.SecretData = secretData
	r.ClusterName = clusterName

	initialVMs, initialVolumes, initialMachines, err := r.probeResources()

	if err != nil {
		fmt.Printf("Error in initial probe of orphaned resources: %s", err.Error())
		return err
	}

	if initialVMs != nil || initialVolumes != nil || initialMachines != nil {
		err := fmt.Errorf("orphan resources are available. Clean them up before proceeding with the test.\nvirtual machines: %v\nvolumes: %v\nmcm machines: %v", initialVMs, initialVolumes, initialMachines)
		return err
	}
	return nil
}

// probeResources will look for orphaned resources and returns
// those resources which could not be deleted in the order
// orphanedInstances, orphanedVolumes, orphanedMachines
func (r *ResourcesTrackerImpl) probeResources() ([]string, []string, []string, error) {

	ms := gcp.NewGCPPlugin(&gcp.PluginSPIImpl{})
	ctx, svc, err := ms.SPI.NewComputeService(&corev1.Secret{Data: r.SecretData})
	if err != nil {
		return nil, nil, nil, err
	}

	orphanedInstances, err := getOrphanedVMs(ctx, svc, IntegrationTestTag, r.MachineClass, r.SecretData)
	if err != nil {
		return orphanedInstances, nil, nil, err
	}
	orphanedVols, err := getOrphanedVolumes(ctx, svc, orphanedInstances, r.MachineClass, r.SecretData)
	if err != nil {
		return orphanedInstances, orphanedVols, nil, err
	}

	orphanedMachines, _ := getMachines(r.MachineClass, r.SecretData)

	return orphanedInstances, orphanedVols, orphanedMachines, err

}

/* IsOrphanedResourcesAvailable checks whether there are any orphaned resources left.
If yes, then prints them and returns true. If not, then returns false
*/
func (r *ResourcesTrackerImpl) IsOrphanedResourcesAvailable() bool {
	afterTestExecutionInstances, afterTestExecutionAvailVols, afterTestExecutionAvailmachines, err := r.probeResources()

	if err != nil {
		fmt.Printf("Error probing orphaned resources: %s", err.Error())
		return true
	}

	if afterTestExecutionInstances != nil || afterTestExecutionAvailVols != nil || afterTestExecutionAvailmachines != nil {
		fmt.Printf("attempting to delete orphaned resources... the following resources are orphaned\n")
		fmt.Printf("Virtual Machines: %v\nVolumes: %v\nMCM Machines: %v\n", afterTestExecutionInstances, afterTestExecutionAvailVols, afterTestExecutionAvailmachines)
		return true
	}
	return false
}
