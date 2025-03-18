#!/usr/bin/env sh
# SPDX-FileCopyrightText: 2022 Comcast Cable Communications Management, LLC
# SPDX-License-Identifier: Apache-2.0
set -e

# check arguments for an option that would cause /themis to stop
# return true if there is one
_want_help() {
    local arg
    for arg; do
        case "$arg" in
            -'?'|--help|-v)
                return 0
                ;;
        esac
    done
    return 1
}

_main() {
    # if command starts with an option, prepend themis
    if [ "${1:0:1}" = '-' ]; then
        set -- /themis "$@"
    fi

    # skip setup if they aren't running /themis or want an option that stops /themis
    if [ "$1" = '/themis' ] && ! _want_help "$@"; then
        echo "Entrypoint script for themis Server ${VERSION} started."

        if [ ! -s /etc/themis/themis.yaml ]; then
            echo "Building out template for file"
            /bin/spruce merge /tmp/themis_spruce.yaml > /etc/themis/themis.yaml
        fi
    fi

    exec "$@"
}

_main "$@"
