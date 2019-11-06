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

package layers

import (
	"net"
	"testing"

	"github.com/google/gopacket"
	glayers "github.com/google/gopacket/layers"
	"github.com/stretchr/testify/suite"
)

type testVRRPv3AdvTestSuite struct {
	suite.Suite
}

func (suite *testVRRPv3AdvTestSuite) TestSerializeVRRPv3Adv() {
	expectedPacket := []byte{
		//    L2 header
		//<------------------------------------------------------------------>
		0x01, 0x00, 0x5e, 0x00, 0x00, 0x12, 0x52, 0x54, 0x00, 0xdc, 0x65, 0x98,
		//    L3 header
		//<------------------------------------------------------------------>
		0x08, 0x00, 0x45, 0xc0, 0x00, 0x24, 0x00, 0x01, 0x00, 0x00, 0xff, 0x70,
		//------------------------------------------------------->
		//                                                          VRRP
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

	ethernet := &glayers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x52, 0x54, 0x00, 0xdc, 0x65, 0x98},
		DstMAC:       net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0x12},
		EthernetType: glayers.EthernetTypeIPv4,
	}
	ip := &glayers.IPv4{
		Version:    4,
		IHL:        5,
		TOS:        0xc0,
		Length:     36,
		Id:         1,
		Flags:      0,
		FragOffset: 0,
		TTL:        255,
		Protocol:   112,
		SrcIP:      net.IP{192, 168, 100, 238},
		DstIP:      net.IP{224, 0, 0, 18},
	}
	vrrp := &VRRPv3Adv{
		Version:      VRRPv3Version,
		Type:         VRRPv3Advertisement,
		VirtualRtrID: 50,
		Priority:     255,
		CountIPAddr:  2,
		MaxAdverInt:  100,
		IPAddress: []net.IP{
			net.IP{10, 0, 0, 1},
			net.IP{10, 0, 0, 2}},
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	vrrp.SetNetworkLayerForChecksum(ip)
	err := gopacket.SerializeLayers(buf, opts,
		ethernet,
		ip,
		vrrp)
	suite.Empty(err)
	suite.Equal(expectedPacket, buf.Bytes())
}

func (suite *testVRRPv3AdvTestSuite) TestSerializeVRRPv3AdvErrorBadVersion() {
	vrrp := &VRRPv3Adv{
		Version:      4, // (4)
		Type:         VRRPv3Advertisement,
		VirtualRtrID: 50,
		Priority:     255,
		CountIPAddr:  2,
		MaxAdverInt:  100,
		IPAddress: []net.IP{
			net.IP{10, 0, 0, 1},
			net.IP{10, 0, 0, 2}},
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts,
		vrrp)
	suite.EqualError(err, "Bad version or type: version = 4, type = 1")
}

