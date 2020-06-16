/*
 * Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
 * limitations under the License.
 *
 */

package gcp

import (
	"context"
	"net/http"

	fake "github.com/gardener/machine-controller-manager-provider-gcp/pkg/gcp/fake"
	v1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	//FailAtNotFound is the error message returned when a resource is not found
	FailAtNotFound string = "machine codes error: code = [NotFound] message = [machine name=non-existent-dummy-machine, uuid= not found]"
	//FailAtJSONUnmarshalling is the error message returned when an malformed JSON is sent to the plugin by the caller
	FailAtJSONUnmarshalling string = "machine codes error: code = [Internal] message = [Machine status \"dummy-machine\" failed on decodeProviderSpecAndSecret: machine codes error: code = [Internal] message = [unexpected end of JSON input]]"
	//CreateFailAtJSONUnmarshalling is the error message returned when an malformed JSON is sent to the plugin by the caller
	CreateFailAtJSONUnmarshalling string = "machine codes error: code = [Internal] message = [Create machine \"dummy-machine\" failed on decodeProviderSpecAndSecret: machine codes error: code = [Internal] message = [unexpected end of JSON input]]"
	//DeleteFailAtJSONUnmarshalling is the error message returned when an malformed JSON is sent to the plugin by the caller
	DeleteFailAtJSONUnmarshalling string = "machine codes error: code = [Internal] message = [Delete machine \"dummy-machine\" failed on decodeProviderSpecAndSecret: machine codes error: code = [Internal] message = [unexpected end of JSON input]]"
	//ListFailAtJSONUnmarshalling is the error message returned when an malformed JSON is sent to the plugin by the caller
	ListFailAtJSONUnmarshalling string = "machine codes error: code = [Internal] message = [List machines failed on decodeProviderSpecAndSecret: machine codes error: code = [Internal] message = [unexpected end of JSON input]]"
	//FailAtNoSecretsPassed is the error message returned when no secrets are passed to the the plugin by the caller
	FailAtNoSecretsPassed string = "machine codes error: code = [Internal] message = [Create machine \"dummy-machine\" failed on decodeProviderSpecAndSecret: machine codes error: code = [Internal] message = [Error while validating ProviderSpec [Secret serviceAccountJSON is required field Secret userData is required field]]]"
	//FailAtSecretsWithNoUserData is the error message returned when secrets map has no userdata provided by the caller
	FailAtSecretsWithNoUserData string = "machine codes error: code = [Internal] message = [Create machine \"dummy-machine\" failed on decodeProviderSpecAndSecret: machine codes error: code = [Internal] message = [Error while validating ProviderSpec [Secret userData is required field]]]"
	// FailAtInvalidProjectID is the error returned when an invalid project id value is provided by the caller
	FailAtInvalidProjectID string = "machine codes error: code = [Internal] message = [Create machine \"dummy-machine\" failed: json: cannot unmarshal number into Go struct field .project_id of type string]"
	//FailAtInvalidZonePostCall is the  error returned when a post call should fail with an invalid zone is sent in the POST call -- this is used to simulate server error
	FailAtInvalidZonePostCall string = "machine codes error: code = [Internal] message = [Create machine \"dummy-machine\" failed: googleapi: got HTTP response code 400 with body: Invalid post zone\n]"
	//FailAtInvalidZoneListCall is the  error returned when a list call should fail with an invalid zone is sent in the LIST call -- this is used to simulate server error
	FailAtInvalidZoneListCall string = "machine codes error: code = [Internal] message = [Machine status \"dummy-machine\" failed: googleapi: got HTTP response code 400 with body: Invalid list zone\n]"
	//CreateFailAtInvalidZoneListCall is the  error returned when a list call should fail with an invalid zone is sent in the CREATE call -- this is used to simulate server error
	CreateFailAtInvalidZoneListCall string = "machine codes error: code = [Internal] message = [Create machine \"dummy-machine\" failed: googleapi: got HTTP response code 400 with body: Invalid list zone\n]"
	//DeleteFailAtInvalidZoneListCall is the  error returned when a list call should fail with an invalid zone is sent in the DELETE call -- this is used to simulate server error
	DeleteFailAtInvalidZoneListCall string = "machine codes error: code = [Internal] message = [Delete machine \"dummy-machine\" failed: googleapi: got HTTP response code 400 with body: Invalid list zone\n]"
	//ListFailAtInvalidZoneListCall is the  error returned when a list call should fail with an invalid zone is sent in the LIST call -- this is used to simulate server error
	ListFailAtInvalidZoneListCall string = "machine codes error: code = [Internal] message = [List machines failed: googleapi: got HTTP response code 400 with body: Invalid list zone\n]"
	//FailAtMethodNotImplemented is the error returned for methods which are not yet implemented
	FailAtMethodNotImplemented string = "rpc error: code = Unimplemented desc = "
	FailAtSpecValidation       string = "machine codes error: code = [Internal] message = [Create machine \"dummy-machine\" failed on decodeProviderSpecAndSecret: machine codes error: code = [Internal] message = [Error while validating ProviderSpec [spec.zone: Required value: zone is required]]]"
	FailAtNonExistingMachine   string = "rpc error: code = NotFound desc = Machine with the name \"non-existent-dummy-machine\" not found"
)

