# Copyright 2010, The "gowizard" Authors.  All rights reserved.
# Use of this source code is governed by the Simplified BSD License
# that can be found in the LICENSE file.

include $(GOROOT)/src/Make.$(GOARCH)

TARG=gowizard
GOFILES=\
	main.go\
	flag.go\
	metadata.go\
	template.go\
	util.go\

include $(GOROOT)/src/Make.cmd

data:
	[ -d $(GOROOT)/lib/$(TARG) ] || mkdir $(GOROOT)/lib/$(TARG)
	cp -R ../data/* $(GOROOT)/lib/$(TARG)/

install: install-default data clean

install-default:
	mv $(TARG) $(GOBIN)/$(TARG)