func (suite *testVRRPv3AdvTestSuite) TestSerializeVRRPv3AdvErrorBadtype() {
	vrrp := &VRRPv3Adv{
		Version:      VRRPv3Version,
		Type:         2, // (2)
		VirtualRtrID: 50,
		Priority:     255,
		CountIPAddr:  2,
		MaxAdverInt:  100,
		IPAddress: []net.IP{
			net.IP{10, 0, 0, 1},
			net.IP{10, 0, 0, 2}},
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	err := gopacket.SerializeLayers(buf, opts,
		vrrp)
	suite.EqualError(err, "Bad version or type: version = 3, type = 2")
}

func (suite *testVRRPv3AdvTestSuite) TestVRRPv3AdvDecode() {
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
	expectedVRRP := &VRRPv3Adv{
		Version:      3,
		Type:         VRRPv3Advertisement,
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
		//                                                          VRRP
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

	var eth glayers.Ethernet
	var ip glayers.IPv4
	var vrrp VRRPv3Adv

	parser := gopacket.NewDecodingLayerParser(glayers.LayerTypeEthernet, &eth, &ip, &vrrp)
	decoded := []gopacket.LayerType{}
	err := parser.DecodeLayers(packet, &decoded)
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
	suite.Equal(expectedVRRP.Version, vrrp.Version)
	suite.Equal(expectedVRRP.Type, vrrp.Type)
	suite.Equal(expectedVRRP.VirtualRtrID, vrrp.VirtualRtrID)
	suite.Equal(expectedVRRP.Priority, vrrp.Priority)
	suite.Equal(expectedVRRP.CountIPAddr, vrrp.CountIPAddr)
	suite.Equal(expectedVRRP.MaxAdverInt, vrrp.MaxAdverInt)
	suite.Equal(expectedVRRP.Checksum, vrrp.Checksum)
	suite.Equal(expectedVRRP.IPAddress, vrrp.IPAddress)
}

func (suite *testVRRPv3AdvTestSuite) TestDecodeVRRPv3AdvErrorBadVersion() {
	packet := []byte{
		//version(4)/type
		//<>  VRID
		//    <-->
		0x41, 0x32,
		//Priority
		//<>   Count IPvX Addr
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x02, 0x00, 0x64, 0xb5, 0x39, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->
		0x00, 0x02,
	}

	var vrrp VRRPv3Adv

	parser := gopacket.NewDecodingLayerParser(LayerTypeVRRPv3Adv, &vrrp)
	decoded := []gopacket.LayerType{}
	err := parser.DecodeLayers(packet, &decoded)
	suite.EqualError(err, "Bad version or type: version = 4, type = 1")
}

func (suite *testVRRPv3AdvTestSuite) TestDecodeVRRPv3AdvErrorBadType() {
	packet := []byte{
		//version/type(2)
		//<>  VRID
		//    <-->
		0x32, 0x32,
		//Priority
		//<>   Count IPvX Addr
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x02, 0x00, 0x64, 0xb5, 0x39, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->
		0x00, 0x02,
	}

	var vrrp VRRPv3Adv

	parser := gopacket.NewDecodingLayerParser(LayerTypeVRRPv3Adv, &vrrp)
	decoded := []gopacket.LayerType{}
	err := parser.DecodeLayers(packet, &decoded)
	suite.EqualError(err, "Bad version or type: version = 3, type = 2")
}

func (suite *testVRRPv3AdvTestSuite) TestDecodeVRRPv3AdvErrorBadCountIPAddr() {
	packet := []byte{
		//version/type
		//<>  VRID
		//    <-->
		0x31, 0x32,
		//Priority
		//<>   Count IPvX Addr(0)
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4
		//                                  <-------------------->  <--------
		0xff, 0x00, 0x00, 0x64, 0xb5, 0x39, 0x0a, 0x00, 0x00, 0x01, 0x0a, 0x00,
		//------->
		0x00, 0x02,
	}

	var vrrp VRRPv3Adv

	parser := gopacket.NewDecodingLayerParser(LayerTypeVRRPv3Adv, &vrrp)
	decoded := []gopacket.LayerType{}
	err := parser.DecodeLayers(packet, &decoded)
	suite.EqualError(err, "VRRPv3 number of IP addresses is not valid")
}

func (suite *testVRRPv3AdvTestSuite) TestDecodeVRRPv3AdvErrorBadIPAddrs() {
	packet := []byte{
		//    L2 header
		//<------------------------------------------------------------------>
		0x01, 0x00, 0x5e, 0x00, 0x00, 0x12, 0x52, 0x54, 0x00, 0xdc, 0x65, 0x98,
		//    L3 header
		//<------------------------------------------------------------------>
		0x08, 0x00, 0x45, 0xc0, 0x00, 0x25, 0x00, 0x01, 0x00, 0x00, 0xff, 0x70,
		//------------------------------------------------------->
		//                                                          VRRP
		//                                                          version/type
		//                                                          <-->  VRID
		//                                                                <-->
		0xb4, 0xff, 0xc0, 0xa8, 0x64, 0xee, 0xe0, 0x00, 0x00, 0x12, 0x31, 0x32,
		//Priority
		//<>   Count IPvX Addr
		//    <-->  rsvd/Max Adver Int
		//          <-------->   Checksum
		//                      <-------->    IPv4                    IPv4(bad)
		//                                  <-------------------->  <--------
		0xff, 0x02, 0x00, 0x64, 0xbf, 0x3f, 0x0a, 0x00, 0x00, 0x01, 0xff, 0xff,
		//------------->  <-- pad
		0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	var eth glayers.Ethernet
	var ip glayers.IPv4
	var vrrp VRRPv3Adv

	parser := gopacket.NewDecodingLayerParser(glayers.LayerTypeEthernet, &eth, &ip, &vrrp)
	decoded := []gopacket.LayerType{}
	err := parser.DecodeLayers(packet, &decoded)
	suite.EqualError(err, "VRRPv3 length of IP addresses is not valid")
}

func TestVRRPv3AdvTestSuites(t *testing.T) {
	suite.Run(t, new(testVRRPv3AdvTestSuite))
}
