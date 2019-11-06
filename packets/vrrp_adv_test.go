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

	glayers "github.com/google/gopacket/layers"
	"github.com/lagopus/vrrpd/packets/layers"
	"github.com/stretchr/testify/suite"
)

type testVRRPAdvTestSuite struct {
	suite.Suite
}

func (suite *testVRRPAdvTestSuite) TestSerializeVRRPAdv() {
	expectedPacket := []byte{
		//    L2 header
		//<------------------------------------------------------------------>
		0x01, 0x00, 0x5e, 0x00, 0x00, 0x12, 0x00, 0x00, 0x5e, 0x00, 0x01, 0x32,
		//    L3 header
		//<------------------------------------------------------------------>
		0x08, 0x00, 0x45, 0x00, 0x00, 0x24, 0x00, 0x00, 0x00, 0x00, 0xff, 0x70,
		//------------------------------------------------------->
		//                                                          VRRPAdv
		//                                                          version/type
		//                                                          <-->  VRID
		//
		0xb5, 0xc0, 0xc0, 0xa8, 0x64, 0xee, 0xe0, 0x00, 0x00, 0x12, 0x31, 0x32,
		//Priority
		//<>   Count IPvX Addr
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x02, 0x00, 0x64, 0xb5, 0x39, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->  <-- pad
		0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	vrrp := &layers.VRRPv3Adv{
		VirtualRtrID: 50,
		Priority:     255,
		MaxAdverInt:  100,
		IPAddress: []net.IP{
			net.IP{10, 0, 0, 1},
			net.IP{10, 0, 0, 2}},
	}

	buf, err := SerializeVRRPAdv(
		net.IP{192, 168, 100, 238},
		vrrp)
	suite.Empty(err)

	// ignore IPv4-Id(random)/chksum
	suite.Equal(expectedPacket, buf)
}

func (suite *testVRRPAdvTestSuite) TestSerializeVRRPAdvErrorInvalidArgs() {
	_, err := SerializeVRRPAdv(
		net.IP{192, 168, 100, 238},
		nil)
	suite.EqualError(err, "Invalid args")
}

func (suite *testVRRPAdvTestSuite) TestDecodeVRRPAdv() {
	expectedEthernet := &glayers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x52, 0x54, 0x00, 0xdc, 0x65, 0x98},
		DstMAC:       net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0x12},
		EthernetType: glayers.EthernetTypeIPv4,
	}
	expectedIP := &glayers.IPv4{
		Version:    4,
		IHL:        5,
		TOS:        0xc0,
		Length:     36,
		Id:         1,
		Flags:      0,
		FragOffset: 0,
		TTL:        255,
		Protocol:   112,
		Checksum:   0xb4ff,
		SrcIP:      net.IP{192, 168, 100, 238},
		DstIP:      net.IP{224, 0, 0, 18},
	}
	expectedVRRPAdv := &layers.VRRPv3Adv{
		Version:      3,
		Type:         layers.VRRPv3Advertisement,
		VirtualRtrID: 50,
		Priority:     255,
		CountIPAddr:  2,
		MaxAdverInt:  100,
		Checksum:     0xb539,
		IPAddress: []net.IP{
			net.IP{10, 0, 0, 1},
			net.IP{10, 0, 0, 2}},
	}

	packet := []byte{
		//    L2 header
		//<------------------------------------------------------------------>
		0x01, 0x00, 0x5e, 0x00, 0x00, 0x12, 0x52, 0x54, 0x00, 0xdc, 0x65, 0x98,
		//    L3 header
		//<------------------------------------------------------------------>
		0x08, 0x00, 0x45, 0xc0, 0x00, 0x24, 0x00, 0x01, 0x00, 0x00, 0xff, 0x70,
		//------------------------------------------------------->
		//                                                          VRRPAdv
		//                                                          version/type
		//                                                          <-->  VRID
		//                                                                <-->
		0xb4, 0xff, 0xc0, 0xa8, 0x64, 0xee, 0xe0, 0x00, 0x00, 0x12, 0x31, 0x32,
		//Priority
		//<>   Count IPvX Addr
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x02, 0x00, 0x64, 0xb5, 0x39, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->  <-- pad
		0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	eth, ip, vrrp, err := DecodeVRRPAdv(packet)
	suite.Empty(err)

	// ethernet
	suite.Equal(expectedEthernet.SrcMAC, eth.SrcMAC)
	suite.Equal(expectedEthernet.DstMAC, eth.DstMAC)
	suite.Equal(expectedEthernet.EthernetType, eth.EthernetType)
	suite.Equal(expectedEthernet.Length, eth.Length)

	// ipv4
	suite.Equal(expectedIP.Version, ip.Version)
	suite.Equal(expectedIP.IHL, ip.IHL)
	suite.Equal(expectedIP.TOS, ip.TOS)
	suite.Equal(expectedIP.Length, ip.Length)
	suite.Equal(expectedIP.Id, ip.Id)
	suite.Equal(expectedIP.Flags, ip.Flags)
	suite.Equal(expectedIP.FragOffset, ip.FragOffset)
	suite.Equal(expectedIP.TTL, ip.TTL)
	suite.Equal(expectedIP.Protocol, ip.Protocol)
	suite.Equal(expectedIP.Checksum, ip.Checksum)
	suite.Equal(expectedIP.SrcIP, ip.SrcIP)
	suite.Equal(expectedIP.DstIP, ip.DstIP)

	// vrrp
	suite.Equal(expectedVRRPAdv.Version, vrrp.Version)
	suite.Equal(expectedVRRPAdv.Type, vrrp.Type)
	suite.Equal(expectedVRRPAdv.VirtualRtrID, vrrp.VirtualRtrID)
	suite.Equal(expectedVRRPAdv.Priority, vrrp.Priority)
	suite.Equal(expectedVRRPAdv.CountIPAddr, vrrp.CountIPAddr)
	suite.Equal(expectedVRRPAdv.MaxAdverInt, vrrp.MaxAdverInt)
	suite.Equal(expectedVRRPAdv.Checksum, vrrp.Checksum)
	suite.Equal(expectedVRRPAdv.IPAddress, vrrp.IPAddress)
}

