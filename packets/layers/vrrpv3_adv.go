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

// The following was in reference.
// - https://github.com/google/gopacket

package layers

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"

	"github.com/google/gopacket"
	glayers "github.com/google/gopacket/layers"
)

/*
	This layer provides decoding for Virtual Router Redundancy Protocol (VRRP) v3.
	https://tools.ietf.org/html/rfc5798#section-5
     0                   1                   2                   3
     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |Version| Type  | Virtual Rtr ID|   Priority    |Count IPvX Addr|
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |(rsvd) |     Max Adver Int     |          Checksum             |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                                                               |
    +                                                               +
    |                       IPvX Address(es)                        |
    +                                                               +
    +                                                               +
    +                                                               +
    +                                                               +
    |                                                               |
    +                                                               +
    |                                                               |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/

// VRRPv3Type type
type VRRPv3Type uint8

const (
	// VRRPv3Version Version
	VRRPv3Version = 0x03
	// VRRPv3Advertisement Advertisement type
	VRRPv3Advertisement VRRPv3Type = 0x01
	// VRRPv3LayerType Custom LayerType for VRRP
	VRRPv3LayerType = 2001
)

// VRRPv3Adv represents an VRRP v3 message.
type VRRPv3Adv struct {
	glayers.BaseLayer
	checksumPseudoheader
	Version      uint8      // The version field specifies the VRRP protocol version of this packet (v3)
	Type         VRRPv3Type // The type field specifies the type of this VRRP packet.  The only type defined in v3 is ADVERTISEMENT
	VirtualRtrID uint8      // identifies the virtual router this packet is reporting status for
	Priority     uint8      // specifies the sending VRRP router's priority for the virtual router (100 = default)
	CountIPAddr  uint8      // The number of IP addresses contained in this VRRP advertisement.
	MaxAdverInt  uint16     // The Advertisement interval indicates the time interval (in centiseconds) between ADVERTISEMENTS.  The default is 100 second
	Checksum     uint16     // used to detect data corruption in the VRRP message.
	IPAddress    []net.IP   // one or more IP addresses associated with the virtual router. Specified in the CountIPAddr field.
}

var (
	// LayerTypeVRRPv3Adv LayerType
	LayerTypeVRRPv3Adv = gopacket.RegisterLayerType(
		VRRPv3LayerType,
		gopacket.LayerTypeMetadata{
			Name:    "VRRPv3",
			Decoder: gopacket.DecodeFunc(decodeVRRPv3Adv),
		},
	)
)

func init() {
	glayers.IPProtocolMetadata[glayers.IPProtocolVRRP] =
		glayers.EnumMetadata{DecodeWith: gopacket.DecodeFunc(decodeVRRPv3Adv),
			Name: "VRRPv3", LayerType: LayerTypeVRRPv3Adv}
}

func decodeVRRPv3Adv(data []byte, p gopacket.PacketBuilder) error {
	if len(data) < 8 {
		return fmt.Errorf("Not a valid VRRP packet. Packet length is too small")
	}

	v := &VRRPv3Adv{}

	if err := v.DecodeFromBytes(data, p); err == nil {
		return err
	}

	p.AddLayer(v)

	return nil
}

// public funcs

// SerializeTo VRRP Serializer.
func (v *VRRPv3Adv) SerializeTo(b gopacket.SerializeBuffer, opts gopacket.SerializeOptions) error {
	if v.Version != VRRPv3Version ||
		VRRPv3Advertisement != v.Type {
		return fmt.Errorf("Bad version or type: version = %d, type = %d",
			v.Version, v.Type)
	}

	var bytes []byte
	var err error
	if bytes, err = b.PrependBytes(8 + (net.IPv4len * len(v.IPAddress))); err != nil {
		return err
	}

	bytes[0] = (v.Version << 4) | uint8(v.Type)
	bytes[1] = v.VirtualRtrID
	bytes[2] = v.Priority
	bytes[3] = v.CountIPAddr
	binary.BigEndian.PutUint16(bytes[4:], v.MaxAdverInt)
	bytes[6] = 0 // checksum
	bytes[7] = 0
	for i, ip := range v.IPAddress {
		head := 8 + (i * net.IPv4len)
		copy(bytes[head:head+net.IPv4len], ip.To4())
	}

	var csum uint16
	if csum, err = v.ComputeChecksum(bytes, glayers.IPProtocolVRRP); err != nil {
		return err
	}
	binary.BigEndian.PutUint16(bytes[6:], csum)

	return nil
}

// LayerType return layer type
func (v *VRRPv3Adv) LayerType() gopacket.LayerType {
	return LayerTypeVRRPv3Adv
}

// DecodeFromBytes decoder
func (v *VRRPv3Adv) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	v.BaseLayer = glayers.BaseLayer{Contents: data[:len(data)]}

	v.Version = data[0] >> 4
	v.Type = VRRPv3Type(data[0] & 0x0F)

	if v.Version != VRRPv3Version ||
		VRRPv3Advertisement != v.Type {
		return fmt.Errorf("Bad version or type: version = %d, type = %d",
			v.Version, v.Type)
	}

	v.VirtualRtrID = data[1]
	v.Priority = data[2]

	v.CountIPAddr = data[3]
	if v.CountIPAddr < 1 {
		return fmt.Errorf("VRRPv3 number of IP addresses is not valid")
	}

	v.MaxAdverInt = binary.BigEndian.Uint16(data[4:])
	v.Checksum = binary.BigEndian.Uint16(data[6:])

	if len(data[8:]) < 1 && len(data[8:]) > math.MaxUint8 ||
		len(data[8:])%net.IPv4len != 0 {
		return fmt.Errorf("VRRPv3 length of IP addresses is not valid")
	}

	for i := 0; i < len(data[8:])/net.IPv4len; i++ {
		head := 8 + (i * net.IPv4len)
		v.IPAddress = append(v.IPAddress, data[head:head+net.IPv4len])
	}

	return nil
}

// NextLayerType return LayerTypeZero
func (v *VRRPv3Adv) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypeZero
}

// CanDecode return LayerType VRRPv3Adv
func (v *VRRPv3Adv) CanDecode() gopacket.LayerClass {
	return LayerTypeVRRPv3Adv
}
