#!/bin/bash

NETWORK="tp0_testing_net"
PORT=12345
MESSAGE="server testing"

RESPONSE=$(docker run --rm --network=$NETWORK busybox sh -c "echo '$MESSAGE' | nc server $PORT")

if [[ "$RESPONSE" == "$MESSAGE" ]]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
