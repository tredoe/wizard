#!/bin/sh
set -ev

## Install dependencies
goinstall [url]

## Build
cd {{.section dir_is_cmd}}cmd{{.or}}{{package_name}}{{.end}}; make install

## Clean
make clean

## Install succeeded!

