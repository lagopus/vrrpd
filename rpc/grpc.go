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
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Connection gRPC connection.
type Connection struct {
	addr       string
	port       int
	conn       *grpc.ClientConn
	cancelFunc context.CancelFunc
}

// NewConnection New gRPC connection.
func NewConnection(addr string, port int) *Connection {
	c := &Connection{
		addr:       addr,
		port:       port,
		cancelFunc: nil,
	}

	return c
}

// Connect connect gRPC server.
func (c *Connection) Connect() error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithBackoffMaxDelay(ConnectInterval),
	}
	ad := fmt.Sprintf("%s:%d", c.addr, c.port)

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	var conn *grpc.ClientConn
	var err error
	if conn, err = grpc.DialContext(ctx, ad, opts...); err != nil {
		return err
	}

	c.conn = conn

	return nil
}

// Disconnect disconnect gRPC server.
func (c *Connection) Disconnect() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	if c.conn != nil {
		_ = c.conn.Close()
	}
}
