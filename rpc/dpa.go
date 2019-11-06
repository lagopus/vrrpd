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
	"net"
	"sync"

	rpc "github.com/lagopus/vsw/agents/vrrp/rpc"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// DPAgent DataPlane Agent RPC.
type DPAgent struct {
	conn       *Connection
	client     rpc.VrrpClient
	cancelFunc context.CancelFunc
	isRunning  bool
	wg         *sync.WaitGroup
	lock       sync.Mutex
}

// NewDPAgent New DPAgent.
func NewDPAgent(addr string, port int, wg *sync.WaitGroup) *DPAgent {
	return &DPAgent{
		conn: NewConnection(addr, port),
		wg:   wg,
	}
}

func arrayToVifInfo(vifs []string) *rpc.VifInfo {
	entries := []*rpc.VifEntry{}

	for _, vif := range vifs {
		entry := &rpc.VifEntry{
			Name: vif,
			Vaddr: "",
		}

		entries = append(entries, entry)
	}

	return &rpc.VifInfo{
		N:       uint64(len(entries)),
		Entries: entries,
	}
}

func createVifInfo(name string, phyaddr string, vaddr []string) *rpc.VifInfo {
	entries := []*rpc.VifEntry{}

	// name    : vif name
	// phyaddr : IPv4 address(e.g. 192.168.0.1/24)
	// vaddr   : IPv4 addresses(e.g. 192.168.0.1/24)
	for _, addr := range vaddr {
		entry := &rpc.VifEntry{
			Name: name,
			Phyaddr: phyaddr,
			Vaddr: addr,
		}
		entries = append(entries, entry)
	}

	return &rpc.VifInfo{
		N:       uint64(len(entries)),
		Entries: entries,
	}
}

func toHardwareAddrMap(info *rpc.VifInfo) (map[string]net.HardwareAddr, error) {
	addrMap := make(map[string]net.HardwareAddr)

	for _, entry := range info.Entries {
		if mac, err := net.ParseMAC(entry.Vaddr); err == nil {
			addrMap[entry.Name] = mac
		} else {
			return nil, fmt.Errorf("invalid mac address: %s", entry.Vaddr)
		}
	}

	return addrMap, nil
}

// GetVifMacaddr Get Vif mac address.
func (d *DPAgent) GetVifMacaddr(vif string) (net.HardwareAddr, error) {
	log.Debugf("GetVifMacaddr: %s", vif)

	var addrMap map[string]net.HardwareAddr
	var err error
	if addrMap, err = d.GetVifMacaddrs([]string{vif}); err != nil {
		log.Errorf("GetVifMacaddr failed: %v", err)
		return nil, err
	}

	if addr, ok := addrMap[vif]; ok {
		log.Debugf("GetVifMacaddr return: %v", addr)
		return addr, nil
	}

	log.Errorf("GetVifMacaddr failed: Vif not found")
	return nil, fmt.Errorf("Vif not found: %s", vif)
}

// GetVifMacaddrs Get Vif mac addresses.
func (d *DPAgent) GetVifMacaddrs(vifs []string) (map[string]net.HardwareAddr, error) {
	log.Debugf("GetVifMacaddrs param: %v", vifs)

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFunc = cancel

	opts := []grpc.CallOption{}

	info := arrayToVifInfo(vifs)

	var retInfo *rpc.VifInfo
	var err error
	if retInfo, err = d.client.GetVifInfo(ctx, info, opts...); err != nil {
		log.Errorf("GetVifMacaddrs failed: %v", err)
		return nil, err
	}

	log.Debugf("GetVifMacaddrs return: %v", retInfo)
	return toHardwareAddrMap(retInfo)
}

// ToMaster to master.
func (d *DPAgent) ToMaster(name string, phyaddr string, vaddr []string) error {
	log.Debugf("ToMaster: %v, %v, %v", name, phyaddr, vaddr)

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFunc = cancel

	opts := []grpc.CallOption{}

	info := createVifInfo(name, phyaddr, vaddr)

	if _, err := d.client.ToMaster(ctx, info, opts...); err != nil {
		log.Errorf("ToMaster failed: %v", err)
		return err
	}

	log.Debugf("ToMaster success")
	return nil
}

// ToBackup to backup.
func (d *DPAgent) ToBackup(name string, phyaddr string, vaddr []string) error {
	log.Debugf("ToBackup: %v, %v, %v", name, phyaddr, vaddr)

	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFunc = cancel

	opts := []grpc.CallOption{}

	info := createVifInfo(name, phyaddr, vaddr)

	if _, err := d.client.ToBackup(ctx, info, opts...); err != nil {
		log.Errorf("ToBackup failed: %v", err)
		return err
	}

	log.Debugf("ToBackup success")
	return nil
}

// Start Start DPAgent.
func (d *DPAgent) Start() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.isRunning == false {
		if err := d.conn.Connect(); err != nil {
			return err
		}

		d.client = rpc.NewVrrpClient(d.conn.conn)

		d.isRunning = true
	}

	return nil
}

// Stop Stop DPAgent.
func (d *DPAgent) Stop() {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.isRunning == true {
		if d.cancelFunc != nil {
			d.cancelFunc()
		}

		d.conn.Disconnect()
		d.client = nil

		d.isRunning = false
	}
}

// Resume Resume module.
func (d *DPAgent) Resume() error {
	// implement if necessary.
	return nil
}

// Suspend Suspend module.
func (d *DPAgent) Suspend() error {
	// implement if necessary.
	return nil
}

// Name Module name.
func (d *DPAgent) Name() string {
	return DPAgentModuleName
}
