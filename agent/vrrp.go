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
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/lagopus/vrrpd/models"
	"github.com/lagopus/vrrpd/module"
	"github.com/lagopus/vrrpd/packets"
	"github.com/lagopus/vrrpd/packets/layers"
	"github.com/lagopus/vrrpd/rpc"
	log "github.com/sirupsen/logrus"
)

// VRRP structure of VRRPv3.
type VRRP struct {
	objID string
	layers.VRRPv3Adv
	nextMasterAdvTime      time.Time
	masterDownInterval     uint16
	skewTime               uint16
	nextDownTime           time.Time
	preempt                bool
	accept                 bool
	state                  VRRPState
	vaddrs                 []net.IP
	vmac                   net.HardwareAddr
	subifName              string
	subifIP                net.IP
	subifPrefix            uint32
	advPackets             []*rpc.Packet
	advPriorityZeroPackets []*rpc.Packet
	garpPackets            []*rpc.Packet
	advTimer               *AdvTimer
	mDownTimer             *MDownTimer
	hostif                 *rpc.Hostif
	dpagent                *rpc.DPAgent
	// performance-oriented (channel is not used).
	lock sync.Mutex
}

// create VRRP.
func newVRRP(imodel *models.Subinterface, vmodel *models.VRRP) (*VRRP, error) {
	var priority uint8
	if vmodel.IsMaster(imodel.IP) {
		priority = 255
	} else {
		priority = vmodel.Priority
	}

	log.Debugf("vrrp adv priority: %d", priority)

	v := &VRRP{
		VRRPv3Adv: layers.VRRPv3Adv{
			Version:      layers.VRRPv3Version,
			Type:         layers.VRRPv3Advertisement,
			VirtualRtrID: vmodel.Vrid,
			Priority:     priority,
			MaxAdverInt:  vmodel.Interval,
			IPAddress:    vmodel.VirtualAddresses, // TODO: sort
		},
		preempt:     vmodel.Preempt,
		vaddrs:      vmodel.VirtualAddresses,
		subifName:   imodel.Name,
		subifIP:     imodel.IP,
		subifPrefix: imodel.Prefix,
		advTimer:    (module.GetModule(AdvTimerModuleName)).(*AdvTimer),
		mDownTimer:  (module.GetModule(MDownTimerModuleName)).(*MDownTimer),
		hostif:      (module.GetModule(rpc.HostifModuleName)).(*rpc.Hostif),
		dpagent:     (module.GetModule(rpc.DPAgentModuleName)).(*rpc.DPAgent),
	}
	v.objID = fmt.Sprintf("%s:%d", imodel.Name, vmodel.Vrid)
	v.setStateNoLock(StateInitialize)
	v.resetMasterDownInterval(vmodel.Interval)
	now := time.Now()
	v.nextMasterAdvTime = now
	v.setNextDownTimeNoLock(now, v.masterDownInterval)

	// get mac address from DataPlane
	var err error
	if v.vmac, err = v.dpagent.GetVifMacaddr(v.subifName); err != nil {
		log.Errorf("GetVifMacaddr faild: %v", err)
		return nil, err
	}

	if err = v.resetPacket(); err != nil {
		log.Errorf("resetPacket faild: %v", err)
		return nil, err
	}

	return v, nil
}

func (v *VRRP) setState(s VRRPState) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.setStateNoLock(s)
}

func (v *VRRP) setStateNoLock(s VRRPState) {
	v.state = s
}

func (v *VRRP) getStateNoLock() VRRPState {
	return v.state
}

func (v *VRRP) getState() VRRPState {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.getStateNoLock()
}

func (v *VRRP) setNextMasterAdvTime(t time.Time) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.setNextMasterAdvTimeNoLock(t)
}

func (v *VRRP) setNextMasterAdvTimeNoLock(t time.Time) {
	v.nextMasterAdvTime = t.Add(time.Duration(v.MaxAdverInt) * Centisecond)
}

func (v *VRRP) setNextDownTime(t time.Time, interval uint16) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.setNextDownTimeNoLock(t, interval)
}

func (v *VRRP) setNextDownTimeNoLock(t time.Time, interval uint16) {
	v.nextDownTime = t.Add(time.Duration(interval) * Centisecond)
}

func (v *VRRP) getNextDownTime() time.Time {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.getNextDownTimeNoLock()
}

func (v *VRRP) getNextDownTimeNoLock() time.Time {
	return v.nextDownTime
}

func (v *VRRP) getSkewTimeTime() uint16 {
	v.lock.Lock()
	defer v.lock.Unlock()
	return v.getSkewTimeTimeNoLock()
}

func (v *VRRP) getSkewTimeTimeNoLock() uint16 {
	return v.skewTime
}

func (v *VRRP) sendVRRPAdvPriorityZero() {
	log.Debugf("send VRRPPriorityZero.")
	bps := rpc.NewBulkPackets(v.advPriorityZeroPackets)
	v.hostif.PacketoutBulk(bps)
}

