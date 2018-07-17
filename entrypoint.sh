#!/bin/sh

set -e

while true
do
    /goose -dir /migrations mysql $COMEDIAN_DATABASE up && break
    sleep 2
done
/comedian $@
