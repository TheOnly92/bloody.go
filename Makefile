include $(GOROOT)/src/Make.inc

TARG=bloody
GOFILES=\
	bloody.go\
	config.go\
	controllers/admin.go\
	controllers/index.go\
	models/pages.go\
	models/posts.go\
	models/preferences.go\

include $(GOROOT)/src/Make.cmd