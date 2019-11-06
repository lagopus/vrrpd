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
	"net"
	"testing"

	rpc "github.com/lagopus/vsw/agents/vrrp/rpc"
	"github.com/stretchr/testify/suite"
)

type testDpaTestSuite struct {
	suite.Suite
}

func (suite *testDpaTestSuite) TestDpaArrayToVifInfo() {
	entry1 := &rpc.VifEntry{
		Name:  "vif1",
		Vaddr: "",
	}
	entry2 := &rpc.VifEntry{
		Name:  "vif2",
		Vaddr: "",
	}
	expected := &rpc.VifInfo{
		N:       2,
		Entries: []*rpc.VifEntry{entry1, entry2},
	}

	vifs := []string{"vif1", "vif2"}
	actual := arrayToVifInfo(vifs)

	suite.Equal(expected.N, actual.N)
	suite.Equal(expected.Entries[0].Name, actual.Entries[0].Name)
	suite.Equal(expected.Entries[0].Vaddr, actual.Entries[0].Vaddr)
	suite.Equal(expected.Entries[1].Name, actual.Entries[1].Name)
	suite.Equal(expected.Entries[1].Vaddr, actual.Entries[1].Vaddr)

}

func (suite *testDpaTestSuite) TestDpaCreateVifInfo() {
	entry1 := &rpc.VifEntry{
		Name:    "vif1",
		Vaddr:   "192.168.0.1/16",
		Phyaddr: "192.168.0.100",
	}
	entry2 := &rpc.VifEntry{
		Name:    "vif1",
		Vaddr:   "10.0.0.1/8",
		Phyaddr: "192.168.0.100",
	}
	expected := &rpc.VifInfo{
		N:       2,
		Entries: []*rpc.VifEntry{entry1, entry2},
	}

	actual := createVifInfo("vif1", "192.168.0.100", []string{"192.168.0.1/16", "10.0.0.1/8"})

	suite.Equal(expected.N, actual.N)
	for _, entry := range actual.Entries {
		if (entry.Name == "vif1") && (entry.Phyaddr == "192.168.0.100") {
			if (entry.Vaddr != "192.168.0.1/16") && (entry.Vaddr != "10.0.0.1/8") {
				suite.Fail("invalid Vaddr")
			}
		} else {
			suite.Fail("invalid name")
		}
	}
}

func (suite *testDpaTestSuite) TestDpaToHardwareAddrMap() {
	expected := map[string]net.HardwareAddr{
		"vif1": net.HardwareAddr{0x00, 0x00, 0x00, 0x11, 0x11, 0x11},
		"vif2": net.HardwareAddr{0x00, 0x00, 0x00, 0x22, 0x22, 0x22}}

	entry1 := &rpc.VifEntry{
		Name: "vif1",
		Vaddr: "00:00:00:11:11:11",
	}
	entry2 := &rpc.VifEntry{
		Name: "vif2",
		Vaddr: "00:00:00:22:22:22",
	}
	info := &rpc.VifInfo{
		N:       2,
		Entries: []*rpc.VifEntry{entry1, entry2},
	}

	if actual, err := toHardwareAddrMap(info); err == nil {
		suite.Equal(len(expected), len(actual))
		for key := range expected {
			_, ok := actual[key]
			suite.True(ok)
			suite.Equal(expected[key], actual[key])
		}
	} else {
		suite.Fail("error occurred")
	}
}

func TestDpaTestSuite(t *testing.T) {
	suite.Run(t, new(testDpaTestSuite))
}
