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
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/coreswitch/cmd"
	ocd "github.com/coreswitch/openconfigd/proto"
	log "github.com/sirupsen/logrus"
)

// Handler ConfigHandler.
type Handler struct {
	parser *cmd.Node
	state  State
	lock   sync.Mutex
}

var handler = newHandler()

func (h *Handler) doValidate() error {
	if cmgr.GetModifiedConfig().IsValid() {
		cmgr.Rollback()
		return nil
	}

	return errors.New("validation failed")
}

func (h *Handler) doCommit() error {
	if cmgr.Commit() {
		return nil
	}

	cmgr.Rollback()
	return errors.New("commit failed")
}

func (h *Handler) doSet(path []string) error {
	ret, fn, args, _ := h.parser.ParseCmd(path)
	if ret == cmd.ParseSuccess {
		fn.(func(int, cmd.Args) int)(cmd.Set, args)
		return nil
	}

	// Ignore error of parse.
	// Because command don't need.
	log.Warnf("Ignore: parse set command failed: %s", path)
	return nil
}

func (h *Handler) doDelete(path []string) error {
	ret, fn, args, _ := h.parser.ParseCmd(path)
	if ret == cmd.ParseSuccess {
		fn.(func(int, cmd.Args) int)(cmd.Delete, args)
		return nil
	}

	// Ignore error of parse.
	// Because command don't need.
	log.Warnf("Ignore: parse delete command failed: %s", path)
	return nil
}

// State Machine.
//
//     +---+   [Start]
//     | . |------------------+
//     +---+                  |
//                            V
//     [validation end]   +-----------------+     [commit end]
//            +---------->|                 |<----------+
//            |           |    Initialize   |           |
//            |   +-------|                 |-------+   |
//            |   |       +-----------------+       |   |
//            |   |[validation start]               |   |
//            |   |                                 |   |
//            |   |                   [commit start]|   |
//            |   V                                 V   |
//    +---------------+                          +---------------+
//    |               |                          |               |
//    |  Validation   |                          |     Commit    |
//    |               |                          |               |
//    +---------------+                          +---------------+
//

// NextState Next state.
func (h *Handler) NextState(t ocd.ConfigType, path []string) error {
	h.lock.Lock()
	defer func() {
		h.lock.Unlock()
		log.Debugf("handler next state: %v", h.state)
	}()

	log.Debugf("handler state: %v", h.state)

	switch h.state {
	case StateInitialize:
		switch t {
		case ocd.ConfigType_VALIDATE_START:
			h.state = StateValidation
			return nil
		case ocd.ConfigType_COMMIT_START:
			h.state = StateCommit
			return nil
		default:
			return fmt.Errorf("Bad type %v in StateInitialize", t)
		}
	case StateValidation:
		switch t {
		case ocd.ConfigType_VALIDATE_END:
			log.Infof("validate config: %v", cmgr.GetModifiedConfig())

			// validate
			if err := h.doValidate(); err != nil {
				log.Errorf("validate failure: %v", err)
				return err
			}

			h.state = StateInitialize
			return nil
		case ocd.ConfigType_SET:
			if err := h.doSet(path); err != nil {
				log.Errorf("set failure: %v", err)
				return err
			}

			return nil
		case ocd.ConfigType_DELETE:
			if err := h.doDelete(path); err != nil {
				log.Errorf("delete failure: %v", err)
				return err
			}

			return nil
		default:
			return fmt.Errorf("Bad type %v in StateValidation", t)
		}
	case StateCommit:
		switch t {
		case ocd.ConfigType_COMMIT_END:
			log.Infof("commit config: %v", cmgr.GetModifiedConfig())

			// commit
			if err := h.doCommit(); err != nil {
				log.Error("commit failure")
				return err
			}

			log.Info("commit success")
			h.state = StateInitialize
			return nil
		case ocd.ConfigType_SET:
			if err := h.doSet(path); err != nil {
				return err
			}

			return nil
		case ocd.ConfigType_DELETE:
			if err := h.doDelete(path); err != nil {
				return err
			}

			return nil
		default:
			return fmt.Errorf("Bad type %v in StateCommit", t)
		}
	default:
		return fmt.Errorf("Bad config state %v", h.state)
	}
}

// Reset Reset handler processing.
func (h *Handler) Reset() {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.state = StateInitialize
	cmgr.Rollback()
}

func createSubifname(ifname string, subifidx uint64) string {
	return fmt.Sprintf("%s-%d", ifname, subifidx)
}

func interfaceConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)

	if Cmd == cmd.Set {
		cmgr.modified.AddInterface(ifname)
	} else if Cmd == cmd.Delete {
		cmgr.modified.DeleteInterface(ifname)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func interfaceTypeConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	iftype := Args[1].(string)

	if Cmd == cmd.Set {
		cmgr.modified.SetInterfaceType(ifname, iftype)
	} else if Cmd == cmd.Delete {
		cmgr.modified.DeleteInterfaceType(ifname)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func subIfConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.AddSubinterface(ifname, subifname)
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
	} else if Cmd == cmd.Delete {
		cmgr.modified.DeleteSubinterface(ifname, subifname)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func subIfIpv4AddressConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)
	subifaddr := Args[2].(net.IP)

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
		cmgr.modified.SetSubifIP(ifname, subifname, subifaddr)
	} else if Cmd == cmd.Delete {
		cmgr.modified.DeleteSubifIP(ifname, subifname)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func subIfIpv4PrefixConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)
	subifaddr := Args[2].(net.IP)
	prefix := uint32(Args[3].(uint64))

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
		cmgr.modified.SetSubifIP(ifname, subifname, subifaddr)
		cmgr.modified.SetSubifPrefix(ifname, subifname, prefix)
	} else if Cmd == cmd.Delete {
		cmgr.modified.DeleteSubifPrefix(ifname, subifname)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func vrrpGroupConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)
	subifaddr := Args[2].(net.IP)
	vrid := uint8(Args[3].(uint64))

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
		cmgr.modified.SetSubifIP(ifname, subifname, subifaddr)
		cmgr.modified.AddVrrp(ifname, subifname, vrid)
	} else if Cmd == cmd.Delete {
		cmgr.modified.DeleteVrrp(ifname, subifname, vrid)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func vrrpVaddressConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)
	subifaddr := Args[2].(net.IP)
	vrid := uint8(Args[3].(uint64))
	addr := Args[4].(net.IP)

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
		cmgr.modified.SetSubifIP(ifname, subifname, subifaddr)
		cmgr.modified.AddVrrpVirtualAddress(ifname, subifname, vrid, addr)
	} else if Cmd == cmd.Delete {
		cmgr.modified.DeleteVrrpVirtualAddress(ifname, subifname, vrid, addr)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func vrrpPriorityConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)
	subifaddr := Args[2].(net.IP)
	vrid := uint8(Args[3].(uint64))
	priority := uint8(Args[4].(uint64))

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
		cmgr.modified.SetSubifIP(ifname, subifname, subifaddr)
		cmgr.modified.SetVrrpPriority(ifname, subifname, vrid, priority)
	} else if Cmd == cmd.Delete {
		cmgr.modified.SetDefaultVrrpPriority(ifname, subifname, vrid)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func vrrpPreemptConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)
	subifaddr := Args[2].(net.IP)
	vrid := uint8(Args[3].(uint64))
	// TODO: error check
	preempt, _ := strconv.ParseBool(Args[4].(string))

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
		cmgr.modified.SetSubifIP(ifname, subifname, subifaddr)
		cmgr.modified.SetVrrpPreempt(ifname, subifname, vrid, preempt)
	} else if Cmd == cmd.Delete {
		cmgr.modified.SetDefaultVrrpPreempt(ifname, subifname, vrid)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func vrrpAdvIntervalConf(Cmd int, Args cmd.Args) int {
	log.Debugf("command type: %d, args: %v", Cmd, Args)

	ifname := Args[0].(string)
	subifidx := Args[1].(uint64)
	subifaddr := Args[2].(net.IP)
	vrid := uint8(Args[3].(uint64))
	interval := uint16(Args[4].(uint64))

	subifname := createSubifname(ifname, subifidx)

	if Cmd == cmd.Set {
		cmgr.modified.SetSubifIndex(ifname, subifname, subifidx)
		cmgr.modified.SetSubifIP(ifname, subifname, subifaddr)
		cmgr.modified.SetVrrpInterval(ifname, subifname, vrid, interval)
	} else if Cmd == cmd.Delete {
		cmgr.modified.SetDefaultVrrpInterval(ifname, subifname, vrid)
	}

	log.Debugf("modified config: %v", cmgr.modified.String())

	return cmd.Success
}

func newHandler() *Handler {
	p := cmd.NewParser()
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD"},
		interfaceConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"config",
		"type", "WORD"},
		interfaceTypeConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>"},
		subIfConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>",
		"ipv4",
		"addresses",
		"address", "A.B.C.D"},
		subIfIpv4AddressConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>",
		"ipv4",
		"addresses",
		"address", "A.B.C.D",
		"config",
		"prefix-length", "<0-32>"},
		subIfIpv4PrefixConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>",
		"ipv4",
		"addresses",
		"address", "A.B.C.D",
		"vrrp",
		"vrrp-group", "<1-255>"},
		vrrpGroupConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>",
		"ipv4",
		"addresses",
		"address", "A.B.C.D",
		"vrrp",
		"vrrp-group", "<1-255>",
		"config",
		"virtual-address", "A.B.C.D"},
		vrrpVaddressConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>",
		"ipv4",
		"addresses",
		"address", "A.B.C.D",
		"vrrp",
		"vrrp-group", "<1-255>",
		"config",
		"priority", "<1-254>"},
		vrrpPriorityConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>",
		"ipv4",
		"addresses",
		"address", "A.B.C.D",
		"vrrp",
		"vrrp-group", "<1-255>",
		"config",
		"preempt", "WORD"},
		vrrpPreemptConf)
	p.InstallCmd([]string{"interfaces",
		"interface", "WORD",
		"subinterfaces",
		"subinterface", "<0-4294967295>",
		"ipv4",
		"addresses",
		"address", "A.B.C.D",
		"vrrp",
		"vrrp-group", "<1-255>",
		"config",
		"advertisement-interval", "<1-4095>"},
		vrrpAdvIntervalConf)

	return &Handler{
		parser: p,
		state:  StateInitialize,
	}
}

// GetHandler Get ConfigHandler instance.
func GetHandler() *Handler {
	return handler
}
