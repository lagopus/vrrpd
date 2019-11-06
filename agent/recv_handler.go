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

	"github.com/lagopus/vrrpd/rpc"
	log "github.com/sirupsen/logrus"
)

const (
	// RecvHandlerModuleName RecvHandler module name.
	RecvHandlerModuleName = "RecvHandlerModule"
)

// RecvHandler handler
type RecvHandler struct {
	handlerChannel chan *rpc.BulkPackets
	stopChannel    chan bool
	isRunning      bool
	wg             *sync.WaitGroup
	lock           sync.Mutex
}

// NewRecvHandler New RecvHandler module.
func NewRecvHandler(wg *sync.WaitGroup) *RecvHandler {
	h := &RecvHandler{
		handlerChannel: make(chan *rpc.BulkPackets, rpc.RecvChannelSize),
		stopChannel:    make(chan bool),
		wg:             wg,
	}
	return h
}

func (h *RecvHandler) handlerLoop() {
	defer h.wg.Done()

	for {
		select {
		case packets := <-h.handlerChannel:
			vmgr.RecvVRRPAdv(packets)
		case <-h.stopChannel:
			log.Infof("Stop handlerLoop.")
			return
		}
	}
}

// SendHandlerChannel Send event to handlerChannel.
func (h *RecvHandler) SendHandlerChannel(bufs *rpc.BulkPackets) {
	h.handlerChannel <- bufs
}

// Start Start hander.
func (h *RecvHandler) Start() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.isRunning == false {
		h.wg.Add(1)
		go h.handlerLoop()
		h.isRunning = true
	}

	return nil
}

// Stop Stop hander.
func (h *RecvHandler) Stop() {
	h.lock.Lock()
	defer h.lock.Unlock()

	if h.isRunning == true {
		h.stopChannel <- true
		h.isRunning = false
	}
}

// Resume Resume module.
func (h *RecvHandler) Resume() error {
	// implement if necessary.
	return nil
}

// Suspend Suspend module.
func (h *RecvHandler) Suspend() error {
	// implement if necessary.
	return nil
}

// Name Module name.
func (h *RecvHandler) Name() string {
	return RecvHandlerModuleName
}
