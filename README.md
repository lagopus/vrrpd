# Lagopus VRRPd

## Recommended
* Go 1.12.6

## Getting started
### Environment variable
Set GOPATH and PATH.

e.g.)
```
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### Download Lagopus VRRPd

```bash
% go get -d github.com/lagopus/vrrpd

 or

% mkdir -p $GOPATH/src/github.com/lagopus
% cd $GOPATH/src/github.com/lagopus
% git clone https://github.com/lagopus/vrrpd
```

### Build & Install

```bash
% cd $GOPATH/src/github.com/lagopus/vrrpd
% make vendor
% make install
```

### Build only

```bash
% cd $GOPATH/src/github.com/lagopus/vrrpd
% make
```

## How to Run Lagopus VRRP
### Usage

```
vrrpd [OPTIONS]

Application Options:
  -d, --debug     Debug mode
  -f, --logfile=  Log file path (syslog, stderr, LOG_FILE_NAME) (default: syslog)
  -l, --loglevel= Log Level (debug, info, warning, error, fatal, panic) (default: info)
  -c, --conf=     Path to config file (default: /usr/local/etc/vsw_vrrpd.yml)
  -p, --pid=      Path to config file (default: /var/run/vrrpd.pid)

Help Options:
  -h, --help      Show this help message
```

### Run
e.g.)

```bash
# Foreground (Debug mode)
% sudo vrrpd -d -f stderr

# Background
% sudo -b vrrpd -l debug -f <LOG FILE>
```

### Run unit tests

```bash
$ make test
```
