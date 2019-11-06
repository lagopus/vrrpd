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

package config

import (
	"errors"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Mgr Config manager.
type Mgr struct {
	current  *AgentConfig
	modified *AgentConfig
	lock     sync.RWMutex
}

func newMgr() *Mgr {
	return &Mgr{
		current:  newAgentConfig(),
		modified: newAgentConfig(),
	}
}

var cmgr = newMgr()

func (cmgr *Mgr) setModifiedConfig(agencConfig *AgentConfig) {
	cmgr.lock.RLock()
	defer cmgr.lock.RUnlock()
	cmgr.modified = agencConfig
}

// GetCurrentConfig Get current config.
func (cmgr *Mgr) GetCurrentConfig() *AgentConfig {
	cmgr.lock.RLock()
	defer cmgr.lock.RUnlock()
	return cmgr.current.Copy()
}

// GetModifiedConfig Get modified config.
func (cmgr *Mgr) GetModifiedConfig() *AgentConfig {
	cmgr.lock.RLock()
	defer cmgr.lock.RUnlock()
	return cmgr.modified.Copy()
}

// Commit Commit modified config.
func (cmgr *Mgr) Commit() bool {
	cmgr.lock.RLock()
	defer cmgr.lock.RUnlock()

	log.Infof("[commit] current config : %v", cmgr.current.String())
	log.Infof("[commit] modified config: %v", cmgr.modified.String())

	cmgr.current, cmgr.modified = cmgr.modified, cmgr.modified.Copy()

	return true
}

// Rollback Rollback modified config.
func (cmgr *Mgr) Rollback() {
	cmgr.lock.RLock()
	defer cmgr.lock.RUnlock()

	cmgr.modified = cmgr.current.Copy()
}

// ReadConfig Read config(YAML).
func (cmgr *Mgr) ReadConfig(path string) error {
	agentConfig := newAgentConfig()

	dir := filepath.Dir(path)
	r := regexp.MustCompile(`.yaml|.yml$`)
	file := filepath.Base(r.ReplaceAllString(path, ""))

	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)
	viper.SetConfigName(file)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	var dsAddr net.IP
	if viper.IsSet("datastore.addr") {
		dsAddr = net.ParseIP(viper.GetString("datastore.addr"))
		if dsAddr == nil {
			return errors.New("datastore.addr is invalid")
		}
	} else {
		return errors.New("datastore.addr is null")
	}

	var dsPort uint16
	if viper.IsSet("datastore.port") {
		tmp, err := strconv.ParseUint(viper.GetString("datastore.port"), 10, 16)
		if err == nil {
			dsPort = uint16(tmp)
		} else {
			return err
		}
	} else {
		return errors.New("datastore.port is null")
	}

	var dpaAddr net.IP
	if viper.IsSet("dpa.addr") {
		dpaAddr = net.ParseIP(viper.GetString("dpa.addr"))
		if dpaAddr == nil {
			return errors.New("dpa.addr is invalid")
		}
	} else {
		return errors.New("dpa.addr is null")
	}

	var dpaPort uint16
	if viper.IsSet("dpa.port") {
		tmp, err := strconv.ParseUint(viper.GetString("dpa.port"), 10, 16)
		if err == nil {
			dpaPort = uint16(tmp)
		} else {
			return err
		}
	} else {
		return errors.New("dpa.port is null")
	}

	var hostifAddr net.IP
	if viper.IsSet("hostif.addr") {
		hostifAddr = net.ParseIP(viper.GetString("hostif.addr"))
		if hostifAddr == nil {
			return errors.New("hostif.addr is invalid")
		}
	} else {
		return errors.New("hostif.addr is null")
	}

	var hostifPort uint16
	if viper.IsSet("hostif.port") {
		tmp, err := strconv.ParseUint(viper.GetString("hostif.port"), 10, 16)
		if err == nil {
			hostifPort = uint16(tmp)
		} else {
			return err
		}
	} else {
		return errors.New("hostif.port is null")
	}

	agentConfig.DsAddr = dsAddr
	agentConfig.DsPort = dsPort
	agentConfig.DpaAddr = dpaAddr
	agentConfig.DpaPort = dpaPort
	agentConfig.HostifAddr = hostifAddr
	agentConfig.HostifPort = hostifPort

	cmgr.setModifiedConfig(agentConfig)
	cmgr.Commit()

	return nil
}

// GetMgr Get ConfigMgr instance.
func GetMgr() *Mgr {
	return cmgr
}
