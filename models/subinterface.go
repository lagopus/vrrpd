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
	"math"
	"net"
	"sync"
)

// Subinterface model
type Subinterface struct {
	Name   string
	Index  uint64
	IP     net.IP
	Prefix uint32
	MAC    net.HardwareAddr
	VRRPs  map[uint8]*VRRP
	lock   sync.RWMutex
}

// NewSubinterface New Subinterface model.
func NewSubinterface() *Subinterface {
	return &Subinterface{
		Name:   "",
		Index:  math.MaxUint64,
		IP:     nil,
		Prefix: math.MaxUint32,
		MAC:    nil,
		VRRPs:  map[uint8]*VRRP{},
	}
}

// IsValid Reports whether Subinterface represents a valid value.
func (subif *Subinterface) IsValid() bool {
	subif.lock.RLock()
	defer subif.lock.RUnlock()

	if len(subif.Name) > 0 && subif.IP != nil &&
		subif.Index >= 0 && subif.Index <= math.MaxUint32 &&
		subif.Prefix >= 0 && subif.Prefix <= 32 &&
		len(subif.VRRPs) >= 0 {
		for _, vrrp := range subif.VRRPs {
			if vrrp.IsValid() == false {
				return false
			}
		}
		return true
	}

	return false
}

// Copy Copy Subinterface model.
func (subif *Subinterface) Copy() *Subinterface {
	subif.lock.RLock()
	defer subif.lock.RUnlock()

	mac := make(net.HardwareAddr, len(subif.MAC))
	copy(mac, subif.MAC)

	vrrps := map[uint8]*VRRP{}
	for _, vrrp := range subif.VRRPs {
		vrrps[vrrp.Vrid] = vrrp.Copy()
	}

	return &Subinterface{
		Name:   subif.Name,
		Index:  subif.Index,
		IP:     dupIP(subif.IP),
		Prefix: subif.Prefix,
		MAC:    mac,
		VRRPs:  vrrps,
	}
}

// SetIndex Set Index.
func (subif *Subinterface) SetIndex(index uint64) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	subif.Index = index
}

// DeleteIndex Delete Index.
func (subif *Subinterface) DeleteIndex() {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	subif.Index = math.MaxUint64
}

// SetIP Set IP.
func (subif *Subinterface) SetIP(ip net.IP) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	subif.IP = ip
}

// DeleteIP Delete IP.
func (subif *Subinterface) DeleteIP() {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	subif.IP = nil
}

// SetPrefix Set prefix.
func (subif *Subinterface) SetPrefix(prefix uint32) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	subif.Prefix = prefix
}

// DeletePrefix Delete prefix.
func (subif *Subinterface) DeletePrefix() {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	subif.Prefix = math.MaxUint32
}

// AddVrrp Add VRRP.
func (subif *Subinterface) AddVrrp(vrid uint8) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	_, ret := subif.VRRPs[vrid]
	if ret == false {
		vrrp := NewVRRP()
		vrrp.Vrid = vrid
		subif.VRRPs[vrid] = vrrp
	}
}

// DeleteVrrp Delete VRRP.
func (subif *Subinterface) DeleteVrrp(vrid uint8) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	delete(subif.VRRPs, vrid)
}

// SetVrrpPriority Set VRRP priority.
func (subif *Subinterface) SetVrrpPriority(vrid uint8, priority uint8) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.SetPriority(priority)
	} else {
		subif.AddVrrp(vrid)
		subif.VRRPs[vrid].SetPriority(priority)
	}
}

// SetDefaultVrrpPriority Set default VRRP priority.
func (subif *Subinterface) SetDefaultVrrpPriority(vrid uint8) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.SetDefaultPriority()
	}
}

// SetVrrpPreempt Set VRRP preempt.
func (subif *Subinterface) SetVrrpPreempt(vrid uint8, preempt bool) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.SetPreempt(preempt)
	} else {
		subif.AddVrrp(vrid)
		subif.VRRPs[vrid].SetPreempt(preempt)
	}
}

// SetDefaultVrrpPreempt Set default VRRP preempt.
func (subif *Subinterface) SetDefaultVrrpPreempt(vrid uint8) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.SetDefaultPreempt()
	}
}

// SetVrrpInterval Set VRRP interval.
func (subif *Subinterface) SetVrrpInterval(vrid uint8, interval uint16) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.SetInterval(interval)
	} else {
		subif.AddVrrp(vrid)
		subif.VRRPs[vrid].SetInterval(interval)
	}
}

// SetDefaultVrrpInterval Set default VRRP interval.
func (subif *Subinterface) SetDefaultVrrpInterval(vrid uint8) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.SetDefaultInterval()
	}
}

// AddVrrpVirtualAddress Add VRRP virtual address.
func (subif *Subinterface) AddVrrpVirtualAddress(vrid uint8, addr net.IP) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.AddVirtualAddress(addr)
	} else {
		subif.AddVrrp(vrid)
		subif.VRRPs[vrid].AddVirtualAddress(addr)
	}
}

// DeleteVrrpVirtualAddress Delete VRRP virtual address.
func (subif *Subinterface) DeleteVrrpVirtualAddress(vrid uint8, addr net.IP) {
	subif.lock.Lock()
	defer subif.lock.Unlock()

	vrrp, ret := subif.VRRPs[vrid]
	if ret {
		vrrp.DeleteVirtualAddress(addr)
	}
}

// String Returns a string representation of the Interface model.
func (subif *Subinterface) String() string {
	subif.lock.RLock()
	defer subif.lock.RUnlock()

	var str string
	str = fmt.Sprintf("Name: %s", subif.Name)
	str = fmt.Sprintf("%s, Index: %d", str, subif.Index)
	str = fmt.Sprintf("%s, IP: %s", str, subif.IP.String())
	str = fmt.Sprintf("%s, Prefix: %d", str, subif.Prefix)
	str = fmt.Sprintf("%s, MAC: %s", str, subif.MAC.String())
	for _, vrrp := range subif.VRRPs {
		str = fmt.Sprintf("%s, VRRPs(%d): {%s}", str, vrrp.Vrid, vrrp.String())
	}

	return str
}
