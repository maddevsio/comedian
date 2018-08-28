#!/bin/sh

/goose -dir /migrations mysql $COMEDIAN_DATABASE up

sleep 5

/comedian $@
