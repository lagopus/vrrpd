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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/lagopus/vrrpd/agent"
	"github.com/lagopus/vrrpd/config"
	"github.com/lagopus/vrrpd/logger"
	"github.com/lagopus/vrrpd/module"
	"github.com/lagopus/vrrpd/rpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

var (
	version   string
	revision  string
	builddate string
	goversion string
)

var opts struct {
	DebugMode  bool   `short:"d" long:"debug" description:"Debug mode"`
	LogFile    string `short:"f" long:"logfile" default:"syslog" description:"Log file path (syslog, stderr, LOG_FILE_NAME)"`
	LogLevel   string `short:"l" long:"loglevel" default:"info" description:"Log Level (debug, info, warning, error, fatal, panic)"`
	ConfigFile string `short:"c" long:"conf" default:"/usr/local/etc/vsw_vrrpd.yml" description:"Path to config file"`
	PidFile    string `short:"p" long:"pid" default:"/var/run/vrrpd.pid" description:"Path to config file"`
	Version    bool   `short:"v" long:"version" description:"Show version"`
}

func versionString() string {
	return fmt.Sprintf("vrrp-agent: %s-%s(build at %s, %s)",
		version, revision, builddate, goversion)
}

func createPidFile(file string, pid int) error {
	var err error

	s := fmt.Sprintf("%v\n", pid)
	if err = ioutil.WriteFile(file, []byte(s), 644); err != nil {
		return err
	}

	return nil
}

func removePidFile(file string) {
	if _, err := os.Stat(file); err == nil {
		if err = os.Remove(file); err != nil {
			log.Errorf("Can't remove pid file: %v", err)
		}
	}
}

func registModules(agentConfig *config.AgentConfig, wg *sync.WaitGroup) {
	updateHandler := agent.NewUpdateHandler(wg)
	updateFunc := func(conf *config.AgentConfig) {
		updateHandler.SendHandlerChannel(conf)
	}

	datastore := rpc.NewDatastore(agentConfig.DsAddr.String(),
		int(agentConfig.DsPort), updateFunc, wg)

	recvHandler := agent.NewRecvHandler(wg)
	recvFunc := func(packets *rpc.BulkPackets) {
		recvHandler.SendHandlerChannel(packets)
	}

	hostif := rpc.NewHostif(agentConfig.HostifAddr.String(),
		int(agentConfig.HostifPort), recvFunc, wg)

	dpagent := rpc.NewDPAgent(agentConfig.DpaAddr.String(),
		int(agentConfig.DpaPort), wg)

	advTimer := agent.NewAdvTimer(hostif, wg)

	mDownTimer := agent.NewMDownTimer(wg)

	signaleHandler := agent.NewSignalHandler(wg)

	module.RegisterModule(signaleHandler)
	module.RegisterModule(datastore)
	module.RegisterModule(hostif)
	module.RegisterModule(dpagent)
	module.RegisterModule(advTimer)
	module.RegisterModule(mDownTimer)
	module.RegisterModule(recvHandler)
	module.RegisterModule(updateHandler)
}

func daemonize() error {
	// not fork.
	var pid int
	var err error
	if pid, err = syscall.Setsid(); err != nil {
		return err
	}
	if err = createPidFile(opts.PidFile, pid); err != nil {
		return err
	}

	_ = os.Stdin.Close()
	_ = os.Stdout.Close()
	_ = os.Stderr.Close()

	return nil
}

func final() {
	if !opts.DebugMode {
		removePidFile(opts.PidFile)
	}
	logger.Final()
}

// main

func main() {
	defer final()

	var err error
	if _, err = flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println(versionString())
		return
	}

	logLevel := opts.LogLevel
	if opts.DebugMode {
		logLevel = "debug"
	} else {
		if err = daemonize(); err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	// logger.
	if err = logger.Set(opts.LogFile, logLevel); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	grpclog.SetLogger(log.StandardLogger())

	log.Infof(versionString())

	// read config
	cmgr := config.GetMgr()
	if err = cmgr.ReadConfig(opts.ConfigFile); err != nil {
		log.Errorf("config read error: %v", err)
		os.Exit(1)
	}
	agentConfig := cmgr.GetCurrentConfig()
	log.Infof("agent config(%s): %s", opts.ConfigFile, agentConfig.String())

	var wg sync.WaitGroup
	registModules(agentConfig, &wg)

	if err = module.StartModules(); err != nil {
		log.Errorf("module start error: %v", err)
		os.Exit(1)
	}
	log.Infof("agent modules start")

	wg.Wait()
	log.Infof("main exit.")
}
