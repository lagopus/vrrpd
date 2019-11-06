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
	"math"
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testSubinterfaceTestSuite struct {
	suite.Suite
}

func createVRRP1() *VRRP {
	vrrp := NewVRRP()
	vrrp.Vrid = 1
	vas := []net.IP{net.ParseIP("192.168.0.1").To4(), net.ParseIP("10.0.0.1").To4()}
	vrrp.VirtualAddresses = vas
	return vrrp
}

func createVRRP2() *VRRP {
	vrrp := NewVRRP()
	vrrp.Vrid = 2
	vrrp.Priority = 100
	vrrp.Interval = 500
	vrrp.Preempt = false
	vas := []net.IP{net.ParseIP("10.0.0.10").To4(), net.ParseIP("192.168.0.10").To4()}
	vrrp.VirtualAddresses = vas
	return vrrp
}

func createSubinterface(suite *testSubinterfaceTestSuite) *Subinterface {
	iface := NewSubinterface()
	iface.Name = "iface01"
	iface.Index = 0
	iface.IP = net.ParseIP("172.16.0.1").To4()
	iface.Prefix = 24
	var mac net.HardwareAddr
	var err error
	if mac, err = net.ParseMAC("00:11:22:33:44:55"); err != nil {
		suite.Fail("invalid MAC")
	}
	iface.MAC = mac
	vrrp1 := createVRRP1()
	vrrp2 := createVRRP2()
	vrrps := map[uint8]*VRRP{vrrp1.Vrid: vrrp1, vrrp2.Vrid: vrrp2}
	iface.VRRPs = vrrps
	return iface
}

func (suite *testSubinterfaceTestSuite) TestSubinterfaceNewInterface() {
	iface := NewSubinterface()
	suite.Equal("", iface.Name)
	suite.Equal(uint64(math.MaxUint64), iface.Index)
	suite.Nil(iface.IP)
	suite.Equal(uint32(math.MaxUint32), iface.Prefix)
	suite.Nil(iface.MAC)
	suite.EqualValues(map[uint8]*VRRP{}, iface.VRRPs)
}

func (suite *testSubinterfaceTestSuite) TestSubinterfaceIsValid1() {
	iface := NewSubinterface()
	suite.False(iface.IsValid())

	iface.Name = "iface01"
	suite.False(iface.IsValid())

	vrrp := createVRRP1()
	suite.True(vrrp.IsValid())
	iface.Index = 0
	iface.IP = net.ParseIP("192.168.0.1").To4()
	iface.Prefix = 0
	vrrps := map[uint8]*VRRP{vrrp.Vrid: vrrp}
	iface.VRRPs = vrrps
	suite.True(iface.IsValid())
}

func (suite *testSubinterfaceTestSuite) TestSubinterfaceIsValid2() {
	iface := NewSubinterface()
	iface.Name = "iface01"

	vrrp1 := createVRRP1()
	suite.True(vrrp1.IsValid())
	vrrp2 := NewVRRP()
	suite.False(vrrp2.IsValid())
	vrrps := map[uint8]*VRRP{vrrp1.Vrid: vrrp1, vrrp2.Vrid: vrrp2}
	iface.VRRPs = vrrps
	suite.False(iface.IsValid())

	vrrp2 = createVRRP2()
	suite.True(vrrp2.IsValid())
	iface.Index = 0
	iface.IP = net.ParseIP("192.168.0.1").To4()
	iface.Prefix = 32
	vrrps = map[uint8]*VRRP{vrrp1.Vrid: vrrp1, vrrp2.Vrid: vrrp2}
	iface.VRRPs = vrrps
	suite.True(iface.IsValid())
}

func (suite *testSubinterfaceTestSuite) TestSubinterfaceCopy() {
	src := createSubinterface(suite)
	dst := src.Copy()
	suite.True(reflect.DeepEqual(src, dst))
}

func TestSubinterfaceTestSuite(t *testing.T) {
	suite.Run(t, new(testSubinterfaceTestSuite))
}

func BenchmarkSubinterfaceIsValid(b *testing.B) {
	iface := NewSubinterface()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface.IsValid()
	}
}

func BenchmarkSubinterfaceCopy(b *testing.B) {
	iface := NewSubinterface()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iface.Copy()
	}
}