func (v *VRRP) sendGARP() {
	log.Debugf("send GARP.")
	bps := rpc.NewBulkPackets(v.garpPackets)
	v.hostif.PacketoutBulk(bps)
}

func (v *VRRP) toMaster() {
	log.Debugf("set virtual addresses.")

	addrs := []string{}
	for _, vaddr := range v.vaddrs {
		addr := fmt.Sprintf("%s/%d", vaddr.String(), v.subifPrefix)
		addrs = append(addrs, addr)
	}
	phyaddr := fmt.Sprintf("%s/%d", v.subifIP.String(), v.subifPrefix)

	if err := v.dpagent.ToMaster(v.subifName, phyaddr, addrs); err != nil {
		// ignore.
		log.Errorf("%v", err)
	}
}

func (v *VRRP) toBackup() {
	log.Debugf("unset virtual addresses.")

	addrs := []string{}
	for _, vaddr := range v.vaddrs {
		addr := fmt.Sprintf("%s/%d", vaddr.String(), v.subifPrefix)
		addrs = append(addrs, addr)
	}
	phyaddr := fmt.Sprintf("%s/%d", v.subifIP.String(), v.subifPrefix)

	if err := v.dpagent.ToBackup(v.subifName, phyaddr, addrs); err != nil {
		// ignore.
		log.Errorf("%v", err)
	}
}

func (v *VRRP) createGARP() ([]*rpc.Packet, error) {
	ps := []*rpc.Packet{}
	for _, ip := range v.IPAddress {
		// virtual mac address
		//if buf, err := packets.SerializeVirtualMacARP(v.VirtualRtrID, ip); err == nil {
		// physical mac address
		if buf, err := packets.SerializeARP(ip, v.vmac); err == nil {
			p := rpc.NewPacket(v.subifName, buf)
			ps = append(ps, p)
		} else {
			return nil, err
		}
	}
	return ps, nil
}

func (v *VRRP) createVRRPAdv() ([]*rpc.Packet, error) {
	ps := []*rpc.Packet{}
	if buf, err := packets.SerializeVRRPAdv(v.subifIP, &v.VRRPv3Adv); err == nil {
		p := rpc.NewPacket(v.subifName, buf)
		ps = append(ps, p)
	} else {
		return nil, err
	}
	return ps, nil
}

func (v *VRRP) createVRRPAdvPriorityZero() ([]*rpc.Packet, error) {
	cv := *v
	cv.Priority = 0

	return cv.createVRRPAdv()
}

func (v *VRRP) resetPacket() error {
	v.lock.Lock()
	defer v.lock.Unlock()

	var err error
	if v.advPackets, err = v.createVRRPAdv(); err != nil {
		return err
	}
	if v.advPriorityZeroPackets, err = v.createVRRPAdvPriorityZero(); err != nil {
		return err
	}
	if v.garpPackets, err = v.createGARP(); err != nil {
		return err
	}
	return nil
}

func (v *VRRP) resetMasterDownInterval(interval uint16) {
	skewTime := ((256 - uint16(v.Priority)) * interval) / 256
	masterDownInterval := (3 * interval) + skewTime

	if masterDownInterval != v.masterDownInterval ||
		skewTime != v.skewTime {
		v.masterDownInterval = masterDownInterval
		v.skewTime = skewTime
		log.Infof("set skewTime=%v, masterDownInterval=%v",
			skewTime, masterDownInterval)
	}
}

func (v *VRRP) getVRRPAdvExpired(now time.Time) ([]*rpc.Packet, bool) {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.nextMasterAdvTime.UnixNano() > now.UnixNano() {
		return nil, false
	}
	v.setNextMasterAdvTimeNoLock(now)

	return v.advPackets, true
}

func (v *VRRP) masterDownTimeExpired(now time.Time, funcDeleteBackupTable func()) {
	v.lock.Lock()
	defer v.lock.Unlock()

	if v.getNextDownTimeNoLock().UnixNano() <= now.UnixNano() {
		// delete Backuptable in mDownTimer.
		funcDeleteBackupTable()
		if v.getStateNoLock() == StateBackup {
			v.nextStateNoLock(EventMasterDown)
		}
	}
}

func (v *VRRP) containsInterfaceIPs(ips []net.IP) bool {
	for _, ip := range ips {
		if bytes.Equal(v.subifIP, ip) {
			return true
		}
	}

	return false
}

// Initialize

func (v *VRRP) doInitializeTasks() {
	// virtualIP == local router IP.
	if v.containsInterfaceIPs(v.IPAddress) {
		v.nextStateNoLock(EventStartMaster)
		return
	}

	if v.Priority == 255 {
		v.nextStateNoLock(EventStartMaster)
		return
	}

	v.nextStateNoLock(EventStartBackup)
}

func (v *VRRP) becomeInitialize() {
	log.Info("Become Initialize.")

	switch s := v.getStateNoLock(); s {
	case StateInitialize:
		// nothing
	case StateBackup:
		v.mDownTimer.DeleteBackupTable(v)
	case StateMaster:
		v.advTimer.DeleteMasterTable(v)
		v.sendVRRPAdvPriorityZero()
		v.toBackup()
	default:
		log.Errorf("Bad state %v.", s)
	}

	v.setStateNoLock(StateInitialize)
}

