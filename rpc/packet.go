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
	hostif "github.com/lagopus/vsw/modules/hostif/packets_io"
)

// Packet Packet.
type Packet struct {
	Subifname string
	Data      []byte
	hpacket   *hostif.Packet
}

// NewPacket New Packet.
func NewPacket(subifname string, data []byte) *Packet {
	r := &Packet{
		Subifname: subifname,
		Data:      data,
		hpacket: &hostif.Packet{
			Subifname: subifname,
			Len:       uint32(len(data)),
			Data:      data,
		},
	}
	return r
}

// BulkPackets BulkPackets.
type BulkPackets struct {
	Packets     []*Packet
	bulkpackets *hostif.BulkPackets
}

// NewBulkPackets New BulkPackets.
func NewBulkPackets(packets []*Packet) *BulkPackets {
	r := &BulkPackets{
		Packets:     []*Packet{},
		bulkpackets: &hostif.BulkPackets{},
	}

	for _, p := range packets {
		r.Packets = append(r.Packets, p)
		r.bulkpackets.Packets = append(r.bulkpackets.Packets, p.hpacket)
		r.bulkpackets.N++
	}

	return r
}

func newBulkPackets(p *hostif.BulkPackets) *BulkPackets {
	r := &BulkPackets{
		Packets:     []*Packet{},
		bulkpackets: &hostif.BulkPackets{},
	}

	for _, hpacket := range p.Packets {
		packet := NewPacket(hpacket.Subifname, hpacket.Data)
		r.Packets = append(r.Packets, packet)
		r.bulkpackets.Packets = append(r.bulkpackets.Packets, packet.hpacket)
		r.bulkpackets.N++
	}

	return r
}
