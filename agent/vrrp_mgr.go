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

// The following was in reference.
// - https://github.com/google/seesaw

package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/lagopus/vrrpd/models"
	"github.com/lagopus/vrrpd/module"
	"github.com/lagopus/vrrpd/packets"
	"github.com/lagopus/vrrpd/rpc"
	log "github.com/sirupsen/logrus"
)

// VRRPMgr VRRP manager.
type VRRPMgr struct {
	vrrpTable map[string]*VRRP
	lock      sync.RWMutex
}

var vmgr = newVRRPMgr()

func newVRRPMgr() *VRRPMgr {
	vm := &VRRPMgr{
		vrrpTable: map[string]*VRRP{},
	}
	return vm
}

// RecvVRRPAdv Recv VRRP Advertisement.
func (vmgr *VRRPMgr) RecvVRRPAdv(bps *rpc.BulkPackets) {
	vmgr.lock.RLock()
	defer vmgr.lock.RUnlock()

	for _, packet := range bps.Packets {
		now := time.Now()
		if _, ip, vrrpAdv, err := packets.DecodeVRRPAdv(packet.Data); err == nil {
			objID := fmt.Sprintf("%s:%d", packet.Subifname, vrrpAdv.VirtualRtrID)
			if v, ok := vmgr.vrrpTable[objID]; ok {
				v.NextStateForRecv(vrrpAdv, ip.SrcIP, now)
			} else {
				log.Errorf("Unknown vrrp: %s", objID)
				continue
			}
		} else {
			log.Errorf("Bad packet.")
			continue
		}
	}
}

// UpdateSettings Update settings.
func (vmgr *VRRPMgr) UpdateSettings(subifTable map[string]*models.Subinterface) error {
	vmgr.lock.Lock()
	defer vmgr.lock.Unlock()

	log.Debugf("Update settings")

	// delete all vrrp
	for _, v := range vmgr.vrrpTable {
		v.NextState(EventShutdown)
		delete(vmgr.vrrpTable, v.objID)
		log.Debugf("Delete VRRP: %v", v)
	}

	if module.GetState() == module.StateSuspended {
		if err := module.ResumeModules(); err != nil {
			return err
		}
	}

	for _, subifModel := range subifTable {
		if subifModel.IsValid() {
			for _, vrrpModel := range subifModel.VRRPs {
				if v, err := newVRRP(subifModel, vrrpModel); err == nil {
					v.NextState(EventStart)
					vmgr.vrrpTable[v.objID] = v
					log.Debugf("Create VRRP: %v", v)
				} else {
					return err
				}
			}
		} else {
			log.Errorf("Create VRRP failed: %s", subifModel)
		}
	}

	if len(vmgr.vrrpTable) == 0 {
		if err := module.SuspendModules(); err != nil {
			return err
		}
	}

	return nil
}
