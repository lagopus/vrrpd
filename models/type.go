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

package models

import (
	"strings"
)

// IfType Type of interface.
type IfType uint8

const (
	IfTypeUnknown IfType = iota
	IfTypeEthernetCsmacd
	IfTypeTunnel
)

// VRRP
const (
	// DefaultPriority Default priority.
	DefaultPriority = 100
	// DefaultPreempt Default preempt.
	DefaultPreempt = true
	// DefaultAccept Default accept.
	DefaultAccept = false
	// DefaultInterval Default interval.
	DefaultInterval = 100
)

func toIfType(str string) IfType {
	switch strings.ToLower(str) {
	case "ethernetcsmacd":
		return IfTypeEthernetCsmacd
	case "tunnel":
		return IfTypeTunnel
	default:
		return IfTypeUnknown
	}
}
