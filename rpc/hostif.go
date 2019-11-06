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

package rpc

import (
	"fmt"
	"sync"
	"time"

	"github.com/lagopus/vrrpd/module"
	"github.com/lagopus/vsw/modules/hostif/packets_io"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type entry struct {
	data interface{}
}

type recvCallbackType func(bps *BulkPackets)

// Hostif Hostif RPC.
type Hostif struct {
	conn                     *Connection
	client                   packets_io.PacketsIoClient
	sendChannel              chan *entry
	sendStopChannel          chan bool
	sendCancelFuncForConnect context.CancelFunc
	sendCancelFuncForSend    context.CancelFunc
	recvStopChannel          chan bool
	recvCancelFunc           context.CancelFunc
	recvCallback             recvCallbackType
	state                    module.State
	globalWg                 *sync.WaitGroup
	localWg                  *sync.WaitGroup
	lock                     sync.Mutex
}

// NewHostif New hostif instance.
func NewHostif(addr string, port int, recvCallback recvCallbackType,
		globalWg *sync.WaitGroup) *Hostif {
	r := &Hostif{
		conn:                     NewConnection(addr, port),
		client:                   nil,
		sendChannel:              make(chan *entry, SendChannelSize),
		sendStopChannel:          make(chan bool),
		sendCancelFuncForConnect: nil,
		sendCancelFuncForSend:    nil,
		recvStopChannel:          make(chan bool),
		recvCancelFunc:           nil,
		recvCallback:             recvCallback,
		state:                    module.StateInitialize,
		globalWg:                 globalWg,
		localWg:                  new(sync.WaitGroup),
	}

	return r
}

func (h *Hostif) sendGRPC(bps *packets_io.BulkPackets) error {
	log.Debugf("Send packet: %v", bps)

	ctx, cancel := context.WithCancel(context.Background())
	h.sendCancelFuncForSend = cancel

	opts := []grpc.CallOption{}

	if _, err := h.client.SendBulk(ctx, bps, opts...); err != nil {
		return err
	}

	return nil
}

func (h *Hostif) sendLoop() {
	defer h.localWg.Done()

	for {
		select {
		case entry := <-h.sendChannel:
			switch pkt := entry.data.(type) {
			case *packets_io.BulkPackets:
				if err := h.sendGRPC(pkt); err != nil {
					log.Warnf("Can't send packets: %v", err)
				}
			default:
				log.Errorf("Not found type.")
			}
		case <-h.sendStopChannel:
			log.Infof("Stop hostif sendLoop.")
			return
		}
	}
}

// PacketoutBulk PacketoutBulk.
func (h *Hostif) PacketoutBulk(bps *BulkPackets) {
	entry := &entry{
		data: bps.bulkpackets,
	}

	h.sendChannel <- entry
}

func (h *Hostif) recvGRPC() (*packets_io.BulkPackets, error) {
	ctx, cancel := context.WithCancel(context.Background())
	h.recvCancelFunc = cancel

	opts := []grpc.CallOption{}

	var bps *packets_io.BulkPackets
	var err error
	if bps, err = h.client.RecvBulk(ctx, &packets_io.Null{}, opts...); err != nil {
		return nil, err
	}

	return bps, nil
}

func (h *Hostif) recvLoop() {
	defer h.localWg.Done()

	ticker := time.NewTicker(RecvInterval)
	for {
		select {
		case <-ticker.C:
			if bps, err := h.recvGRPC(); err != nil {
				continue
			} else {
				if bps.N != 0 {
					h.recvCallback(newBulkPackets(bps))
				}
			}
		case <-h.recvStopChannel:
			log.Infof("Stop hostif recvLoop.")
			ticker.Stop()
			return
		}
	}
}

func (h *Hostif) startLoopNoLock() {
	log.Debugf("Start loop.")

	// start send
	h.localWg.Add(1)
	go h.sendLoop()

	// start receive
	h.localWg.Add(1)
	go h.recvLoop()

	h.globalWg.Add(1)
}

// Start Stop send/receive loop.
func (h *Hostif) Start() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	log.Debugf("Start module.")

	switch h.state {
	case module.StateInitialize:
		if err := h.conn.Connect(); err != nil {
			return err
		}

		h.client = packets_io.NewPacketsIoClient(h.conn.conn)

		// start send/receive loop
		h.startLoopNoLock()
		h.state = module.StateStarted
	case module.StateSuspended:
		// start send/receive loop
		h.startLoopNoLock()
		h.state = module.StateStarted
	default:
		return fmt.Errorf("Bad state %v", h.state)
	}

	return nil
}

func (h *Hostif) stopLoopNoLock() {
	log.Debugf("Stop loop.")

	// stop send
	if h.sendCancelFuncForConnect != nil {
		h.sendCancelFuncForConnect()
	}
	if h.sendCancelFuncForSend != nil {
		h.sendCancelFuncForSend()
	}
	h.sendStopChannel <- true

	// stop receive
	if h.recvCancelFunc != nil {
		h.recvCancelFunc()
	}
	h.recvStopChannel <- true

	h.localWg.Wait()
}

// Stop Stop send/receive loop.
func (h *Hostif) Stop() {
	h.lock.Lock()
	defer h.lock.Unlock()

	log.Debugf("Stop module.")

	switch h.state {
	case module.StateStarted:
		// stop send/receive loop
		h.stopLoopNoLock()
		h.conn.Disconnect()
		h.globalWg.Done()
		h.state = module.StateInitialize
	case module.StateSuspended:
		h.conn.Disconnect()
		h.globalWg.Done()
		h.state = module.StateInitialize
	default:
		log.Errorf("Bad state %v", h.state)
	}
}

// Resume Resume send/receive loop.
func (h *Hostif) Resume() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	log.Debugf("Resume module.")

	switch h.state {
	case module.StateSuspended:
		// start send/receive loop
		h.startLoopNoLock()
		h.state = module.StateStarted
	default:
		return fmt.Errorf("Bad state %v", h.state)
	}

	return nil
}

// Suspend Suspend send/receive loop.
func (h *Hostif) Suspend() error {
	h.lock.Lock()
	defer h.lock.Unlock()

	log.Debugf("Suspend module.")

	switch h.state {
	case module.StateStarted:
		// stop send/receive loop
		h.stopLoopNoLock()
		h.state = module.StateSuspended
	default:
		return fmt.Errorf("Bad state %v", h.state)
	}

	return nil
}

// Name Module name.
func (h *Hostif) Name() string {
	return HostifModuleName
}
