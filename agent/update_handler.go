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
	"os"
	"sync"

	"github.com/lagopus/vrrpd/config"
	"github.com/lagopus/vrrpd/models"
	log "github.com/sirupsen/logrus"
)

const (
	// UpdateHandlerModuleName UpdateHandler module name.
	UpdateHandlerModuleName = "UpdateHandlerModule"

	// UpdateChannelSize Size of UpdateChannel.
	UpdateChannelSize = 1000
)

// UpdateHandler handler
type UpdateHandler struct {
	handlerChannel chan *config.AgentConfig
	stopChannel    chan bool
	isRunning      bool
	wg             *sync.WaitGroup
	lock           sync.Mutex
}

// NewUpdateHandler New UpdateHandler module.
func NewUpdateHandler(wg *sync.WaitGroup) *UpdateHandler {
	h := &UpdateHandler{
		handlerChannel: make(chan *config.AgentConfig, UpdateChannelSize),
		stopChannel:    make(chan bool),
		wg:             wg,
	}
	return h
}

func (u *UpdateHandler) handlerLoop() {
	defer u.wg.Done()

	for {
		select {
		case conf := <-u.handlerChannel:
			// overwrite same interface
			subifTable := map[string]*models.Subinterface{}
			for _, iface := range conf.Interfaces {
				// ignore tunnel interface
				if iface.Type != models.IfTypeTunnel {
					for _, subiface := range iface.Subinterfaces {
						subifTable[subiface.Name] = subiface
					}
				}
			}
			if err := vmgr.UpdateSettings(subifTable); err != nil {
				// TODO: graceful shutdown
				log.Errorf("UpdateSettings failure: %v", err)
				os.Exit(1)
			}
		case <-u.stopChannel:
			log.Infof("Stop handlerLoop.")
			return
		}
	}
}

// SendHandlerChannel Send event to handlerChannel.
func (u *UpdateHandler) SendHandlerChannel(conf *config.AgentConfig) {
	u.handlerChannel <- conf
}

// Start Start hander.
func (u *UpdateHandler) Start() error {
	u.lock.Lock()
	defer u.lock.Unlock()

	if u.isRunning == false {
		u.wg.Add(1)
		go u.handlerLoop()
		u.isRunning = true
	}

	return nil
}

// Stop Stop hander.
func (u *UpdateHandler) Stop() {
	u.lock.Lock()
	defer u.lock.Unlock()

	if u.isRunning == true {
		u.stopChannel <- true
		u.isRunning = false
	}
}

// Resume Resume module.
func (u *UpdateHandler) Resume() error {
	// implement if necessary.
	return nil
}

// Suspend Suspend module.
func (u *UpdateHandler) Suspend() error {
	// implement if necessary.
	return nil
}

// Name Module name.
func (u *UpdateHandler) Name() string {
	return UpdateHandlerModuleName
}
