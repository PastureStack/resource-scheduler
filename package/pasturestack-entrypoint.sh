#!/bin/bash
set -euo pipefail

if [ "${PASTURESTACK_DEBUG:-${RANCHER_DEBUG:-false}}" != "false" ]; then
    set -x
fi

cert_bundle=$(/usr/bin/update-pasturestack-ca)
if [ -n "${cert_bundle}" ]; then
    export SSL_CERT_FILE="${cert_bundle}"
fi

if [ "$#" -eq 0 ]; then
    set -- resource-scheduler
elif [ "${1#-}" != "$1" ]; then
    set -- resource-scheduler "$@"
fi

exec "$@"