var _ = Describe("MachineController", func() {
	gcpProviderSpec := []byte("{\"canIpForward\":true,\"deletionProtection\":false,\"description\":\"Machine created to test out-of-tree gcp mcm driver.\",\"disks\":[{\"autoDelete\":true,\"boot\":true,\"sizeGb\":50,\"type\":\"pd-standard\",\"image\":\"projects/coreos-cloud/global/images/coreos-stable-2135-6-0-v20190801\",\"labels\":{\"name\":\"test-mc-gcp\"}}],\"labels\":{\"name\":\"test-mc-gcp\"},\"machineType\":\"n1-standard-2\",\"metadata\":[{\"key\":\"gcp\",\"value\":\"my-value\"}],\"networkInterfaces\":[{\"network\":\"dummyShoot\",\"subnetwork\":\"dummyShoot\"}],\"scheduling\":{\"automaticRestart\":true,\"onHostMaintenance\":\"MIGRATE\",\"preemptible\":false},\"secretRef\":{\"name\":\"dummySecret\",\"namespace\":\"dummy\"},\"serviceAccounts\":[{\"email\":\"mcmDummy@dummy.com\",\"scopes\":[\"https://www.googleapis.com/auth/compute\"]}],\"tags\":[\"kubernetes-io-cluster-dummy-machine\",\"kubernetes-io-role-mcm\",\"dummy-machine\"],\"region\":\"europe-dummy\",\"zone\":\"europe-dummy\"}")
	gcpProviderSpecValidationErr := []byte("{\"canIpForward\":true,\"deletionProtection\":false,\"description\":\"Machine created to test out-of-tree gcp mcm driver.\",\"disks\":[{\"autoDelete\":true,\"boot\":true,\"sizeGb\":50,\"type\":\"pd-standard\",\"image\":\"projects/coreos-cloud/global/images/coreos-stable-2135-6-0-v20190801\",\"labels\":{\"name\":\"test-mc-gcp\"}}],\"labels\":{\"name\":\"test-mc-gcp\"},\"machineType\":\"n1-standard-2\",\"metadata\":[{\"key\":\"gcp\",\"value\":\"my-value\"}],\"networkInterfaces\":[{\"network\":\"dummyShoot\",\"subnetwork\":\"dummyShoot\"}],\"scheduling\":{\"automaticRestart\":true,\"onHostMaintenance\":\"MIGRATE\",\"preemptible\":false},\"secretRef\":{\"name\":\"dummySecret\",\"namespace\":\"dummy\"},\"serviceAccounts\":[{\"email\":\"mcmDummy@dummy.com\",\"scopes\":[\"https://www.googleapis.com/auth/compute\"]}],\"tags\":[\"kubernetes-io-cluster-dummy-machine\",\"kubernetes-io-role-mcm\",\"dummy-machine\"],\"region\":\"europe-dummy\",\"zone\":\"\"}")
	gcpProviderSpecNoTagsToSearch := []byte("{\"canIpForward\":true,\"deletionProtection\":false,\"description\":\"Machine created to test out-of-tree gcp mcm driver.\",\"disks\":[{\"autoDelete\":true,\"boot\":true,\"sizeGb\":50,\"type\":\"pd-standard\",\"image\":\"projects/coreos-cloud/global/images/coreos-stable-2135-6-0-v20190801\",\"labels\":{\"name\":\"test-mc-gcp\"}}],\"labels\":{\"name\":\"test-mc-gcp\"},\"machineType\":\"n1-standard-2\",\"metadata\":[{\"key\":\"gcp\",\"value\":\"my-value\"}],\"networkInterfaces\":[{\"network\":\"dummyShoot\",\"subnetwork\":\"dummyShoot\"}],\"scheduling\":{\"automaticRestart\":true,\"onHostMaintenance\":\"MIGRATE\",\"preemptible\":false},\"secretRef\":{\"name\":\"dummySecret\",\"namespace\":\"dummy\"},\"serviceAccounts\":[{\"email\":\"mcmDummy@dummy.com\",\"scopes\":[\"https://www.googleapis.com/auth/compute\"]}],\"region\":\"europe-dummy\",\"zone\":\"europe-dummy\"}")
	gcpProviderSpecInvalidPostZone := []byte("{\"canIpForward\":true,\"deletionProtection\":false,\"description\":\"Machine created to test out-of-tree gcp mcm driver.\",\"disks\":[{\"autoDelete\":true,\"boot\":true,\"sizeGb\":50,\"type\":\"pd-standard\",\"image\":\"projects/coreos-cloud/global/images/coreos-stable-2135-6-0-v20190801\",\"labels\":{\"name\":\"test-mc-gcp\"}}],\"labels\":{\"name\":\"test-mc-gcp\"},\"machineType\":\"n1-standard-2\",\"metadata\":[{\"key\":\"gcp\",\"value\":\"my-value\"}],\"networkInterfaces\":[{\"network\":\"dummyShoot\",\"subnetwork\":\"dummyShoot\"}],\"scheduling\":{\"automaticRestart\":true,\"onHostMaintenance\":\"MIGRATE\",\"preemptible\":false},\"secretRef\":{\"name\":\"dummySecret\",\"namespace\":\"dummy\"},\"serviceAccounts\":[{\"email\":\"mcmDummy@dummy.com\",\"scopes\":[\"https://www.googleapis.com/auth/compute\"]}],\"tags\":[\"kubernetes-io-cluster-dummy-machine\",\"kubernetes-io-role-mcm\",\"dummy-machine\"],\"region\":\"europe-dummy\",\"zone\":\"invalid post\"}")
	gcpProviderSpecInvalidListZone := []byte("{\"canIpForward\":true,\"deletionProtection\":false,\"description\":\"Machine created to test out-of-tree gcp mcm driver.\",\"disks\":[{\"autoDelete\":true,\"boot\":true,\"sizeGb\":50,\"type\":\"pd-standard\",\"image\":\"projects/coreos-cloud/global/images/coreos-stable-2135-6-0-v20190801\",\"labels\":{\"name\":\"test-mc-gcp\"}}],\"labels\":{\"name\":\"test-mc-gcp\"},\"machineType\":\"n1-standard-2\",\"metadata\":[{\"key\":\"gcp\",\"value\":\"my-value\"}],\"networkInterfaces\":[{\"network\":\"dummyShoot\",\"subnetwork\":\"dummyShoot\"}],\"scheduling\":{\"automaticRestart\":true,\"onHostMaintenance\":\"MIGRATE\",\"preemptible\":false},\"secretRef\":{\"name\":\"dummySecret\",\"namespace\":\"dummy\"},\"serviceAccounts\":[{\"email\":\"mcmDummy@dummy.com\",\"scopes\":[\"https://www.googleapis.com/auth/compute\"]}],\"tags\":[\"kubernetes-io-cluster-dummy-machine\",\"kubernetes-io-role-mcm\",\"dummy-machine\"],\"region\":\"europe-dummy\",\"zone\":\"invalid list\"}")
	gcpPVSpec := &corev1.PersistentVolumeSpec{
		PersistentVolumeSource: corev1.PersistentVolumeSource{
			GCEPersistentDisk: &corev1.GCEPersistentDiskVolumeSource{
				PDName: "dummyPD",
			},
		},
	}
	gcpPVSpecEmptyPD := &corev1.PersistentVolumeSpec{
		PersistentVolumeSource: corev1.PersistentVolumeSource{},
	}
	gcpPVResponse := []string{"dummyPD"}

	gcpProviderSecret := map[string][]byte{
		"userData":           []byte("dummy-data"),
		"serviceAccountJSON": []byte("{\"type\":\"service_account\",\"project_id\":\"sap-se-gcp-scp-k8s-dev\"}"),
	}

	gcpProviderSecretWithMisssingUserData := map[string][]byte{
		//"userData":           []byte(""),
		"serviceAccountJSON": []byte("{\"type\":\"service_account\",\"project_id\":\"sap-se-gcp-scp-k8s-dev\"}"),
	}
	gcpProviderSecretWithoutProjectID := map[string][]byte{
		"userData":           []byte("dummy-data"),
		"serviceAccountJSON": []byte("{\"type\":\"service_account\",\"project_id\":10}"),
	}

	var ms *MachinePlugin
	var mockPluginSPIImpl *fake.PluginSPIImpl

	var _ = BeforeSuite(func() {
		//Start a mock server to listen to mock client requests
		//This is rquired as compute sdk doesn't offer any interface so the mocking is done via a mock http client pass to the compute service
		go fake.NewMockServer()
		mockPluginSPIImpl = &fake.PluginSPIImpl{Client: &http.Client{}}
		ms = NewGCPPlugin(mockPluginSPIImpl)
	})

	var _ = BeforeEach(func() {
		// Reinitialise instances
		fake.Instances = nil
	})

	Describe("#Create Machine", func() {
		type setup struct {
		}
		type action struct {
			machineRequest *driver.CreateMachineRequest
		}
		type expect struct {
			machineResponse   *driver.CreateMachineResponse
			errToHaveOccurred bool
			errMessage        string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("###table",
			func(data *data) {

				ctx := context.Background()
				response, err := ms.CreateMachine(ctx, data.action.machineRequest)
				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(data.expect.machineResponse.ProviderID).To(Equal(response.ProviderID))
					Expect(data.expect.machineResponse.NodeName).To(Equal(response.NodeName))
				}
			},

			Entry("Creata a simple machine", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					machineResponse: &driver.CreateMachineResponse{
						ProviderID: "gce:///sap-se-gcp-scp-k8s-dev/europe-dummy/dummy-machine",
						NodeName:   "dummy-machine",
					},
					errToHaveOccurred: false,
				},
			}),
			Entry("With zone missing", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpecValidationErr),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        FailAtSpecValidation,
				},
			}),
			Entry("With no provider spec", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass([]byte("")),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        CreateFailAtJSONUnmarshalling,
				},
			}),

			Entry("With no secrets", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(make(map[string][]byte)),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        FailAtNoSecretsPassed,
				},
			}),
			Entry("With secrets missing user data", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecretWithMisssingUserData),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        FailAtSecretsWithNoUserData,
				},
			}),
			Entry("With secrets having invalid projectID value", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecretWithoutProjectID),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        FailAtInvalidProjectID,
				},
			}),
			Entry("With Post for invalid zone", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpecInvalidPostZone),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        FailAtInvalidZonePostCall,
				},
			}),
			Entry("With invalid list zone", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpecInvalidListZone),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        CreateFailAtInvalidZoneListCall,
				},
			}),
		)
	})
	Describe("#Delete Machine", func() {
		type setup struct {
		}
		type action struct {
			machineRequest *driver.DeleteMachineRequest
		}
		type expect struct {
			machineResponse   *driver.DeleteMachineResponse
			errToHaveOccurred bool
			errMessage        string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("###table",
			func(data *data) {

				ctx := context.Background()
				_, err := ms.DeleteMachine(ctx, data.action.machineRequest)
				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},
			Entry("Delete a non-existent machine", &data{
				action: action{
					machineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine("non-existent-dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					machineResponse:   &driver.DeleteMachineResponse{},
					errToHaveOccurred: true,
					errMessage:        FailAtNotFound,
				},
			}),

			Entry("With no provider spec", &data{
				action: action{
					machineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass([]byte("")),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        DeleteFailAtJSONUnmarshalling,
				},
			}),

			Entry("With invalid list zone", &data{
				action: action{
					machineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpecInvalidListZone),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        DeleteFailAtInvalidZoneListCall,
				},
			}),
		)
	})
	Describe("#List Machines", func() {
		type setup struct {
		}
		type action struct {
			createMachine bool
			createRequest *driver.CreateMachineRequest
			listRequest   *driver.ListMachinesRequest
		}
		type expect struct {
			createResponse          *driver.CreateMachineResponse
			listResponse            *driver.ListMachinesResponse
			errToHaveOccurred       bool
			listErrToHaveOccurred   bool
			createErrToHaveOccurred bool
			machineCount            int
			errMessage              string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}

		var listResponse *driver.ListMachinesResponse
		var createResponse *driver.CreateMachineResponse
		var listErr error
		var createErr error
		DescribeTable("###table",
			func(data *data) {

				ctx := context.Background()
				createErr = nil
				listErr = nil
				if data.action.createMachine {
					createResponse, createErr = ms.CreateMachine(ctx, data.action.createRequest)
				}
				listResponse, listErr = ms.ListMachines(ctx, data.action.listRequest)
				if data.expect.errToHaveOccurred {
					if data.expect.createErrToHaveOccurred {
						Expect(createErr).To(HaveOccurred())
						Expect(createErr.Error()).To(Equal(data.expect.errMessage))
					}
					if data.expect.listErrToHaveOccurred {
						Expect(listErr).To(HaveOccurred())
						Expect(listErr.Error()).To(Equal(data.expect.errMessage))
					}
				} else {
					Expect(createErr).ToNot(HaveOccurred())
					Expect(listErr).ToNot(HaveOccurred())
					if data.action.createMachine {
						Expect(data.expect.createResponse.ProviderID).To(Equal(createResponse.ProviderID))
					}
					Expect(data.expect.machineCount).To(Equal(len(listResponse.MachineList)))
				}

			},
			Entry("Create and List a simple machine", &data{
				action: action{
					createRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecret),
					},
					createMachine: true,
					listRequest: &driver.ListMachinesRequest{
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					createResponse: &driver.CreateMachineResponse{
						ProviderID: "gce:///sap-se-gcp-scp-k8s-dev/europe-dummy/dummy-machine",
						NodeName:   "dummy-machine",
					},
					errToHaveOccurred: false,
					machineCount:      1,
				},
			}),
			Entry("List with no provider spec", &data{
				action: action{
					createMachine: false,
					listRequest: &driver.ListMachinesRequest{
						MachineClass: newGCPMachineClass([]byte("")),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred:     true,
					listErrToHaveOccurred: true,
					errMessage:            ListFailAtJSONUnmarshalling,
				},
			}),
			Entry("List with Get call for invalid zone", &data{
				action: action{
					createMachine: false,
					listRequest: &driver.ListMachinesRequest{
						MachineClass: newGCPMachineClass(gcpProviderSpecInvalidListZone),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred:     true,
					listErrToHaveOccurred: true,
					errMessage:            ListFailAtInvalidZoneListCall,
				},
			}),
		)

	})
	Describe("#Get MachineStatus", func() {
		type setup struct {
		}
		type action struct {
			createMachine    bool
			createRequest    *driver.CreateMachineRequest
			getStatusRequest *driver.GetMachineStatusRequest
		}
		type expect struct {
			createResponse             *driver.CreateMachineResponse
			getStatusResponse          *driver.GetMachineStatusResponse
			errToHaveOccurred          bool
			createErrToHaveOccurred    bool
			getStatusErrToHaveOccurred bool
			errMessage                 string
			machineCount               int
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}
		var getStatusResponse *driver.GetMachineStatusResponse
		var createResponse *driver.CreateMachineResponse
		var getStatusErr error
		var createErr error
		DescribeTable("###table",
			func(data *data) {

				ctx := context.Background()
				createErr = nil
				getStatusErr = nil

				if data.action.createMachine {
					createResponse, createErr = ms.CreateMachine(ctx, data.action.createRequest)
					getStatusResponse, getStatusErr = ms.GetMachineStatus(ctx, data.action.getStatusRequest)
				} else {

					getStatusResponse, getStatusErr = ms.GetMachineStatus(ctx, data.action.getStatusRequest)
				}
				if data.expect.errToHaveOccurred {
					if data.expect.createErrToHaveOccurred {
						Expect(createErr).To(HaveOccurred())
						Expect(createErr.Error()).To(Equal(data.expect.errMessage))
					}
					if data.expect.getStatusErrToHaveOccurred {
						Expect(getStatusErr).To(HaveOccurred())
						Expect(getStatusErr.Error()).To(Equal(data.expect.errMessage))
					}
				} else {
					Expect(createErr).ToNot(HaveOccurred())
					Expect(getStatusErr).ToNot(HaveOccurred())
					if data.action.createMachine {
						Expect(data.expect.createResponse.ProviderID).To(Equal(createResponse.ProviderID))
					}
					Expect(data.expect.getStatusResponse).To(Equal(getStatusResponse))
				}

			},
			Entry("Get status with non-existent machine name", &data{
				action: action{

					getStatusRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine("non-existent-dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpecNoTagsToSearch),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					machineCount:      0,
					errMessage:        FailAtNonExistingMachine,
				},
			}),
			Entry("Create and Get a simple machine", &data{
				action: action{
					createRequest: &driver.CreateMachineRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecret),
					},
					createMachine: true,
					getStatusRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpec),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					createResponse: &driver.CreateMachineResponse{
						ProviderID: "gce:///sap-se-gcp-scp-k8s-dev/europe-dummy/dummy-machine",
						NodeName:   "dummy-machine",
					},
					errToHaveOccurred: false,
					machineCount:      1,
					getStatusResponse: &driver.GetMachineStatusResponse{
						ProviderID: "gce:///sap-se-gcp-scp-k8s-dev/europe-dummy/dummy-machine",
						NodeName:   "dummy-machine",
					},
				},
			}),
			Entry("Get status with no provider spec", &data{
				action: action{
					createMachine: false,
					getStatusRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass([]byte("")),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred:          true,
					getStatusErrToHaveOccurred: true,
					errMessage:                 FailAtJSONUnmarshalling,
				},
			}),
			Entry("Get status with call for invalid zone", &data{
				action: action{
					createMachine: false,
					getStatusRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine("dummy-machine"),
						MachineClass: newGCPMachineClass(gcpProviderSpecInvalidListZone),
						Secret:       newSecret(gcpProviderSecret),
					},
				},
				expect: expect{
					errToHaveOccurred:          true,
					getStatusErrToHaveOccurred: true,
					errMessage:                 FailAtInvalidZoneListCall,
				},
			}),
		)
	})
	Describe("#Get VolumeIDs", func() {
		type setup struct {
		}
		type action struct {
			machineRequest *driver.GetVolumeIDsRequest
		}
		type expect struct {
			machineResponse   *driver.GetVolumeIDsResponse
			errToHaveOccurred bool
			errMessage        string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("###table",
			func(data *data) {

				ctx := context.Background()
				resp, err := ms.GetVolumeIDs(ctx, data.action.machineRequest)
				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(resp.VolumeIDs)).To(Equal(len(data.expect.machineResponse.VolumeIDs)))
					for i, r := range resp.VolumeIDs {
						Expect(r).To(Equal(data.expect.machineResponse.VolumeIDs[i]))
					}
				}

			},

			Entry("With valid PV list", &data{
				action: action{
					machineRequest: &driver.GetVolumeIDsRequest{
						PVSpecs: []*corev1.PersistentVolumeSpec{gcpPVSpec},
					},
				},
				expect: expect{
					machineResponse: &driver.GetVolumeIDsResponse{
						VolumeIDs: gcpPVResponse,
					},
					errToHaveOccurred: false,
					errMessage:        FailAtMethodNotImplemented,
				},
			}),
			Entry("With emtpy PV list", &data{
				action: action{
					machineRequest: &driver.GetVolumeIDsRequest{
						PVSpecs: []*corev1.PersistentVolumeSpec{gcpPVSpecEmptyPD},
					},
				},
				expect: expect{
					machineResponse: &driver.GetVolumeIDsResponse{
						VolumeIDs: []string{},
					},
					errToHaveOccurred: false,
				},
			}),
		)
	})

})

func newMachine(name string) *v1alpha1.Machine {
	return &v1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newGCPMachineClass(gcpProviderSpec []byte) *v1alpha1.MachineClass {
	return &v1alpha1.MachineClass{
		ProviderSpec: runtime.RawExtension{
			Raw: gcpProviderSpec,
		},
	}
}

func newSecret(gcpProviderSecretRaw map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		Data: gcpProviderSecretRaw,
	}
}
