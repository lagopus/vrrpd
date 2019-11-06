//
// Copyright 2017 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package agent

import (
	"time"
)

// VRRPState Type of state.
type VRRPState uint8

const (
	// StateInitialize Initialize.
	StateInitialize VRRPState = iota
	// StateBackup Backup.
	StateBackup
	// StateMaster Master.
	StateMaster
)

func (s VRRPState) String() string {
	var str string
	switch s {
	case StateInitialize:
		str = "Initialize"
	case StateBackup:
		str = "Backup"
	case StateMaster:
		str = "Master"
	default:
		str = "UNKNOWN"
	}
	return str
}

// VRRPEvent Type of event
type VRRPEvent uint8

const (
	// EventStart Start.
	EventStart VRRPEvent = iota
	// EventStartMaster Master.
	EventStartMaster
	// EventStartBackup Backup.
	EventStartBackup
	// EventMasterDown MasterDown.
	EventMasterDown
	// EventDetectedNewMaster DetectedNewMaster.
	EventDetectedNewMaster
	// EventPreempt Preempt mode.
	EventPreempt
	// EventShutdown Shutdown.
	EventShutdown
)

func (e VRRPEvent) String() string {
	var str string
	switch e {
	case EventStart:
		str = "Start"
	case EventStartMaster:
		str = "StartMaster"
	case EventStartBackup:
		str = "StartBackup"
	case EventMasterDown:
		str = "MasterDown"
	case EventDetectedNewMaster:
		str = "DetectedNewMaster"
	case EventPreempt:
		str = "Preempt"
	case EventShutdown:
		str = "Shutdown"
	default:
		str = "UNKNOWN"
	}
	return str
}

const (
	// Centisecond Centisecond for time.Duration.
	Centisecond = time.Millisecond * 10
	// MinInterval 1 centiseconds.
	MinInterval time.Duration = time.Duration(1) * Centisecond
)
