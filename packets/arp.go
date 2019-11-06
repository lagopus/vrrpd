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

	"github.com/google/gopacket"
	glayers "github.com/google/gopacket/layers"
)

const (
	// ARPHwAddressSize size of MAC addr
	ARPHwAddressSize = 6
)

var (
	// BroadcastMAC broadcast mac address
	BroadcastMAC = net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

// SerializeARP Serialize ARP
func SerializeARP(ip net.IP, mac net.HardwareAddr) ([]byte, error) {
	// ethernet
	ethernet := &glayers.Ethernet{
		DstMAC:       BroadcastMAC,
		SrcMAC:       mac,
		EthernetType: glayers.EthernetTypeARP,
	}

	// arp
	arp := &glayers.ARP{
		AddrType:          glayers.LinkTypeEthernet,
		Protocol:          glayers.EthernetTypeIPv4,
		HwAddressSize:     ARPHwAddressSize,
		ProtAddressSize:   net.IPv4len,
		Operation:         glayers.ARPRequest,
		SourceHwAddress:   mac,
		SourceProtAddress: ip.To4(),
		DstHwAddress:      BroadcastMAC,
		DstProtAddress:    ip.To4(),
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths: true,
	}

	if err := gopacket.SerializeLayers(buf, opts,
		ethernet, arp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SerializeVirtualMacARP Serialize ARP
func SerializeVirtualMacARP(vrid uint8, ip net.IP) ([]byte, error) {
	// virtual router mac address
	vmac := net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x01, vrid}

	return SerializeARP(ip, vmac)
}
