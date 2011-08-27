#!/bin/sh
set -ev

## Install dependencies
#goinstall github.com/kless/goconfig/config
#goinstall github.com/kless/inline

## Build
make install

## Clean
make clean

## Install templates and licenses
make data

## Install configuration file
make config

## Install succeeded!

