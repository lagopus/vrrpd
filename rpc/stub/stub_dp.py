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

import pio_pb2
import pio_pb2_grpc
import vrrp_pb2
import vrrp_pb2_grpc
import time
import grpc
import socket
import time
import threading
import logging
from const import *
from scapy.all import *
from concurrent import futures
from pytun import TunTapDevice, IFF_TAP

rawsock_out = None
rawsock_in = None

class StubHostIF(pio_pb2_grpc.packets_ioServicer):
    def __init__(self):
        pio_pb2_grpc.packets_ioServicer.__init__(self)
        self.lock = threading.Lock()
        self.send_pkts = []

    def send_bulk(self, request, context):
        logging.info('send_bulk: %s' % request)

        for p in request.packets:
            rawsock_out.send(p.data)
        return pio_pb2.result(r=0)

    def recv_bulk(self, request, context):
        #p = Ether(dst='01:00:5e:00:00:12',src='00:00:5e:00:01:32')/ \
        #    IP(ttl=255,src='192.168.122.202',dst='244.0.0.18')/ \
        #    VRRP(vrid=1,priority=100,maxadvint=101,addrlist=['10.0.0.1', '10.0.0.2'])
        #
        #b = bytes(p)
        #packet = pio_pb2.packet(port=1, len=len(b), data=b)
        pkts = []
        if self.send_pkts:
            print(self.send_pkts)
            self.lock.acquire()
            for p in self.send_pkts:
                pkt = pio_pb2.packet(subifname='if0-0', len=len(p), data=p)
                pkts.append(pkt)
            self.send_pkts = []
            self.lock.release()
        return pio_pb2.bulk_packets(n=len(pkts), packets=pkts)

    def add_send_pkts(self, pkt):
        self.lock.acquire()
        self.send_pkts.append(pkt)
        self.lock.release()

class StubDPAgent(vrrp_pb2_grpc.VrrpServicer):
    def GetVifInfo(self, request, context):
        logging.info('GetVifInfo: %s' % request)

        for e in request.entries:
            e.addr='52:54:00:00:00:01'
        return request

    def ToMaster(self, request, context):
        logging.info('ToMaster: %s' % request)
        return vrrp_pb2.Reply(code=0)

    def ToBackup(self, request, context):
        logging.info('ToBackup: %s' % request)
        return vrrp_pb2.Reply(code=0)

BUFSIZE = 4096
def read(stub_host_if):
    p = rawsock_in.recv(BUFSIZE)
    stub_host_if.add_send_pkts(p)

stop = False
def serve():
    stub_host_if_serv = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    stub_host_if = StubHostIF()
    pio_pb2_grpc.add_packets_ioServicer_to_server(
        stub_host_if, stub_host_if_serv)
    stub_host_if_serv.add_insecure_port('[::]:30020')
    stub_host_if_serv.start()

    stub_dp_agent_serv = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    stub_dp_agent = StubDPAgent()
    vrrp_pb2_grpc.add_VrrpServicer_to_server(
        stub_dp_agent, stub_dp_agent_serv)
    stub_dp_agent_serv.add_insecure_port('[::]:30010')
    stub_dp_agent_serv.start()

    try:
        while True:
            read(stub_host_if)
    except KeyboardInterrupt:
        global stop
        stop = True
        stub_host_if_serv.stop(0)
        stub_dp_agent_serv.stop(0)

def main():
    logging.basicConfig(level=logging.INFO)
    global rawsock_out
    global rawsock_in
    try:
        tap_out = TunTapDevice(name=TAP_OUT_NAME, flags=IFF_TAP)
        tap_in = TunTapDevice(name=TAP_IN_NAME, flags=IFF_TAP)
        tap_out.up()
        tap_in.up()

        rawsock_out = socket.socket(socket.AF_PACKET, socket.SOCK_RAW)
        rawsock_out.bind((TAP_OUT_NAME, 0))
        rawsock_in = socket.socket(socket.AF_PACKET, socket.SOCK_RAW, socket.htons(ETH_P_ALL))
        rawsock_in.bind((TAP_IN_NAME, 0))
        serve()
    finally:
        if rawsock_out:
            rawsock_out.close()
        if rawsock_in:
            rawsock_in.close()
        if tap_out:
            tap_out.down()
            tap_out.close()
        if tap_in:
            tap_in.down()
            tap_in.close()

if __name__ == '__main__':
    main()
