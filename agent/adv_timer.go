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
	"sync"
	"time"

	"github.com/lagopus/vrrpd/rpc"
	log "github.com/sirupsen/logrus"
)

const (
	// AdvTimerModuleName AdvTimer module name.
	AdvTimerModuleName = "AdvTimerModule"
)

// AdvTimer Advertisement interval timer.
type AdvTimer struct {
	masterTable map[string]*VRRP
	stopChannel chan bool
	hostif      *rpc.Hostif
	isRunning   bool
	wg          *sync.WaitGroup
	lock        sync.Mutex
}

// NewAdvTimer New AdvTimer module.
func NewAdvTimer(hostif *rpc.Hostif, wg *sync.WaitGroup) *AdvTimer {
	at := &AdvTimer{
		masterTable: map[string]*VRRP{},
		stopChannel: make(chan bool),
		hostif:      hostif,
		wg:          wg,
	}
	return at
}

// send advertisement.
func (at *AdvTimer) sendVRRPAdv(bps *rpc.BulkPackets) {
	log.Debugf("send VRRP.")
	at.hostif.PacketoutBulk(bps)
}

// event.
func (at *AdvTimer) timeOutEvent() {
	at.lock.Lock()
	defer at.lock.Unlock()

	packets := []*rpc.Packet{}
	now := time.Now()
	for _, v := range at.masterTable {
		if ps, ok := v.getVRRPAdvExpired(now); ok {
			packets = append(packets, ps...)
		}
	}

	if len(packets) != 0 {
		at.sendVRRPAdv(rpc.NewBulkPackets(packets))
	}
}

// Advertisement interval timer loop.
func (at *AdvTimer) advTimerLoop() {
	defer at.wg.Done()

	ticker := time.NewTicker(MinInterval)
	for {
		select {
		case <-ticker.C:
			at.timeOutEvent()
		case <-at.stopChannel:
			log.Infof("Stop advTimerLoop.")
			ticker.Stop()
			return
		}
	}
}

// Start Start sdvertisemen interval timer.
func (at *AdvTimer) Start() error {
	at.lock.Lock()
	defer at.lock.Unlock()

	if at.isRunning == false {
		at.wg.Add(1)
		go at.advTimerLoop()
		at.isRunning = true
	}

	return nil
}

// Stop Stop sdvertisemen interval timer.
func (at *AdvTimer) Stop() {
	at.lock.Lock()
	defer at.lock.Unlock()

	if at.isRunning == true {
		at.stopChannel <- true
		at.isRunning = false
	}
}

// Resume Resume module.
func (at *AdvTimer) Resume() error {
	// implement if necessary.
	return nil
}

// Suspend Suspend module.
func (at *AdvTimer) Suspend() error {
	// implement if necessary.
	return nil
}

// Name Module name.
func (at *AdvTimer) Name() string {
	return AdvTimerModuleName
}

// AddMasterTable Add entry in MasterTable.
func (at *AdvTimer) AddMasterTable(v *VRRP) {
	at.lock.Lock()
	defer at.lock.Unlock()
	at.masterTable[v.objID] = v
}

// DeleteMasterTable Delete entry in MasterTable.
func (at *AdvTimer) DeleteMasterTable(v *VRRP) {
	at.lock.Lock()
	defer at.lock.Unlock()
	delete(at.masterTable, v.objID)
}
