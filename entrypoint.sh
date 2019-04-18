#!/bin/bash
dockerize -wait tcp://db:3306 -timeout 90s

echo "Running migrations"
/goose -dir /migrations mysql $DATABASE up
/comedian $@
