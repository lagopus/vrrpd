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
	"net"
	"reflect"
	"testing"

	"github.com/lagopus/vrrpd/models"
	"github.com/stretchr/testify/suite"
)

type testAgentConfigTestSuite struct {
	suite.Suite
}

func createInterface1(suite *testAgentConfigTestSuite) *models.Interface {
	vrrp1 := models.NewVRRP()
	vrrp1.Vrid = 1
	vrrp1.Priority = 255
	vrrp1.Interval = 100
	vrrp1.Preempt = true
	vrrp1.VirtualAddresses = []net.IP{net.ParseIP("192.168.0.1"), net.ParseIP("10.0.0.1")}

	vrrp2 := models.NewVRRP()
	vrrp2.Vrid = 2
	vrrp2.Priority = 100
	vrrp2.Interval = 500
	vrrp2.Preempt = false
	vrrp2.VirtualAddresses = []net.IP{net.ParseIP("10.0.0.10"), net.ParseIP("192.168.0.10")}

	subiface := models.NewSubinterface()
	subiface.Name = "subinterface01"
	subiface.Index = 0
	subiface.IP = net.ParseIP("172.16.0.1")
	subiface.Prefix = 24
	var mac net.HardwareAddr
	var err error
	if mac, err = net.ParseMAC("00:11:22:33:44:55"); err != nil {
		suite.Fail("invalid MAC")
	}
	subiface.MAC = mac
	subiface.VRRPs = map[uint8]*models.VRRP{vrrp1.Vrid: vrrp1, vrrp2.Vrid: vrrp2}

	iface := models.NewInterface()
	iface.Name = "iface01"
	iface.Type = models.IfTypeEthernetCsmacd
	iface.Subinterfaces = map[string]*models.Subinterface{subiface.Name: subiface}

	return iface
}

func createInterface2(suite *testAgentConfigTestSuite) *models.Interface {
	vrrp1 := models.NewVRRP()
	vrrp1.Vrid = 100
	vrrp1.Priority = 100
	vrrp1.Interval = 1000
	vrrp1.Preempt = true
	vrrp1.VirtualAddresses = []net.IP{net.ParseIP("192.168.0.1"), net.ParseIP("10.0.0.1")}

	vrrp2 := models.NewVRRP()
	vrrp2.Vrid = 200
	vrrp2.Priority = 255
	vrrp2.Interval = 5000
	vrrp2.Preempt = false
	vrrp2.VirtualAddresses = []net.IP{net.ParseIP("10.0.0.10"), net.ParseIP("192.168.0.10")}

	subiface := models.NewSubinterface()
	subiface.Name = "subinterface02"
	subiface.Index = 0
	subiface.IP = net.ParseIP("172.16.0.10")
	subiface.Prefix = 32
	var mac net.HardwareAddr
	var err error
	if mac, err = net.ParseMAC("00:11:22:aa:bb:cc"); err != nil {
		suite.Fail("invalid MAC")
	}
	subiface.MAC = mac
	subiface.VRRPs = map[uint8]*models.VRRP{vrrp1.Vrid: vrrp1, vrrp2.Vrid: vrrp2}

	iface := models.NewInterface()
	iface.Name = "iface02"
	iface.Type = models.IfTypeEthernetCsmacd
	iface.Subinterfaces = map[string]*models.Subinterface{subiface.Name: subiface}

	return iface
}

func createTunnelInterface(suite *testAgentConfigTestSuite) *models.Interface {
	subiface := models.NewSubinterface()
	subiface.Name = "tunnel-subinterface01"
	subiface.Index = 0

	iface := models.NewInterface()
	iface.Name = "tunnel-iface01"
	iface.Type = models.IfTypeTunnel
	iface.Subinterfaces = map[string]*models.Subinterface{subiface.Name: subiface}

	return iface
}

func (suite *testAgentConfigTestSuite) TestAgentConfigNewAgentConfig() {
	agentConfig := newAgentConfig()
	suite.Equal(net.ParseIP("127.0.0.1"), agentConfig.DsAddr)
	suite.Equal(uint16(2650), agentConfig.DsPort)
	suite.Equal(net.ParseIP("127.0.0.1"), agentConfig.DpaAddr)
	suite.Equal(uint16(30010), agentConfig.DpaPort)
	suite.Equal(net.ParseIP("127.0.0.1"), agentConfig.HostifAddr)
	suite.Equal(uint16(30020), agentConfig.HostifPort)
	suite.EqualValues(map[string]*models.Interface{}, agentConfig.Interfaces)
}

func (suite *testAgentConfigTestSuite) TestAgentConfigIsValid() {
	agentConfig := newAgentConfig()
	suite.True(agentConfig.IsValid())

	interfaces := map[string]*models.Interface{}
	iface1 := createInterface1(suite)
	iface2 := createInterface2(suite)
	tunnelIface := createTunnelInterface(suite)
	interfaces[iface1.Name] = iface1
	interfaces[iface2.Name] = iface2
	interfaces[tunnelIface.Name] = tunnelIface
	agentConfig.Interfaces = interfaces
	suite.True(agentConfig.IsValid())
}

func (suite *testAgentConfigTestSuite) TestAgentConfigCopy() {
	src := createInterface1(suite)
	dst := src.Copy()
	suite.True(reflect.DeepEqual(src, dst))
}

func TestAgentConfigTestSuite(t *testing.T) {
	suite.Run(t, new(testAgentConfigTestSuite))
}

func BenchmarkAgentConfigIsValid(b *testing.B) {
	agentConfig := newAgentConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agentConfig.IsValid()
	}
}

func BenchmarkAgentConfigCopy(b *testing.B) {
	agentConfig := newAgentConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agentConfig.Copy()
	}
}
