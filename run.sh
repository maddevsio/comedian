#!/bin/bash

set -e

go build
docker-compose rm
docker-compose build
docker-compose up
