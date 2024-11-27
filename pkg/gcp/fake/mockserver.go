// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"

	compute "google.golang.org/api/compute/v1"
)

// Instances stores and manages the instances during create,delete and list calls
var Instances []*compute.Instance

var singleConnHandler = make(chan struct{})

type httpHandler struct {
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//opType := decodeOperationType(r)
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "POST":
		handleCreate(w, r)
	case "GET":
		handleList(w, r)
	case "DELETE":
		handleDelete(w, r)
	}
}

// NewMockServer creates an http server to mock the gcp compute api
func NewMockServer() {

	var srv = http.Server{ // #nosec  G112 (CWE-400) -- Only used for testing
		Addr:    ":6666",
		Handler: new(httpHandler),
	}
	//http.HandleFunc("/", handler)
	go handleShutdown(&srv)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Printf("Failed to shutdown server %v", err)
	}
	<-singleConnHandler
}

func handleShutdown(srv *http.Server) {
	shutDownSignal := make(chan os.Signal, 1)
	signal.Notify(shutDownSignal, os.Interrupt)

	<-shutDownSignal
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server Shutdown: %v", err)
	}
	close(singleConnHandler)
}

/*
1. Extract the method from the Request to delegat to the correct handler
2. Extract the instance name which has the test sceanrio for which response needs to be formed
3. Create a response object according to the call
4. Prepare response.body with the json structure expected by the googe-api-go-client for compute
*/

func decodeOperationType(r *http.Request, index int) string {
	opTypes := strings.Split(r.URL.Path, "/")

	return opTypes[len(opTypes)-index]
}

func handleCreate(w http.ResponseWriter, r *http.Request) {

	if decodeOperationType(r, 2) == "invalid post" {
		http.Error(w, "Invalid post zone", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error in reading request body", err)
	}
	var instance *compute.Instance
	err = json.Unmarshal(body, &instance)
	if err != nil {
		fmt.Println("Error in unmarshalling request body", err)
	}

	operation := compute.Operation{
		Status:        "RUNNING",
		OperationType: "insert",
		Kind:          "compute#operation",
	}

	Instances = append(Instances, instance)
	_ = json.NewEncoder(w).Encode(operation)
}

func handleList(w http.ResponseWriter, r *http.Request) {
	//error mock handling for create/delete calls
	if decodeOperationType(r, 3) == "invalid list" {
		http.Error(w, "Invalid list zone", http.StatusBadRequest)
		return
	}
	//error mock handling for listMachines call
	if decodeOperationType(r, 2) == "invalid list" {
		http.Error(w, "Invalid list zone", http.StatusBadRequest)
		return
	}

	// operation call is made for wait loop to let the create/delete operation to complete
	if decodeOperationType(r, 2) == "operations" {
		operation := compute.Operation{
			Status:        "DONE",
			OperationType: "insert",
			Kind:          "compute#operation",
		}
		_ = json.NewEncoder(w).Encode(operation)
	} else { // this is the regular list call handling for VM

		instances := compute.InstanceList{
			Items: Instances,
		}
		_ = json.NewEncoder(w).Encode(instances)

	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if decodeOperationType(r, 3) == "invalid post" {
		http.Error(w, "Invalid post zone", http.StatusBadRequest)
		return
	}

	if decodeOperationType(r, 1) == "reset-machine-count" {
		Instances = nil
	}

	operation := compute.Operation{
		Status:        "RUNNING",
		OperationType: "delete",
		Kind:          "compute#operation",
	}
	_ = json.NewEncoder(w).Encode(operation)
}