// Master.

func (v *VRRP) becomeMaster() {
	log.Info("Become Master.")

	// to master
	v.toMaster()
	// send GARP
	v.sendGARP()
	// not called DeleteBackupTable() (called in mDownTimer).
	v.advTimer.AddMasterTable(v)
	v.setStateNoLock(StateMaster)
}

// Backup.

func (v *VRRP) becomeBackup() {
	log.Info("Become Backup.")
	v.advTimer.DeleteMasterTable(v)
	v.mDownTimer.AddBackupTable(v)
	v.setStateNoLock(StateBackup)
	// to backup
	v.toBackup()
}

// State Machine.
//
//     +---+   [EventStart]
//     | . |------------------+
//     +---+                  |
//                            V
//     [EventShutdown]    +-----------------+     [EventShutdown]
//            +---------->|                 |<----------+
//            |           | StateInitialize |           |
//            |   +-------|                 |-------+   |
//            |   |       +-----------------+       |   |
//            |   |[EventStartMaster]               |   |
//            |   |                                 |   |
//            |   |               [EventStartBackup]|   |
//            |   V                                 V   |
//    +---------------+ [EventDetectedNewMaster] +---------------+
//    |               |------------------------->|               |
//    |  StateMaster  |                          |  StateBackup  |
//    |               |<-------------------------|               |
//    +---------------+    [EventMasterDown]     +---------------+
//                         [EventPreempt]
//

func (v *VRRP) nextStateNoLock(e VRRPEvent) {
	switch s := v.getStateNoLock(); s {
	case StateInitialize:
		switch e {
		case EventStart:
			v.doInitializeTasks()
		case EventStartMaster:
			v.becomeMaster()
		case EventStartBackup:
			v.becomeBackup()
		default:
			log.Errorf("Bad event %v in StateInitialize", e)
		}
	case StateMaster:
		switch e {
		case EventDetectedNewMaster:
			v.becomeBackup()
		case EventShutdown:
			v.becomeInitialize()
		default:
			log.Errorf("Bad event %v in StateMaster", e)
		}
	case StateBackup:
		switch e {
		case EventMasterDown:
			v.becomeMaster()
		case EventPreempt:
			v.becomeMaster()
		case EventShutdown:
			v.becomeInitialize()
		default:
			log.Errorf("Bad event %v in StateBackup", e)
		}
	default:
		log.Errorf("Bad state %v", s)
	}
}

// NextState Next state.
func (v *VRRP) NextState(e VRRPEvent) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.nextStateNoLock(e)
}

// NextStateForRecv  Next state for recv.
func (v *VRRP) NextStateForRecv(vrrpAdv *layers.VRRPv3Adv,
	advSrcIP net.IP,
	now time.Time) {
	v.lock.Lock()
	defer v.lock.Unlock()

	// check virtualIP in adv == local router IP.
	if v.containsInterfaceIPs(vrrpAdv.IPAddress) {
		log.Debugf("Discard adv: %v", vrrpAdv)
		return
	}

	switch s := v.getStateNoLock(); s {
	case StateInitialize:
		// do nothing.
	case StateBackup:
		switch {
		case v.Priority == 0:
			v.setNextDownTimeNoLock(now, v.getSkewTimeTimeNoLock())
		case v.preempt == true &&
			vrrpAdv.Priority < v.Priority:
			log.Debugf("Event = %v: adv.Priority = %v, Priority = %v",
				EventPreempt, vrrpAdv.Priority, v.Priority)
			v.nextStateNoLock(EventPreempt)
		case v.preempt == false ||
			vrrpAdv.Priority >= v.Priority:
			v.MaxAdverInt = vrrpAdv.MaxAdverInt
			v.resetMasterDownInterval(v.MaxAdverInt)
			v.setNextDownTimeNoLock(now, v.masterDownInterval)
		default:
			log.Debugf("Discard adv: %v", vrrpAdv)
		}
	case StateMaster:
		switch {
		case v.Priority == 0:
			v.nextMasterAdvTime = now
		case vrrpAdv.Priority > v.Priority ||
			(vrrpAdv.Priority == v.Priority &&
				bytes.Compare(advSrcIP, v.subifIP) > 0):
			log.Debugf("Event = %v: adv.Priority = %v, Priority = %v"+
				"adv.srcIP = %v, subifIP = %v",
				EventDetectedNewMaster,
				vrrpAdv.Priority, v.Priority,
				vrrpAdv.Priority, v.subifIP)
			v.advTimer.DeleteMasterTable(v)
			v.MaxAdverInt = vrrpAdv.MaxAdverInt
			v.resetMasterDownInterval(v.MaxAdverInt)
			v.setNextDownTimeNoLock(now, v.masterDownInterval)
			v.nextStateNoLock(EventDetectedNewMaster)
		default:
			log.Debugf("Discard adv: %v", vrrpAdv)
		}
	default:
		log.Errorf("Bad state %v.", s)
	}
}
