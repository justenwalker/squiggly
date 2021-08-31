# Source this file in your .bashrc/.zshrc

## ---- Settings (You can change these) ----
## Which port should squiggly listen on
SQUIGGLY_PROXY_PORT=8800

## Does your proxy require auth?
SQUIGGLY_PROXY_USERNAME=${SQUIGGLY_PROXY_USERNAME:-"${USER}"}
## Use squiggly_auth to set the password in your system keychain

## Pick one:
# 1. Proxy auto config
#SQUIGGLY_PAC_URL=http://example.com/proxy.pac
# 2. Use a single upstream proxy
SQUIGGLY_FORWARD_PROXY="http://proxy.example.com:8080"

## Where should squiggly store its support files
SQUIGGLY_PIDFILE=$HOME/.squiggly
SQUIGGLY_LOGFILE=$HOME/squiggly.log

## Should the logs be verbose?
SQUIGGLY_VERBOSE=N

# Does your proxy use SPNEGO/GSSAPI/NTLM SSP/Kerberos to authenticate? (Active Directory)
# You'll probably need to set these
#SQUIGGLY_KRB5_REALM=realm.example.com
#SQUIGGLY_KRB5_CONF=/etc/krb5.conf

# You don't have a krb5.conf? Run squiggly_krb5conf to generate one

## ---vvv--- Don't change anything below ---vvv--- ##
SQUIGGLY_PROXY_ADDR="localhost:${SQUIGGLY_PROXY_PORT}"
run_squiggly() {
  echo "Running Squiggly"
  _SQUIGGLY_ARGS=(--address "${SQUIGGLY_PROXY_ADDR}")
  if [ -n "${KRB5_REALM}" ]; then
    echo "- Realm: ${SQUIGGLY_KRB5_REALM}"
    _SQUIGGLY_ARGS+=(--realm "${SQUIGGLY_KRB5_REALM}")
    echo "- krb5.conf: ${SQUIGGLY_KRB5_CONF}"
    _SQUIGGLY_ARGS+=(--krb5conf "${SQUIGGLY_KRB5_CONF}")
  fi
  if [ -n "${SQUIGGLY_PROXY_USERNAME}" ]; then
    echo "- User: ${SQUIGGLY_PROXY_USERNAME}"
    _SQUIGGLY_ARGS+=(--user "${SQUIGGLY_PROXY_USERNAME}")
  fi
  if [ -n "${SQUIGGLY_FORWARD_PROXY}" ]; then
    echo "- Proxy: ${SQUIGGLY_FORWARD_PROXY}"
    _SQUIGGLY_ARGS+=(--proxy "${SQUIGGLY_FORWARD_PROXY}")
  elif [ -n "${SQUIGGLY_PAC_URL}" ]; then
    echo "- PAC: ${SQUIGGLY_PAC_URL}"
    _SQUIGGLY_ARGS+=(--pac "${SQUIGGLY_PAC_URL}")
  else
    echo "No upstream proxy"
  fi
  if [ "${SQUIGGLY_VERBOSE}" = "Y" ]; then
    _SQUIGGLY_ARGS+=(--verbose)
  fi
  squiggly proxy "${_SQUIGGLY_ARGS[@]}" > "${SQUIGGLY_LOGFILE}" 2>&1 &
  echo "$!" > "${SQUIGGLY_PIDFILE}"
  disown '%%'
  echo "Squiggly Up ($PID): http://${SQUIGGLY_PROXY_ADDR}"
}

squiggly_auth() {
  echo "Setting proxy credentials for ${SQUIGGLY_PROXY_USERNAME}"
  squiggly auth -u "${SQUIGGLY_PROXY_USERNAME}"
}

squiggly_krb5conf() {
  if [ -z "${SQUIGGLY_KRB5_REALM}" ]; then
    echo "Must set SQUIGGLY_KRB5_REALM" >&2
    exit 1
  fi
  if [ -z "${SQUIGGLY_KRB5_CONF}" ]; then
    echo "Must set SQUIGGLY_KRB5_CONF" >&2
    exit 1
  fi
  echo "Generating ${SQUIGGLY_KRB5_CONF} for ${SQUIGGLY_KRB5_REALM}"
  squiggly krb5conf -r ${SQUIGGLY_KRB5_REALM} | tee ${SQUIGGLY_KRB5_CONF}
}

squiggly_up() {
  if [ -r "${SQUIGGLY_PIDFILE}" ]; then
    PID=$(cat "${SQUIGGLY_PIDFILE}")
    if ! ps -p "${PID}" > /dev/null 2>&1; then
      run_squiggly
    fi
  else
    run_squiggly
  fi

  # Set proxy environment variables
  export HTTP_PROXY="http://${SQUIGGLY_PROXY_ADDR}"
  export HTTPS_PROXY=${HTTP_PROXY}
  export NO_PROXY="localhost,127.0.0.1,${HOST}"

  export http_proxy=${HTTP_PROXY}
  export https_proxy=${HTTP_PROXY}
  export no_proxy=${NO_PROXY}
}

squiggly_down() {
  if [ -r "${SQUIGGLY_PIDFILE}" ]; then
    PID=$(cat "${SQUIGGLY_PIDFILE}")
    if ps -p "${PID}" > /dev/null 2>&1; then
      echo "Stopping Squiggly ($PID)"
      kill "${PID}"
    fi
  fi
  unset HTTP_PROXY
  unset HTTPS_PROXY
  unset NO_PROXY
  unset http_proxy
  unset https_proxy
  unset no_proxy
}