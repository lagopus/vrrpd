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
	"time"
)

const (
	// DatastoreModuleName Datastore module name.
	DatastoreModuleName = "DatastoreModule"

	// DPAgentModuleName DataPlane Agent module name.
	DPAgentModuleName = "DPAgentModule"

	// HostifModuleName Hostif module name.
	HostifModuleName = "HostifModule"

	// SendChannelSize Size of SendChannel.
	SendChannelSize = 1000

	// RecvChannelSize Size of RecvChannel.
	RecvChannelSize = 1000

	// RecvInterval interval(1ms)
	RecvInterval time.Duration = time.Duration(1) * time.Millisecond

	// ConnectInterval interval(1s)
	ConnectInterval time.Duration = time.Duration(1) * time.Second
)
