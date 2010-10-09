#!/usr/bin/env bash
set -ev

## Install dependencies
goinstall github.com/kless/goconfig/config
goinstall github.com/kless/go-readin/readin

## Build the command
cd cmd; make install

## Install templates and licenses
make data

## Install configuration file
make config

## Install succeeded!

