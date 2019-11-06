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

package models

import (
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testInterfaceTestSuite struct {
	suite.Suite
}

func createSubinterface1(suite *testInterfaceTestSuite) *Subinterface {
	vrrp1 := NewVRRP()
	vrrp1.Vrid = 1
	vrrp1.Priority = 255
	vrrp1.Interval = 100
	vrrp1.Preempt = true
	vrrp1.VirtualAddresses = []net.IP{net.ParseIP("192.168.0.1").To4(), net.ParseIP("10.0.0.1").To4()}

	vrrp2 := NewVRRP()
	vrrp2.Vrid = 2
	vrrp2.Priority = 100
	vrrp2.Interval = 500
	vrrp2.Preempt = false
	vrrp2.VirtualAddresses = []net.IP{net.ParseIP("10.0.0.10").To4(), net.ParseIP("192.168.0.10").To4()}

	subiface := NewSubinterface()
	subiface.Name = "subinterface01"
	subiface.Index = 0
	subiface.IP = net.ParseIP("172.16.0.1").To4()
	subiface.Prefix = 24
	var mac net.HardwareAddr
	var err error
	if mac, err = net.ParseMAC("00:11:22:33:44:55"); err != nil {
		suite.Fail("invalid MAC")
	}
	subiface.MAC = mac
	subiface.VRRPs = map[uint8]*VRRP{vrrp1.Vrid: vrrp1, vrrp2.Vrid: vrrp2}

	return subiface
}

func createSubinterface2(suite *testInterfaceTestSuite) *Subinterface {
	vrrp1 := NewVRRP()
	vrrp1.Vrid = 10
	vrrp1.Priority = 100
	vrrp1.Interval = 1000
	vrrp1.Preempt = true
	vrrp1.VirtualAddresses = []net.IP{net.ParseIP("192.168.0.1").To4(), net.ParseIP("10.0.0.1").To4()}

	vrrp2 := NewVRRP()
	vrrp2.Vrid = 20
	vrrp2.Priority = 255
	vrrp2.Interval = 5000
	vrrp2.Preempt = false
	vrrp2.VirtualAddresses = []net.IP{net.ParseIP("10.0.0.10").To4(), net.ParseIP("192.168.0.10").To4()}

	subiface := NewSubinterface()
	subiface.Name = "subinterface02"
	subiface.Index = 0
	subiface.IP = net.ParseIP("172.16.0.10").To4()
	subiface.Prefix = 32
	var mac net.HardwareAddr
	var err error
	if mac, err = net.ParseMAC("00:11:22:aa:bb:cc"); err != nil {
		suite.Fail("invalid MAC")
	}
	subiface.MAC = mac
	subiface.VRRPs = map[uint8]*VRRP{vrrp1.Vrid: vrrp1, vrrp2.Vrid: vrrp2}

	return subiface
}

func (suite *testInterfaceTestSuite) TestInterfaceNewInterface() {
	iface := NewInterface()
	suite.Equal("", iface.Name)
	suite.EqualValues(map[string]*Subinterface{}, iface.Subinterfaces)
}

func (suite *testInterfaceTestSuite) TestInterfaceIsValid() {
	iface := NewInterface()
	suite.True(iface.IsValid())

	iface.Name = "iface01"
	suite.True(iface.IsValid())

	subiface1 := createSubinterface1(suite)
	suite.True(subiface1.IsValid())
	subiface2 := createSubinterface2(suite)
	suite.True(subiface2.IsValid())
	subinterfaces := map[string]*Subinterface{subiface1.Name: subiface1, subiface2.Name: subiface2}
	iface.Subinterfaces = subinterfaces
	suite.True(iface.IsValid())
}

func (suite *testInterfaceTestSuite) TestInterfaceCopy() {
	src := createSubinterface1(suite)
	dst := src.Copy()
	suite.True(reflect.DeepEqual(src, dst))
}

func TestInterfaceTestSuite(t *testing.T) {
	suite.Run(t, new(testInterfaceTestSuite))
}

func BenchmarkInterfaceIsValid(b *testing.B) {
	iface := NewInterface()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface.IsValid()
	}
}

func BenchmarkInterfaceCopy(b *testing.B) {
	iface := NewInterface()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface.Copy()
	}
}
