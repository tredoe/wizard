# Copyright 2010  The "gowizard" Authors
#
# Use of this source code is governed by the Simplified BSD License
# that can be found in the LICENSE file.
#
# This software is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
# OR CONDITIONS OF ANY KIND, either express or implied. See the License
# for more details.

include $(GOROOT)/src/Make.inc

TARG=gowizard
GOFILES=\
	main.go\
	error.go\
	flag.go\
	metadata.go\
	replace.go\
	template.go\
	tmpl-data.go\
	util.go\

include $(GOROOT)/src/Make.cmd

# Copy both templates and licenses to the system.
data:
	[ -d $(GOROOT)/lib/$(TARG) ] || mkdir $(GOROOT)/lib/$(TARG)
	cp -R ../data/* $(GOROOT)/lib/$(TARG)/

# Add a configuration file to the current user.
config:
	cp $(GOROOT)/lib/$(TARG)/config/user $(HOME)/.gowizard

# === Installation
install: install-default

install-default:
	mv $(TARG) $(GOBIN)/$(TARG)

