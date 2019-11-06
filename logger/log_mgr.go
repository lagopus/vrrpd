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
	"io/ioutil"
	"log/syslog"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
)

var logmgr = newLogMgr()

// LogType type.
type LogType uint8

const (
	// LogTypeFile Type of file.
	LogTypeFile LogType = iota
	// LogTypeSyslog Type of syslog.
	LogTypeSyslog
	// LogTypeStderr Type of stderr.
	LogTypeStderr
)

// LogMgr Logger
type LogMgr struct {
	logType  LogType
	logFile  string
	logLevel string
	file     *os.File
	lock     sync.Mutex
}

func newLogMgr() *LogMgr {
	l := &LogMgr{
		logType: LogTypeSyslog,
		file:    nil,
	}

	// format.
	f := new(Formatter)
	f.TimestampFormat = "2006-01-02 15:04:05.000000"
	log.SetFormatter(f)
	log.SetReportCaller(true)

	return l
}

func setLoglevel(logLevel string) error {
	if l, err := log.ParseLevel(logLevel); err == nil {
		log.SetLevel(l)
		logmgr.logLevel = logLevel
	} else {
		return err
	}

	return nil
}

func setSyslog(logLevel string) error {
	logmgr.logType = LogTypeSyslog
	p := syslog.LOG_EMERG | syslog.LOG_ALERT | syslog.LOG_CRIT |
		syslog.LOG_ERR | syslog.LOG_WARNING | syslog.LOG_NOTICE |
		syslog.LOG_INFO | syslog.LOG_DEBUG
	hook, err := logrus_syslog.NewSyslogHook("", "", p, "")
	if err != nil {
		return err
	}

	err = setLoglevel(logLevel)
	if err != nil {
		return err
	}

	log.SetOutput(ioutil.Discard)
	log.AddHook(hook)

	return nil
}

func setFilelog(logFile string,
	logLevel string) error {
	logmgr.logType = LogTypeFile

	f, err := os.OpenFile(
		logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	logmgr.logFile = logFile
	logmgr.logLevel = logLevel
	logmgr.file = f
	log.SetOutput(f)

	err = setLoglevel(logLevel)
	if err != nil {
		return err
	}

	return nil
}

func setStderr(logLevel string) error {
	logmgr.logType = LogTypeStderr
	return setLoglevel(logLevel)
}

// Set Set logger.
func Set(logFile string,
	logLevel string) error {
	logmgr.lock.Lock()
	defer logmgr.lock.Unlock()

	switch strings.ToLower(logFile) {
	case "syslog":
		return setSyslog(logLevel)
	case "stderr":
		return setStderr(logLevel)
	}

	return setFilelog(logFile, logLevel)
}

// Final Set Finalize logger.
func Final() {
	logmgr.lock.Lock()
	defer logmgr.lock.Unlock()

	if logmgr.logType == LogTypeFile && logmgr.file != nil {
		_ = logmgr.file.Close()
		logmgr.file = nil
	}
}

// Rotate Log rotate.
func Rotate() error {
	logmgr.lock.Lock()
	defer logmgr.lock.Unlock()

	if logmgr.logType == LogTypeFile && logmgr.file != nil {
		_ = logmgr.file.Close()
		return setFilelog(logmgr.logFile, logmgr.logLevel)
	}

	return nil
}