func (suite *testVRRPAdvTestSuite) TestDecodeVRRPAdvErrorBadChecksum() {
	packet := []byte{
		//    L2 header
		//<------------------------------------------------------------------>
		0x01, 0x00, 0x5e, 0x00, 0x00, 0x12, 0x52, 0x54, 0x00, 0xdc, 0x65, 0x98,
		//    L3 header
		//<------------------------------------------------------------------>
		0x08, 0x00, 0x45, 0xc0, 0x00, 0x24, 0x00, 0x01, 0x00, 0x00, 0xff, 0x70,
		//------------------------------------------------------->
		//                                                          VRRPAdv
		//                                                          version/type
		//                                                          <-->  VRID
		//                                                                <-->
		0xb4, 0xff, 0xc0, 0xa8, 0x64, 0xee, 0xe0, 0x00, 0x00, 0x12, 0x31, 0x32,
		//Priority
		//<>   Count IPvX Addr
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum(bad)
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x02, 0x00, 0x64, 0xb5, 0x38, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->  <-- pad
		0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	_, _, _, err := DecodeVRRPAdv(packet)
	suite.EqualError(err, "Bad checksum")
}

func (suite *testVRRPAdvTestSuite) TestDecodeVRRPAdvErrorBadTTL() {
	packet := []byte{
		//    L2 header
		//<------------------------------------------------------------------>
		0x01, 0x00, 0x5e, 0x00, 0x00, 0x12, 0x52, 0x54, 0x00, 0xdc, 0x65, 0x98,
		//    L3 header                                             <---> TTL(bad)
		//<------------------------------------------------------------------>
		0x08, 0x00, 0x45, 0xc0, 0x00, 0x24, 0x00, 0x01, 0x00, 0x00, 0xfe, 0x70,
		//------------------------------------------------------->
		//                                                          VRRPAdv
		//                                                          version/type
		//                                                          <-->  VRID
		//                                                                <-->
		0xb4, 0xff, 0xc0, 0xa8, 0x64, 0xee, 0xe0, 0x00, 0x00, 0x12, 0x31, 0x32,
		//Priority
		//<>   Count IPvX Addr
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x02, 0x00, 0x64, 0xb5, 0x39, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->  <-- pad
		0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	_, _, _, err := DecodeVRRPAdv(packet)
	suite.EqualError(err, "Bad TTL: 254")
}

func (suite *testVRRPAdvTestSuite) TestDecodeVRRPAdvErrorBadCountIPAddr() {
	packet := []byte{
		//    L2 header
		//<------------------------------------------------------------------>
		0x01, 0x00, 0x5e, 0x00, 0x00, 0x12, 0x52, 0x54, 0x00, 0xdc, 0x65, 0x98,
		//    L3 header
		//<------------------------------------------------------------------>
		0x08, 0x00, 0x45, 0xc0, 0x00, 0x24, 0x00, 0x01, 0x00, 0x00, 0xff, 0x70,
		//------------------------------------------------------->
		//                                                          VRRPAdv
		//                                                          version/type
		//                                                          <-->  VRID
		//                                                                <-->
		0xb4, 0xff, 0xc0, 0xa8, 0x64, 0xee, 0xe0, 0x00, 0x00, 0x12, 0x31, 0x32,
		//Priority
		//<>   Count IPvX Addr(bad)
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x03, 0x00, 0x64, 0xb5, 0x38, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->  <-- pad
		0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	_, _, _, err := DecodeVRRPAdv(packet)
	suite.EqualError(err, "Bad CountIPAddr. CountIPAddr: 3, IPAddress: 2")
}

func TestVRRPAdvTestSuites(t *testing.T) {
	suite.Run(t, new(testVRRPAdvTestSuite))
}
