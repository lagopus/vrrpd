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

package module

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

var moduleArray = []Module{}
var moduleMap = map[string]Module{}
var state = StateInitialize
var lock sync.Mutex

// Module Module.
type Module interface {
	Start() error
	Stop()
	Resume() error
	Suspend() error
	Name() string
}

// RegisterModule Regist module.
func RegisterModule(module Module) {
	lock.Lock()
	defer lock.Unlock()
	moduleArray = append(moduleArray, module)
	moduleMap[module.Name()] = module
}

// GetModule Get module.
func GetModule(name string) Module {
	lock.Lock()
	defer lock.Unlock()
	return moduleMap[name]
}

// GetState Get state.
func GetState() State {
	lock.Lock()
	defer lock.Unlock()
	return state
}

// StartModules Start All Modules.
func StartModules() error {
	lock.Lock()
	defer lock.Unlock()
	for _, module := range moduleArray {
		log.Debugf("Start module: %s", module.Name())
		if err := module.Start(); err != nil {
			return err
		}
	}

	state = StateStarted

	return nil
}

// StopModules Stop All Modules.
func StopModules() {
	lock.Lock()
	defer lock.Unlock()
	for i := len(moduleArray) - 1; i >= 0; i-- {
		log.Debugf("Stop module: %s", moduleArray[i].Name())
		moduleArray[i].Stop()
	}

	state = StateInitialize
}

// ResumeModules Start All Modules.
func ResumeModules() error {
	lock.Lock()
	defer lock.Unlock()
	for _, module := range moduleArray {
		log.Debugf("Resume module: %s", module.Name())
		if err := module.Resume(); err != nil {
			return err
		}
	}

	state = StateStarted

	return nil
}

// SuspendModules Suspend All Modules.
func SuspendModules() error {
	lock.Lock()
	defer lock.Unlock()
	for i := len(moduleArray) - 1; i >= 0; i-- {
		log.Debugf("Suspend module: %s", moduleArray[i].Name())
		if err := moduleArray[i].Suspend(); err != nil {
			return err
		}
	}

	state = StateSuspended

	return nil
}
