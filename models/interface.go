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

// Interface model
type Interface struct {
	Name          string
	Type          IfType
	Subinterfaces map[string]*Subinterface
	lock          sync.RWMutex
}

// NewInterface New Interface model.
func NewInterface() *Interface {
	return &Interface{
		Name:          "",
		Type:          IfTypeUnknown,
		Subinterfaces: map[string]*Subinterface{},
	}
}

// IsValid Reports whether Instance represents a valid value.
func (iface *Interface) IsValid() bool {
	iface.lock.RLock()
	defer iface.lock.RUnlock()

	// ignore tunnel interface
	if iface.Type == IfTypeTunnel {
		return true
	}

	if len(iface.Subinterfaces) >= 0 {
		for _, subiface := range iface.Subinterfaces {
			if subiface.IsValid() == false {
				return false
			}
		}
		return true
	}

	return false
}

// Copy Copy Instance model.
func (iface *Interface) Copy() *Interface {
	iface.lock.RLock()
	defer iface.lock.RUnlock()

	subinterfaces := map[string]*Subinterface{}
	for _, subiface := range iface.Subinterfaces {
		subinterfaces[subiface.Name] = subiface.Copy()
	}

	return &Interface{
		Name:          iface.Name,
		Type:          iface.Type,
		Subinterfaces: subinterfaces,
	}
}

// SetType Set interface type.
func (iface *Interface) SetType(ifType string) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	iface.Type = toIfType(ifType)
}

// DeleteType Set interface type.
func (iface *Interface) DeleteType() {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	iface.Type = IfTypeUnknown
}

// AddSubinterface Add interface.
func (iface *Interface) AddSubinterface(subifname string) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	_, ret := iface.Subinterfaces[subifname]
	if ret == false {
		subiface := NewSubinterface()
		subiface.Name = subifname
		iface.Subinterfaces[subifname] = subiface
	}
}

// DeleteSubinterface Delete interface.
func (iface *Interface) DeleteSubinterface(subifname string) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	delete(iface.Subinterfaces, subifname)
}

// SetSubifIndex Set index.
func (iface *Interface) SetSubifIndex(subifname string, index uint64) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetIndex(index)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].SetIndex(index)
	}
}

// DeleteSubifIndex Delete index.
func (iface *Interface) DeleteSubifIndex(subifname string) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.DeleteIndex()
	}
}

// SetSubifIP Set IP.
func (iface *Interface) SetSubifIP(subifname string, ip net.IP) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetIP(ip)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].SetIP(ip)
	}
}

// DeleteSubifIP Delete IP.
func (iface *Interface) DeleteSubifIP(subifname string) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.DeleteIP()
	}
}

// SetSubifPrefix Set prefix.
func (iface *Interface) SetSubifPrefix(subifname string, prefix uint32) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetPrefix(prefix)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].SetPrefix(prefix)
	}
}

// DeleteSubifPrefix Delete prefix.
func (iface *Interface) DeleteSubifPrefix(subifname string) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.DeletePrefix()
	}
}

// AddVrrp Add VRRP.
func (iface *Interface) AddVrrp(subifname string, vrid uint8) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.AddVrrp(vrid)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].AddVrrp(vrid)
	}
}

// DeleteVrrp Add VRRP.
func (iface *Interface) DeleteVrrp(subifname string, vrid uint8) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.DeleteVrrp(vrid)
	}
}

// SetVrrpPriority Set priority.
func (iface *Interface) SetVrrpPriority(subifname string, vrid uint8, priority uint8) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetVrrpPriority(vrid, priority)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].SetVrrpPriority(vrid, priority)
	}
}

// SetDefaultVrrpPriority Set default priority.
func (iface *Interface) SetDefaultVrrpPriority(subifname string, vrid uint8) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetDefaultVrrpPriority(vrid)
	}
}

// SetVrrpPreempt Set preempt.
func (iface *Interface) SetVrrpPreempt(subifname string, vrid uint8, preempt bool) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetVrrpPreempt(vrid, preempt)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].SetVrrpPreempt(vrid, preempt)
	}
}

// SetDefaultVrrpPreempt Set default preempt.
func (iface *Interface) SetDefaultVrrpPreempt(subifname string, vrid uint8) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetDefaultVrrpPreempt(vrid)
	}
}

// SetVrrpInterval Set interval.
func (iface *Interface) SetVrrpInterval(subifname string, vrid uint8, interval uint16) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetVrrpInterval(vrid, interval)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].SetVrrpInterval(vrid, interval)
	}
}

// SetDefaultVrrpInterval Set default interval.
func (iface *Interface) SetDefaultVrrpInterval(subifname string, vrid uint8) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.SetDefaultVrrpInterval(vrid)
	}
}

// AddVrrpVirtualAddress Add virtual address.
func (iface *Interface) AddVrrpVirtualAddress(subifname string, vrid uint8, addr net.IP) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.AddVrrpVirtualAddress(vrid, addr)
	} else {
		iface.AddSubinterface(subifname)
		iface.Subinterfaces[subifname].AddVrrpVirtualAddress(vrid, addr)
	}
}

// DeleteVrrpVirtualAddress Delete virtual address.
func (iface *Interface) DeleteVrrpVirtualAddress(subifname string, vrid uint8, addr net.IP) {
	iface.lock.Lock()
	defer iface.lock.Unlock()

	subiface, ret := iface.Subinterfaces[subifname]
	if ret {
		subiface.DeleteVrrpVirtualAddress(vrid, addr)
	}
}

// String Returns a string representation of the Instance model.
func (iface *Interface) String() string {
	iface.lock.RLock()
	defer iface.lock.RUnlock()

	var str string
	str = fmt.Sprintf("Name: %s", iface.Name)
	str = fmt.Sprintf("Type: %d", iface.Type)
	for _, subiface := range iface.Subinterfaces {
		str = fmt.Sprintf("%s, Interfaces(%s): {%s}", str, subiface.Name, subiface.String())
	}

	return str
}
