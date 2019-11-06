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

package packets

import (
	"net"
	"testing"

	"github.com/lagopus/vrrpd/packets/layers"
	"github.com/stretchr/testify/suite"
)

type testARPTestSuite struct {
	suite.Suite
}

func (suite *testARPTestSuite) TestSerializeARP() {
	expectedPacket := []byte{
		//    L2 header
		//<-------------------------------------------------------------------
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x52, 0x54, 0x00, 0xce, 0xd1, 0xa3,
		//------->  <-- ARP --
		//            HTYPE       PTYPE     HLEN  PLEN    OPER
		//          <-------->  <-------->  <-->  <-->  <-------->  <---------
		0x08, 0x06, 0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x01, 0x52, 0x54,
		// SHA                    SPA
		//------------------->  <-------------------->  <---------------------
		0x00, 0xce, 0xd1, 0xa3, 0x0a, 0x00, 0x00, 0x01, 0xff, 0xff, 0xff, 0xff,
		// THA       TPA                      pad
		//------->  <-------------------->  <----
		0xff, 0xff, 0x0a, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	vrrp := &layers.VRRPv3Adv{
		VirtualRtrID: 50,
		IPAddress: []net.IP{
			net.IP{10, 0, 0, 1}},
	}

	buf, err := SerializeARP(vrrp.IPAddress[0],
		net.HardwareAddr{0x52, 0x54, 0x00, 0xce, 0xd1, 0xa3})
	suite.Empty(err)

	suite.Equal(expectedPacket, buf)
}

func (suite *testARPTestSuite) TestSerializeVirtualMacARP() {
	expectedPacket := []byte{
		//    L2 header
		//<-------------------------------------------------------------------
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x5e, 0x00, 0x01, 0x32,
		//------->  <-- ARP --
		//            HTYPE       PTYPE     HLEN  PLEN    OPER
		//          <-------->  <-------->  <-->  <-->  <-------->  <---------
		0x08, 0x06, 0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x01, 0x00, 0x00,
		// SHA                    SPA
		//------------------->  <-------------------->  <---------------------
		0x5e, 0x00, 0x01, 0x32, 0x0a, 0x00, 0x00, 0x01, 0xff, 0xff, 0xff, 0xff,
		// THA       TPA                      pad
		//------->  <-------------------->  <----
		0xff, 0xff, 0x0a, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	vrrp := &layers.VRRPv3Adv{
		VirtualRtrID: 50,
		IPAddress: []net.IP{
			net.IP{10, 0, 0, 1}},
	}

	buf, err := SerializeVirtualMacARP(vrrp.VirtualRtrID, vrrp.IPAddress[0])
	suite.Empty(err)

	suite.Equal(expectedPacket, buf)
}

func TestARPTestSuite(t *testing.T) {
	suite.Run(t, new(testARPTestSuite))
}
