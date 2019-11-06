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
	"testing"

	hostif "github.com/lagopus/vsw/modules/hostif/packets_io"
	"github.com/stretchr/testify/suite"
)

type testPacketTestSuite struct {
	suite.Suite
}

func (suite *testPacketTestSuite) TestNewPacket() {
	expectedSubifname := "vif1"
	expectedData := []byte{0x01, 0x00, 0x5e, 0x00, 0x00, 0x12}
	actual := NewPacket(expectedSubifname, expectedData)

	suite.Equal(expectedSubifname, actual.Subifname)
	suite.Equal(expectedData, actual.Data)
}

func (suite *testPacketTestSuite) TestNewBulkPackets1() {
	expectedSubifname1 := "vif1"
	expectedData1 := []byte{0x01, 0x00, 0x5e, 0x00, 0x00, 0x12}

	expectedSubifname2 := "vif2"
	expectedData2 := []byte{0x00, 0x01, 0x00, 0x5e, 0x12, 0x00}

	expected1 := NewPacket(expectedSubifname1, expectedData1)
	expected2 := NewPacket(expectedSubifname2, expectedData2)

	actual := NewBulkPackets([]*Packet{expected1, expected2})

	suite.Equal(2, len(actual.Packets))
	suite.Equal(expectedSubifname1, actual.Packets[0].Subifname)
	suite.Equal(expectedData1, actual.Packets[0].Data)
	suite.Equal(expectedSubifname2, actual.Packets[1].Subifname)
	suite.Equal(expectedData2, actual.Packets[1].Data)
}

func (suite *testPacketTestSuite) TestNewBulkPackets2() {
	expectedSubifname1 := "vif1"
	expectedData1 := []byte{0x01, 0x00, 0x5e, 0x00, 0x00, 0x12}

	expectedSubifname2 := "vif2"
	expectedData2 := []byte{0x00, 0x01, 0x00, 0x5e, 0x12, 0x00}

	p1 := &hostif.Packet{
		Subifname: expectedSubifname1,
		Len:       uint32(len(expectedData1)),
		Data:      expectedData1,
	}

	p2 := &hostif.Packet{
		Subifname: expectedSubifname2,
		Len:       uint32(len(expectedData2)),
		Data:      expectedData2,
	}

	hbps := &hostif.BulkPackets{
		N:       2,
		Packets: []*hostif.Packet{p1, p2},
	}

	bps := newBulkPackets(hbps)

	actual := bps.bulkpackets

	suite.Equal(2, len(actual.Packets))
	suite.Equal(expectedSubifname1, actual.Packets[0].Subifname)
	suite.Equal(expectedData1, actual.Packets[0].Data)
	suite.Equal(expectedSubifname2, actual.Packets[1].Subifname)
	suite.Equal(expectedData2, actual.Packets[1].Data)
}

func TestPacketTestSuite(t *testing.T) {
	suite.Run(t, new(testPacketTestSuite))
}
