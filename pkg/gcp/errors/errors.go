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
