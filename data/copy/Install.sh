#!/usr/bin/env bash
set -ev

## Install dependencies
goinstall [url]

## Build the command
cd cmd; make install

## Install succeeded!

