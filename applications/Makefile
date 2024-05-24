SUBDIRS := $(wildcard */.)

all :: $(SUBDIRS) deploy

$(SUBDIRS): deploy
	@$(MAKE) -C $@ $(MAKECMDGOALS)

deploy:
