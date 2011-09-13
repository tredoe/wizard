#!/bin/sh
set -ev

cd gowizard

## Install dependencies
#goinstall github.com/kless/goconfig/config
#goinstall github.com/kless/Go-Inline/inline

## Build
make install

## Clean
make clean

## Install templates and licenses
make data

## Install configuration file
make config

## Install succeeded!

