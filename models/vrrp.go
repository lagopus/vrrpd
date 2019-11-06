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
	"fmt"
	"net"
	"sync"
)

// VRRP model
type VRRP struct {
	Vrid             uint8
	Priority         uint8
	Preempt          bool
	Accept           bool
	Interval         uint16
	VirtualAddresses []net.IP
	lock             sync.RWMutex
}

// NewVRRP New VRRP model.
func NewVRRP() *VRRP {
	return &VRRP{
		Vrid:             0,
		Priority:         DefaultPriority,
		Preempt:          DefaultPreempt,
		Accept:           DefaultAccept,
		Interval:         DefaultInterval,
		VirtualAddresses: []net.IP{},
	}
}

// IsValid Reports whether VRRP represents a valid value.
func (vrrp *VRRP) IsValid() bool {
	vrrp.lock.RLock()
	defer vrrp.lock.RUnlock()

	if vrrp.Vrid > 0 && len(vrrp.VirtualAddresses) > 0 {
		return true
	}

	return false
}

// Copy Copy VRRP model.
func (vrrp *VRRP) Copy() *VRRP {
	vrrp.lock.RLock()
	defer vrrp.lock.RUnlock()

	vas := make([]net.IP, len(vrrp.VirtualAddresses))
	copy(vas, vrrp.VirtualAddresses)

	return &VRRP{
		Vrid:             vrrp.Vrid,
		Priority:         vrrp.Priority,
		Preempt:          vrrp.Preempt,
		Accept:           vrrp.Accept,
		Interval:         vrrp.Interval,
		VirtualAddresses: vas,
	}
}

// SetPriority Set priority.
func (vrrp *VRRP) SetPriority(priority uint8) {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	vrrp.Priority = priority
}

// SetDefaultPriority Set default priority.
func (vrrp *VRRP) SetDefaultPriority() {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	vrrp.Priority = DefaultPriority
}

// SetPreempt Set preempt.
func (vrrp *VRRP) SetPreempt(preempt bool) {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	vrrp.Preempt = preempt
}

// SetDefaultPreempt Set default preempt.
func (vrrp *VRRP) SetDefaultPreempt() {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	vrrp.Preempt = DefaultPreempt
}

// SetInterval Set interval.
func (vrrp *VRRP) SetInterval(interval uint16) {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	vrrp.Interval = interval
}

// SetDefaultInterval Set default interval.
func (vrrp *VRRP) SetDefaultInterval() {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	vrrp.Interval = DefaultInterval
}

// AddVirtualAddress Add virtual address.
func (vrrp *VRRP) AddVirtualAddress(addr net.IP) {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	for _, vaddr := range vrrp.VirtualAddresses {
		if vaddr.Equal(addr) {
			return
		}
	}

	vrrp.VirtualAddresses = append(vrrp.VirtualAddresses, addr)
}

// DeleteVirtualAddress Delete virtual address.
func (vrrp *VRRP) DeleteVirtualAddress(addr net.IP) {
	vrrp.lock.Lock()
	defer vrrp.lock.Unlock()

	tmp := []net.IP{}

	// delete address if vaddr and addr match
	for _, vaddr := range vrrp.VirtualAddresses {
		if !vaddr.Equal(addr) {
			tmp = append(tmp, vaddr)
		}
	}

	vrrp.VirtualAddresses = tmp
}

// IsMaster Report whether VRRP is Master.
func (vrrp *VRRP) IsMaster(addr net.IP) bool {
	vrrp.lock.RLock()
	defer vrrp.lock.RUnlock()

	for _, vaddr := range vrrp.VirtualAddresses {
		if vaddr.Equal(addr) {
			return true
		}
	}
	return false
}

// String Returns a string representation of the VRRP model.
func (vrrp *VRRP) String() string {
	vrrp.lock.RLock()
	defer vrrp.lock.RUnlock()

	var str string
	str = fmt.Sprintf("Vrid: %d", vrrp.Vrid)
	str = fmt.Sprintf("%s, Priority: %d", str, vrrp.Priority)
	str = fmt.Sprintf("%s, Preempt: %t", str, vrrp.Preempt)
	str = fmt.Sprintf("%s, Accept: %t", str, vrrp.Accept)
	str = fmt.Sprintf("%s, Interval: %d", str, vrrp.Interval)
	str = fmt.Sprintf("%s, VirtualAddresses: %v", str, vrrp.VirtualAddresses)

	return str
}
