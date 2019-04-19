#!/bin/sh -x

exit_script() {
    echo "Shutting down..."
    trap - EXIT HUP INT QUIT PIPE TERM # clear the trap
    kill -- -$$ # Sends SIGTERM to child/sub processes
}

trap exit_script EXIT HUP INT QUIT PIPE TERM

echo "Sleeping.  Pid=$$"
sleep 2147483647 &

# Install dev helpers
apk add --no-cache \
  alpine-sdk

go mod download

wait $!
