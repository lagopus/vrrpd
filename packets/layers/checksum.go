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
// - https://github.com/google/seesaw
// - https://github.com/google/gopacket

package layers

import (
	"fmt"

	"github.com/google/gopacket"
	glayers "github.com/google/gopacket/layers"
)

type checksumPseudoheader struct {
	pseudoheader interface{}
}

func pseudoheaderIPv4(ip *glayers.IPv4) (csum uint32, err error) {
	if err := ip.AddressTo4(); err != nil {
		return 0, err
	}

	csum += (uint32(ip.SrcIP[0]) + uint32(ip.SrcIP[2])) << 8
	csum += uint32(ip.SrcIP[1]) + uint32(ip.SrcIP[3])
	csum += (uint32(ip.DstIP[0]) + uint32(ip.DstIP[2])) << 8
	csum += uint32(ip.DstIP[1]) + uint32(ip.DstIP[3])

	return csum, nil
}

func ipChecksum(b []byte, csum uint32) uint16 {
	for ; len(b) >= 2; b = b[2:] {
		csum += uint32(b[0])<<8 | uint32(b[1])
	}
	if len(b) == 1 {
		csum += uint32(b[0]) << 8
	}
	for csum>>16 != 0 {
		csum = (csum & 0xffff) + (csum >> 16)
	}
	return uint16(^csum)
}

// public funcs

// ComputeChecksum compute checksum
func (c *checksumPseudoheader) ComputeChecksum(headerAndPayload []byte,
	headerProtocol glayers.IPProtocol) (uint16, error) {
	if c.pseudoheader == nil {
		return 0, fmt.Errorf("bad pseudoheader")
	}
	length := uint32(len(headerAndPayload))

	var csum uint32
	var err error
	switch v := c.pseudoheader.(type) {
	case *glayers.IPv4:
		csum, err = pseudoheaderIPv4(v)
		if err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("cannot use layer type for tcp checksum network layer")
	}

	csum += uint32(headerProtocol)
	csum += length & 0xffff
	csum += length >> 16

	return ipChecksum(headerAndPayload, csum), nil
}

// SetNetworkLayerForChecksum set checksum flag
func (c *checksumPseudoheader) SetNetworkLayerForChecksum(l gopacket.NetworkLayer) {
	c.pseudoheader = l
}
