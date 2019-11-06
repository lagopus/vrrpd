#!/usr/bin/env python3

#
# Copyright 2017 Nippon Telegraph and Telephone Corporation.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

import socket
import time
import logging
from const import *
from argparse import ArgumentParser
from argparse import ArgumentDefaultsHelpFormatter
from scapy.all import *

class VRRP(Packet):
    # VRRP v3.
    # https://tools.ietf.org/html/rfc5798#section-5.1
    fields_desc = [
        BitField('version' , 3, 4),
        BitField('type' , 1, 4),
        ByteField('vrid', 1),
        ByteField('priority', 100),
        FieldLenField('ipcount', None, count_of='addrlist', fmt='B'),
        XShortField('maxadvint', 1),
        XShortField('chksum', None),
        FieldListField('addrlist', [], IPField('', '0.0.0.0'),
                       count_from = lambda pkt: pkt.ipcount),]

    def post_build(self, p, pay):
        if self.underlayer.len is not None:
            ln = self.underlayer.len - 20
        else:
            ln = len(p)
            psdhdr = struct.pack('!4s4sHH',
                                 inet_aton(self.underlayer.src),
                                 inet_aton(self.underlayer.dst),
                                 self.underlayer.proto,
                                 ln)
            ck = checksum(psdhdr + p)
            p = p[:6] + struct.pack('!H', ck) + p[8:]
        return p

bind_layers(IP, VRRP, proto=IPPROTO_VRRP)


DEFAULT_SRC_IP = '192.168.200.1'
DEFAULT_VRID = 1
DEFAULT_PRIORITY = 200
DEFAULT_MAX_ADV_INT = 100
DEFAULT_VIP_ADDRS = ['192.168.200.100']

def parse_opts():
    parser = ArgumentParser(description='send VRRP adv.',
                            formatter_class=ArgumentDefaultsHelpFormatter)

    parser.add_argument('--srcip',
                        type=str,
                        default=DEFAULT_SRC_IP,
                        help='Src IP addr')

    parser.add_argument('--vrid',
                        type=int,
                        default=DEFAULT_VRID,
                        help='VRID')

    parser.add_argument('--priority',
                        type=int,
                        default=DEFAULT_PRIORITY,
                        help='Priority')

    parser.add_argument('--advint',
                        type=int,
                        default=DEFAULT_MAX_ADV_INT,
                        help='Max adv int')

    parser.add_argument('--vips',
                        type=str,
                        nargs='+',
                        default=DEFAULT_VIP_ADDRS,
                        help='Virtual IPs')

    opts = parser.parse_args()

    return opts


def main():
    logging.basicConfig(level=logging.INFO)
    opts = parse_opts()

    p = Ether(dst='01:00:5e:00:00:12', src='00:00:5e:00:01:%02x' % opts.vrid)/ \
        IP(dst='224.0.0.18', src=opts.srcip, ttl=255)/ \
        VRRP(vrid=opts.vrid, priority=opts.priority,
             maxadvint=opts.advint, addrlist=opts.vips)

    rawsock = socket.socket(socket.AF_PACKET, socket.SOCK_RAW)
    for i in range(100):
        try:
            rawsock.bind((TAP_IN_NAME, 0))
            break
        except:
            logging.info('retry: %d' % (i + 1))
            time.sleep(0.5)

    logging.info('sending...')
    while True:
        rawsock.send(bytes(p))
        time.sleep(1)


if __name__ == '__main__':
    main()
