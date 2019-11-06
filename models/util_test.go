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

package models

import (
	"bytes"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testUtilTestSuite struct {
	suite.Suite
}

func (suite *testUtilTestSuite) TestDupIP() {
	srcIP1 := net.ParseIP("172.16.110.1")
	dstIP1 := dupIP(srcIP1)
	suite.True(bytes.Compare(srcIP1, dstIP1) == 0)

	srcIP2 := net.ParseIP("172.16.110.2").To4()
	dstIP2 := dupIP(srcIP2)
	suite.True(bytes.Compare(srcIP2, dstIP2) == 0)
}

func TestUtilTestSuite(t *testing.T) {
	suite.Run(t, new(testUtilTestSuite))
}
