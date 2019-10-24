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
SQUIGGLY_KRB5_REALM=realm.example.com
SQUIGGLY_KRB5_CONF=/etc/krb5.conf

run_proxy() {
  echo "Running Squiggly"
  squiggly proxy \
      --address "localhost:$SQUIGGLY_PROXY_PORT" \
      --pac "http://wmtpac.wal-mart.com/proxies/anycast-universal.pac" \
      --user "${SQUIGGLY_PROXY_USERNAME}" \
      --realm "${SQUIGGLY_KRB5_REALM}" \
      --krb5conf "${SQUIGGLY_KRB5_CONF}" \
      --verbose > ${SQUIGGLY_LOGFILE} 2>&1 &
  echo "$!" > $SQUIGGLY_PIDFILE
  disown '%%'
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
export HTTPS_PROXY=${HTTP_PROXY}
export NO_PROXY="localhost,127.0.0.1,${HOST}"

export http_proxy=${HTTP_PROXY}
export https_proxy=${HTTP_PROXY}
export no_proxy=${NO_PROXY}
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
  -a, --address string    listen address for the proxy server (default "localhost:8800")
  -h, --help              help for proxy
  -k, --krb5conf string   kerberos config
  -p, --pac string        url to the proxy auto config (PAC) file
  -r, --realm string      realm for kerberos/negotiate authentication
  -s, --service string    service name, used to distinguish between auth configurations (default "squiggly")
  -u, --user string       user name, used to log into proxy servers. Omit to use an unauthenticated proxy.
  -v, --verbose           enable verbose logging
```

### Example

Runs the proxy on the default address (`localhost:8800`) using the default service name (`squiggly`) to retreive proxy authentication from the keyring.

```bash
$ squiggly proxy --pac http://example.com/proxy.pac --verbose --user myusername
```

## Kerberos Config

There is a utility method for writing a default `krb5.conf` that uses dns to discover the servers, to make it easier to configure the Kerberos auth.
It will probably need some modifications to work correctly, but it might be a good starting point.

### Usage

```
Generate a krb5.conf

Usage:
  squiggly krb5conf [flags]

Flags:
  -h, --help           help for krb5conf
  -r, --realm string   kerberos realm
```

### Example

```bash
$ squiggly krb5conf --realm REALM.EXAMPLE.COM > krb5.conf
```