#!/bin/sh
set -ev

## Install dependencies
goinstall [url]

## Build
cd cmd; make install

## Clean
make clean

## Install succeeded!

