#!/bin/bash

base_cmd="/proxy/coap-proxy --maps-dir /proxy/maps --debug-log --http-target http://synapse:8008 --disable-encryption"

if [ -n "$PROXY_COAP_TARGET" ]; then
    $base_cmd --coap-target $PROXY_COAP_TARGET
else
    $base_cmd
fi
