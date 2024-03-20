// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package errors

import "fmt"

// MachineNotFoundError is used to indicate not found error in PluginSPI
type MachineNotFoundError struct {
	// Name is the machine name
	Name string
	// MachineID is the machine uuid
	MachineID string
}

func (e *MachineNotFoundError) Error() string {
	return fmt.Sprintf("machine name=%s, uuid=%s not found", e.Name, e.MachineID)
}

// MachineResourceExhaustedError is used to indicate resource exhausted error in PluginSPI
type MachineResourceExhaustedError struct {
	// error msg of error classified as resource exhausted
	Msg string
}

func (e *MachineResourceExhaustedError) Error() string {
	return e.Msg
}
