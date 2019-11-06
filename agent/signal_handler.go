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
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lagopus/vrrpd/logger"
	"github.com/lagopus/vrrpd/module"
	log "github.com/sirupsen/logrus"
)

const (
	// SignalHandlerModuleName SignalHandler module name.
	SignalHandlerModuleName = "SignalHandlerModule"

	// ForceShutdownTime Force shutdown time.
	ForceShutdownTime = 5 * time.Second
)

// SignalHandler Signal handler.
type SignalHandler struct {
	signalChannel chan os.Signal
	isRunning     bool
	wg            *sync.WaitGroup
	lock          sync.Mutex
}

// NewSignalHandler New SignalHandler module.
func NewSignalHandler(wg *sync.WaitGroup) *SignalHandler {
	sh := &SignalHandler{
		signalChannel: make(chan os.Signal),
		wg:            wg,
	}
	return sh
}

func (sh *SignalHandler) forceShutdownTimer() {
	select {
	case <-time.After(ForceShutdownTime):
		log.Errorf("Force shutdown.")
		_ = syscall.Kill(os.Getpid(), syscall.SIGKILL)
	}
}

func (sh *SignalHandler) handleExit() {
	log.Debugf("call handleExit.")
	// Force shutdown timer.
	// NOTE: Can't be canceled at reconnection retry for GRPC.
	go sh.forceShutdownTimer()
	module.StopModules()
	log.Infof("Stop signalHandlerLoop.")
}

func (sh *SignalHandler) handleHup() {
	log.Debugf("call handleHup.")
	if err := logger.Rotate(); err != nil {
		log.Errorf("Can't log rotate: %v.", err)
	}
}

func (sh *SignalHandler) signalHandlerLoop() {
	defer sh.wg.Done()

	signal.Notify(sh.signalChannel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	for {
		select {
		case s := <-sh.signalChannel:
			switch s {
			case syscall.SIGHUP:
				sh.handleHup()
			case syscall.SIGINT:
				sh.handleExit()
				return
			case syscall.SIGTERM:
				sh.handleExit()
				return
			case syscall.SIGQUIT:
				sh.handleExit()
				return
			}
		}
	}
}

// Start Start signal handler.
func (sh *SignalHandler) Start() error {
	sh.lock.Lock()
	defer sh.lock.Unlock()

	if sh.isRunning == false {
		sh.wg.Add(1)
		go sh.signalHandlerLoop()
		sh.isRunning = true
	}

	return nil
}

// Stop Stop signal handler.
func (sh *SignalHandler) Stop() {
	sh.lock.Lock()
	defer sh.lock.Unlock()

	if sh.isRunning == true {
		sh.isRunning = false
	}
}

// Resume Resume module.
func (sh *SignalHandler) Resume() error {
	// implement if necessary.
	return nil
}

// Suspend Suspend module.
func (sh *SignalHandler) Suspend() error {
	// implement if necessary.
	return nil
}

// Name Module name.
func (sh *SignalHandler) Name() string {
	return SignalHandlerModuleName
}
