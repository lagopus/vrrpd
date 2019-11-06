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

	log "github.com/sirupsen/logrus"
)

const (
	// MDownTimerModuleName MDownTimer module name.
	MDownTimerModuleName = "MDownTimerModule"
)

// MDownTimer Master down timer.
type MDownTimer struct {
	backupTable map[string]*VRRP
	stopChannel chan bool
	isRunning   bool
	wg          *sync.WaitGroup
	lock        sync.Mutex
}

// NewMDownTimer New MDownTimer module.
func NewMDownTimer(wg *sync.WaitGroup) *MDownTimer {
	mdt := &MDownTimer{
		backupTable: map[string]*VRRP{},
		stopChannel: make(chan bool),
		wg:          wg,
	}
	return mdt
}

// Master down timer loop.
func (mdt *MDownTimer) mDownTimerLoop() {
	defer mdt.wg.Done()

	ticker := time.NewTicker(MinInterval)
	for {
		select {
		case <-ticker.C:
			now := time.Now()

			mdt.lock.Lock()
			for _, v := range mdt.backupTable {
				f := func() { mdt.deleteBackupTable(v) }
				v.masterDownTimeExpired(now, f)
			}
			mdt.lock.Unlock()
		case <-mdt.stopChannel:
			log.Infof("Stop MDownTimerLoop.")
			ticker.Stop()
			return
		}
	}
}

func (mdt *MDownTimer) deleteBackupTable(v *VRRP) {
	delete(mdt.backupTable, v.objID)
}

// Start Start master down timer.
func (mdt *MDownTimer) Start() error {
	mdt.lock.Lock()
	defer mdt.lock.Unlock()

	if mdt.isRunning == false {
		mdt.wg.Add(1)
		go mdt.mDownTimerLoop()
		mdt.isRunning = true
	}

	return nil
}

// Stop Stop master down timer.
func (mdt *MDownTimer) Stop() {
	mdt.lock.Lock()
	defer mdt.lock.Unlock()

	if mdt.isRunning == true {
		mdt.stopChannel <- true
		mdt.isRunning = false
	}
}

// Resume Resume module.
func (mdt *MDownTimer) Resume() error {
	// implement if necessary.
	return nil
}

// Suspend Suspend module.
func (mdt *MDownTimer) Suspend() error {
	// implement if necessary.
	return nil
}

// Name Module name.
func (mdt *MDownTimer) Name() string {
	return MDownTimerModuleName
}

// AddBackupTable Add entry in BackupTable.
func (mdt *MDownTimer) AddBackupTable(v *VRRP) {
	mdt.lock.Lock()
	defer mdt.lock.Unlock()
	mdt.backupTable[v.objID] = v
}

// DeleteBackupTable Delete entry in BackupTable.
func (mdt *MDownTimer) DeleteBackupTable(v *VRRP) {
	mdt.lock.Lock()
	defer mdt.lock.Unlock()

	mdt.deleteBackupTable(v)
}
