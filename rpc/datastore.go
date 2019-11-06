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
	"sync"

	ocd "github.com/coreswitch/openconfigd/proto"
	"github.com/lagopus/vrrpd/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// DatastoreCallbackType Type of Callback func for Handler.
type DatastoreCallbackType func(conf *config.AgentConfig)

// Datastore Datastore.
type Datastore struct {
	conn          *Connection
	client        ocd.ConfigClient
	recvChannel   chan *ocd.ConfigReply
	stopChannel   chan bool
	cancelFunc    context.CancelFunc
	isRunning     bool
	configHandler *config.Handler
	configMgr     *config.Mgr
	callbackFunc  DatastoreCallbackType
	wg            *sync.WaitGroup
	lock          sync.Mutex
}

// NewDatastore New Datastore instance.
func NewDatastore(addr string, port int, f DatastoreCallbackType, wg *sync.WaitGroup) *Datastore {
	return &Datastore{
		conn:          NewConnection(addr, port),
		client:        nil,
		recvChannel:   make(chan *ocd.ConfigReply),
		stopChannel:   make(chan bool),
		configHandler: config.GetHandler(),
		configMgr:     config.GetMgr(),
		callbackFunc:  f,
		wg:            wg,
	}
}

func (d *Datastore) recvConfig(stream ocd.Config_DoConfigClient) {
	conf, err := stream.Recv()
	if err != nil {
		log.Errorf("receive config error: %v", err)
		return
	}
	d.recvChannel <- conf
}

func (d *Datastore) recvLoop() {
	defer d.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFunc = cancel

	opts := []grpc.CallOption{}

	stream, err := d.client.DoConfig(ctx, opts...)
	if err != nil {
		log.Errorf("DoConfig error: %v", err)
		return
	}

	// subscribe
	msg := &ocd.ConfigRequest{
		Type:   ocd.ConfigType_SUBSCRIBE,
		Module: "vrrp-agent",
		Path:   []string{"interfaces", "interface"},
	}
	err = stream.Send(msg)
	if err != nil {
		log.Errorf("subscribe send error: %v", err)
		return
	}

	for {
		go d.recvConfig(stream)

		select {
		case conf := <-d.recvChannel:
			log.Infof("[recv] type: %s, path: %s", conf.Type, conf.Path)

			err := d.configHandler.NextState(conf.Type, conf.Path)
			switch conf.Type {
			case ocd.ConfigType_VALIDATE_END:
				if err == nil {
					msg.Type = ocd.ConfigType_VALIDATE_SUCCESS
				} else {
					msg.Type = ocd.ConfigType_VALIDATE_FAILED

					log.Errorf("validate failure: %v", err)
				}

				// send validate result.
				err = stream.Send(msg)
				if err != nil {
					log.Errorf("validation send error: %v", err)
					return
				}
			case ocd.ConfigType_COMMIT_END:
				if err == nil {
					log.Info("commit success")
					d.callbackFunc(d.configMgr.GetCurrentConfig())
				} else {
					log.Errorf("commit failure: %v", err)
				}
			default:
				// NOP
			}
		case <-d.stopChannel:
			log.Infof("Stop recvConfig loop.")

			d.configHandler.Reset()

			if d.cancelFunc != nil {
				d.cancelFunc()
			}
			_ = stream.CloseSend()
			d.conn.Disconnect()
			return
		}
	}
}

// Start Start Receiver loop.
func (d *Datastore) Start() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.isRunning == false {
		if err := d.conn.Connect(); err != nil {
			return err
		}

		d.client = ocd.NewConfigClient(d.conn.conn)

		d.wg.Add(1)
		go d.recvLoop()

		d.isRunning = true
	}

	return nil
}

// Stop Stop Receiver loop.
func (d *Datastore) Stop() {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.isRunning == true {
		d.stopChannel <- true
		d.isRunning = false
	}
}

// Resume Resume module.
func (d *Datastore) Resume() error {
	// implement if necessary.
	return nil
}

// Suspend Suspend module.
func (d *Datastore) Suspend() error {
	// implement if necessary.
	return nil
}

// Name Module name.
func (d *Datastore) Name() string {
	return DatastoreModuleName
}
