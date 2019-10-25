# Squiggly

A Forwarding proxy with support for upstream Proxy Auto Config (PAC) written in Go.

`squiggly` will configure the upstream proxy based on the PAC file given. It will check every so often to determine if the upstream proxies are still reachable, and if not it will automatically disable routing to the upstream proxy and use a direct connection. Once the PAC file becomes reachable again, it will enable the upstream proxy.

Optionally, `squiggly` support proxy Basic authentication.

## Installation

### via go get

This will install to your `$GOPATH/bin` folder.

```bash
$ go install github.com/justenwalker/squiggly
```

## Quick Start

Set your password if you need to

```bash
$ squiggly auth --user "yourusername"
```

In your `.bashrc`, put something like this

```bash
SQUIGGLY_PROXY_PORT=8800
SQUIGGLY_PROXY_USERNAME=${PROXY_USERNAME:-"$USER"}
SQUIGGLY_PAC_URL=http://example.com/proxy.pac
SQUIGGLY_PIDFILE=$HOME/.squiggly
SQUIGGLY_LOGFILE=$HOME/squiggly.log

run_proxy() {
  echo "Running Squiggly"
  squiggly proxy --address "localhost:$SQUIGGLY_PROXY_PORT" --pac "${SQUIGGLY_PAC_URL}" --verbose --user "${SQUIGGLY_PROXY_USERNAME}" > ${SQUIGGLY_LOGFILE} 2>&1 &
  echo "$!" > $SQUIGGLY_PIDFILE
  disown
}

if [ -r $SQUIGGLY_PIDFILE ]; then
  PID=$(cat $SQUIGGLY_PIDFILE)
  if ! ps -p $PID > /dev/null 2>&1; then
    run_proxy
  fi
else
  run_proxy
fi

# Set proxy environment variables
export HTTP_PROXY=http://localhost:$SQUIGGLY_PROXY_PORT
export HTTPS_PROXY=http://localhost:$SQUIGGLY_PROXY_PORT
export NO_PROXY="localhost,127.0.0.1,${HOST},127.0.0.1"

export http_proxy=http://localhost:$SQUIGGLY_PROXY_PORT
export https_proxy=http://localhost:$SQUIGGLY_PROXY_PORT
export no_proxy=localhost,127.0.0.1,${HOST},127.0.0.1
```

After you open a new terminal, all your command-line application that support the standard proxy environment variables should start proxying through squiggly to your upstreams defined in the PAC.

## Authenticate

### Usage

```
Sets the proxy authentication credentials

Usage:
  squiggly auth [flags]

Flags:
  -h, --help             help for auth
  -s, --service string   service name, used to distinguish between auth configurations (default "squiggly")
  -u, --user string      user name, used to log into proxy servers (default "yourusername")
```

### Example

This will create a keychain entry for "squiggly" using the given "username", and will prompt for a password:

```bash
$ squiggly auth -u justen
[justen] Password: ********************
```

## Running the Proxy

### Usage

```
Starts the proxy server.

Usage:
  squiggly proxy [flags]

Flags:
  -a, --address string      listen address for the proxy server (default "localhost:8800")
  -h, --help                help for proxy
  -p, --pac string          url to the proxy auto config (PAC) file
  -s, --service string      service name, used to distinguish between auth configurations (default "squiggly")
  -u, --user string         user name, used to log into proxy servers (default "juw0006")
  -v, --verbose             enable verbose logging
```

### Example

Runs the proxy on the default address (`localhost:8800`) using the default service name (`squiggly`) to retreive proxy authentication from the keyring.

```bash
$ squiggly proxy --pac http://example.com/proxy.pac --verbose --user myusername
```

