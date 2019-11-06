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

type testVRRPTestSuite struct {
	suite.Suite
}

func (suite *testVRRPTestSuite) TestVRRPNewVRRP() {
	vrrp := NewVRRP()
	suite.Equal(uint8(0), vrrp.Vrid)
	suite.Equal(uint8(100), vrrp.Priority)
	suite.Equal(true, vrrp.Preempt)
	suite.Equal(false, vrrp.Accept)
	suite.Equal(uint16(100), vrrp.Interval)
	suite.EqualValues([]net.IP{}, vrrp.VirtualAddresses)
}

func (suite *testVRRPTestSuite) TestVRRPIsValid() {
	vrrp := NewVRRP()
	suite.False(vrrp.IsValid())

	vrrp.Vrid = 1
	suite.False(vrrp.IsValid())

	vas := []net.IP{net.ParseIP("192.168.0.1").To4()}
	vrrp.VirtualAddresses = vas
	suite.True(vrrp.IsValid())

	vrrp.DeleteVirtualAddress(net.ParseIP("192.168.0.1").To4())
	suite.False(vrrp.IsValid())
}

func (suite *testVRRPTestSuite) TestVRRPCopy() {
	src := NewVRRP()
	src.Vrid = 1
	vas := []net.IP{net.ParseIP("192.168.0.1").To4(), net.ParseIP("10.0.0.1").To4()}
	src.VirtualAddresses = vas

	dst := src.Copy()
	suite.True(reflect.DeepEqual(src, dst))
}

func (suite *testVRRPTestSuite) TestVRRPIsMaster() {
	vrrp := NewVRRP()
	suite.False(vrrp.IsValid())

	vrrp.Vrid = 1
	suite.False(vrrp.IsValid())

	vas := []net.IP{net.ParseIP("192.168.0.1").To4()}
	vrrp.VirtualAddresses = vas

	suite.True(vrrp.IsMaster(net.ParseIP("192.168.0.1").To4()))
	suite.False(vrrp.IsMaster(net.ParseIP("10.0.0.1").To4()))
}

func TestVRRPTestSuite(t *testing.T) {
	suite.Run(t, new(testVRRPTestSuite))
}

func BenchmarkVRRPIsValid(b *testing.B) {
	vrrp := NewVRRP()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vrrp.IsValid()
	}
}

func BenchmarkVRRPCopy(b *testing.B) {
	vrrp := NewVRRP()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vrrp.Copy()
	}
}
