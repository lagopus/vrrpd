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

package config

import (
	"fmt"
	"net"
	"sync"

	"github.com/lagopus/vrrpd/models"
)

// AgentConfig agent config.
type AgentConfig struct {
	DsAddr     net.IP
	DsPort     uint16
	DpaAddr    net.IP
	DpaPort    uint16
	HostifAddr net.IP
	HostifPort uint16
	Interfaces map[string]*models.Interface
	lock       sync.RWMutex
}

func newAgentConfig() *AgentConfig {
	return &AgentConfig{
		DsAddr:     net.ParseIP("127.0.0.1"),
		DsPort:     2650,
		DpaAddr:    net.ParseIP("127.0.0.1"),
		DpaPort:    30010,
		HostifAddr: net.ParseIP("127.0.0.1"),
		HostifPort: 30020,
		Interfaces: map[string]*models.Interface{},
	}
}

// IsValid Reports whether AgentConfig represents a valid value.
func (agentConfig *AgentConfig) IsValid() bool {
	agentConfig.lock.RLock()
	defer agentConfig.lock.RUnlock()

	if agentConfig.DsAddr != nil && agentConfig.DsPort > 0 &&
		agentConfig.DpaAddr != nil && agentConfig.DpaPort > 0 &&
		agentConfig.HostifAddr != nil && agentConfig.HostifPort > 0 &&
		len(agentConfig.Interfaces) >= 0 {
		for _, iface := range agentConfig.Interfaces {
			if iface.IsValid() == false {
				return false
			}
		}
		return true
	}

	return false
}

// Copy Copy Agent config.
func (agentConfig *AgentConfig) Copy() *AgentConfig {
	agentConfig.lock.RLock()
	defer agentConfig.lock.RUnlock()

	ifaces := map[string]*models.Interface{}
	for _, iface := range agentConfig.Interfaces {
		ifaces[iface.Name] = iface.Copy()
	}

	return &AgentConfig{
		DsAddr:     agentConfig.DsAddr,
		DsPort:     agentConfig.DsPort,
		DpaAddr:    agentConfig.DpaAddr,
		DpaPort:    agentConfig.DpaPort,
		HostifAddr: agentConfig.HostifAddr,
		HostifPort: agentConfig.HostifPort,
		Interfaces: ifaces,
	}
}

// AddInterface Add interface.
func (agentConfig *AgentConfig) AddInterface(ifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	_, ret := agentConfig.Interfaces[ifname]
	if ret == false {
		iface := models.NewInterface()
		iface.Name = ifname
		agentConfig.Interfaces[ifname] = iface
	}
}

// DeleteInterface Delete interface.
func (agentConfig *AgentConfig) DeleteInterface(ifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	delete(agentConfig.Interfaces, ifname)
}

// SetInterfaceType Add interface type.
func (agentConfig *AgentConfig) SetInterfaceType(ifname string, iftype string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetType(iftype)
	} else {
		iface := models.NewInterface()
		iface.Name = ifname
		iface.SetType(iftype)
		agentConfig.Interfaces[ifname] = iface
	}
}

// DeleteInterfaceType Add interface type.
func (agentConfig *AgentConfig) DeleteInterfaceType(ifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.DeleteType()
	}
}

// AddSubinterface Add subinterface.
func (agentConfig *AgentConfig) AddSubinterface(ifname string, subifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.AddSubinterface(subifname)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].AddSubinterface(subifname)
	}
}

// DeleteSubinterface Delete subinterface.
func (agentConfig *AgentConfig) DeleteSubinterface(ifname string, subifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.DeleteSubinterface(subifname)
	}
}

// SetSubifIndex Set subinterface index.
func (agentConfig *AgentConfig) SetSubifIndex(ifname string, subifname string, index uint64) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetSubifIndex(subifname, index)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].SetSubifIndex(subifname, index)
	}
}

// DeleteSubifIndex Delete subinterface index.
func (agentConfig *AgentConfig) DeleteSubifIndex(ifname string, subifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.DeleteSubifPrefix(subifname)
	}
}

// SetSubifIP Set subinterface IP.
func (agentConfig *AgentConfig) SetSubifIP(ifname string, subifname string, ip net.IP) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetSubifIP(subifname, ip)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].SetSubifIP(subifname, ip)
	}
}

// DeleteSubifIP Delete subinterface IP.
func (agentConfig *AgentConfig) DeleteSubifIP(ifname string, subifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.DeleteSubifIP(subifname)
	}
}

