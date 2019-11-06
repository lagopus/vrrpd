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
	"fmt"
	"net"

	"github.com/google/gopacket"
	glayers "github.com/google/gopacket/layers"
	"github.com/lagopus/vrrpd/packets/layers"
)

const (
	// VRRPAdvIPver IPv4
	VRRPAdvIPver = 4
	// VRRPAdvIHL IPv4 IHL
	VRRPAdvIHL = 5
	// VRRPAdvTTL TTL
	VRRPAdvTTL = 255
)

var (
	// VRRPAdvDstMAC dst MAC addr
	VRRPAdvDstMAC = net.HardwareAddr{0x01, 0x00, 0x5e, 0x00, 0x00, 0x12}
	// VRRPAdvDstIP dst IP addr
	VRRPAdvDstIP = net.IP{224, 0, 0, 18}
)

// SerializeVRRPAdv Serialize VRRPAdv
func SerializeVRRPAdv(srcIP net.IP,
	vrrp *layers.VRRPv3Adv) ([]byte, error) {
	if vrrp == nil {
		return nil, fmt.Errorf("Invalid args")
	}

	// virtual router mac address
	vmac := net.HardwareAddr{0x00, 0x00, 0x5e, 0x00, 0x01, vrrp.VirtualRtrID}

	// ethernet
	ethernet := &glayers.Ethernet{
		DstMAC:       VRRPAdvDstMAC,
		SrcMAC:       vmac,
		EthernetType: glayers.EthernetTypeIPv4,
	}

	// IP
	ip := &glayers.IPv4{
		Version:    VRRPAdvIPver,
		IHL:        VRRPAdvIHL,
		TOS:        0,
		Id:         0,
		Flags:      0,
		FragOffset: 0,
		TTL:        VRRPAdvTTL,
		Protocol:   glayers.IPProtocolVRRP,
		DstIP:      VRRPAdvDstIP,
		SrcIP:      srcIP,
	}

	// VRRPAdv
	// args: VirtualRtrID, MaxAdverInt, Priority, IPAddress
	vrrp.Version = layers.VRRPv3Version
	vrrp.Type = layers.VRRPv3Advertisement
	vrrp.CountIPAddr = uint8(len(vrrp.IPAddress))

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	vrrp.SetNetworkLayerForChecksum(ip)
	if err := gopacket.SerializeLayers(buf, opts,
		ethernet, ip, vrrp); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func checkVRRPAdv(ip *glayers.IPv4, vrrp *layers.VRRPv3Adv) error {
	// TTL
	if ip.TTL != VRRPAdvTTL {
		return fmt.Errorf("Bad TTL: %v", ip.TTL)
	}

	// CountIPAddr
	if vrrp.CountIPAddr != uint8(len(vrrp.IPAddress)) {
		return fmt.Errorf("Bad CountIPAddr. CountIPAddr: %v, IPAddress: %v",
			vrrp.CountIPAddr, uint8(len(vrrp.IPAddress)))
	}

	// checksum
	vrrp.SetNetworkLayerForChecksum(ip)
	if csum, err := vrrp.ComputeChecksum(vrrp.Contents, glayers.IPProtocolVRRP); err != nil {
		return err
	} else if csum != 0 {
		return fmt.Errorf("Bad checksum")
	}

	return nil
}

// DecodeVRRPAdv Decode VRRPAdv
func DecodeVRRPAdv(packet []byte) (*glayers.Ethernet, *glayers.IPv4, *layers.VRRPv3Adv, error) {
	var eth glayers.Ethernet
	var ip glayers.IPv4
	var vrrp layers.VRRPv3Adv

	parser := gopacket.NewDecodingLayerParser(glayers.LayerTypeEthernet,
		&eth, &ip, &vrrp)
	decoded := []gopacket.LayerType{}

	if err := parser.DecodeLayers(packet, &decoded); err != nil {
		return nil, nil, nil, err
	}

	if err := checkVRRPAdv(&ip, &vrrp); err != nil {
		return nil, nil, nil, err
	}

	return &eth, &ip, &vrrp, nil
}
