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

package logger

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Formatter Logger formatter.
type Formatter struct {
	TimestampFormat string
}

// Format Format.
func (f *Formatter) Format(entry *log.Entry) ([]byte, error) {
	var buf *bytes.Buffer

	if entry.Buffer == nil {
		buf = &bytes.Buffer{}
	} else {
		buf = entry.Buffer
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	funcName := ""
	if len(entry.Caller.Function) != 0 {
		s := strings.Split(entry.Caller.Function, "/")
		funcName = s[len(s)-1]
	}
	fmt.Fprintf(buf, "[%s][%s]:%s:%d %s: %s",
		entry.Time.Format(timestampFormat),
		strings.ToUpper(entry.Level.String()),
		entry.Caller.File,
		entry.Caller.Line,
		funcName,
		entry.Message)

	buf.WriteByte('\n')

	return buf.Bytes(), nil
}
