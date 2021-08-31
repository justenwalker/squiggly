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

Download [scripts/squiggly.sh](./scripts/squiggly.sh), update the settings, and source it from your `.bashrc/.zshrc`

After you open a new terminal, you can run `squiggly_up` to start using the proxy, and `squiggly_down` to switch it off.

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
      --pac string        url to the proxy auto config (PAC) file
  -p, --proxy string      the upstream HTTP Proxy
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