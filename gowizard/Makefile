include $(GOROOT)/src/Make.$(GOARCH)

TARG=gowizard
GOFILES=\
	main.go\
	metadata.go\
	template.go\

include $(GOROOT)/src/Make.cmd

data:
	[ -d $(GOROOT)/lib/$(TARG) ] || mkdir $(GOROOT)/lib/$(TARG)
	cp -R ../data/* $(GOROOT)/lib/$(TARG)/