// SetSubifPrefix Set subinterface prefix.
func (agentConfig *AgentConfig) SetSubifPrefix(ifname string, subifname string, prefix uint32) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetSubifPrefix(subifname, prefix)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].SetSubifPrefix(subifname, prefix)
	}
}

// DeleteSubifPrefix Delete subinterface prefix.
func (agentConfig *AgentConfig) DeleteSubifPrefix(ifname string, subifname string) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.DeleteSubifPrefix(subifname)
	}
}

// AddVrrp Add VRRP.
func (agentConfig *AgentConfig) AddVrrp(ifname string, subifname string, vrid uint8) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.AddVrrp(subifname, vrid)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].AddVrrp(subifname, vrid)
	}
}

// DeleteVrrp Delete VRRP.
func (agentConfig *AgentConfig) DeleteVrrp(ifname string, subifname string, vrid uint8) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.DeleteVrrp(subifname, vrid)
	}
}

// SetVrrpPriority Set priority.
func (agentConfig *AgentConfig) SetVrrpPriority(ifname string, subifname string,
	vrid uint8, priority uint8) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetVrrpPriority(subifname, vrid, priority)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].SetVrrpPriority(subifname, vrid, priority)
	}
}

// SetDefaultVrrpPriority SetDefault priority.
func (agentConfig *AgentConfig) SetDefaultVrrpPriority(ifname string, subifname string,
	vrid uint8) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetDefaultVrrpPriority(subifname, vrid)
	}
}

// SetVrrpPreempt Set preempt.
func (agentConfig *AgentConfig) SetVrrpPreempt(ifname string, subifname string,
	vrid uint8, preempt bool) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetVrrpPreempt(subifname, vrid, preempt)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].SetVrrpPreempt(subifname, vrid, preempt)
	}
}

// SetDefaultVrrpPreempt SetDefault preempt.
func (agentConfig *AgentConfig) SetDefaultVrrpPreempt(ifname string, subifname string,
	vrid uint8) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetDefaultVrrpPreempt(subifname, vrid)
	}
}

// SetVrrpInterval Set interval.
func (agentConfig *AgentConfig) SetVrrpInterval(ifname string, subifname string,
	vrid uint8, interval uint16) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetVrrpInterval(subifname, vrid, interval)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].SetVrrpInterval(subifname, vrid, interval)
	}
}

// SetDefaultVrrpInterval SetDefault interval.
func (agentConfig *AgentConfig) SetDefaultVrrpInterval(ifname string, subifname string,
	vrid uint8) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.SetDefaultVrrpInterval(subifname, vrid)
	}
}

// AddVrrpVirtualAddress Add VRRP VirtualAddress.
func (agentConfig *AgentConfig) AddVrrpVirtualAddress(ifname string, subifname string,
	vrid uint8, addr net.IP) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.AddVrrpVirtualAddress(subifname, vrid, addr)
	} else {
		agentConfig.AddInterface(ifname)
		agentConfig.Interfaces[ifname].AddVrrpVirtualAddress(subifname, vrid, addr)
	}
}

// DeleteVrrpVirtualAddress Delete VRRP VirtualAddress.
func (agentConfig *AgentConfig) DeleteVrrpVirtualAddress(ifname string, subifname string,
	vrid uint8, addr net.IP) {
	agentConfig.lock.Lock()
	defer agentConfig.lock.Unlock()

	iface, ret := agentConfig.Interfaces[ifname]
	if ret {
		iface.DeleteVrrpVirtualAddress(subifname, vrid, addr)
	}
}

// String Returns a string representation of the Agent config.
func (agentConfig *AgentConfig) String() string {
	agentConfig.lock.RLock()
	defer agentConfig.lock.RUnlock()

	var str string
	str = fmt.Sprintf("DsAddr: %s", agentConfig.DsAddr.String())
	str = fmt.Sprintf("%s, DsPort: %d", str, agentConfig.DsPort)
	str = fmt.Sprintf("%s, DpaAddr: %s", str, agentConfig.DpaAddr.String())
	str = fmt.Sprintf("%s, DpaPort: %d", str, agentConfig.DpaPort)
	str = fmt.Sprintf("%s, HostifAddr: %s", str, agentConfig.HostifAddr.String())
	str = fmt.Sprintf("%s, HostifPort: %d", str, agentConfig.HostifPort)
	for _, iface := range agentConfig.Interfaces {
		str = fmt.Sprintf("%s, Instances(%s): {%s}", str, iface.Name, iface.String())
	}

	return str
}
