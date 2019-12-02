.SILENT:
.PHONY: sh install

##########
# Docker #
##########

define docker_run
	ID=$$( \
		docker build \
			--quiet \
			. \
	) \
	&& docker run \
		--rm \
		--tty \
		--interactive \
		--mount type=bind,consistency=delegated,source=$(CURDIR),target=/srv \
		--workdir /srv \
		$${ID} $(if $(1),$(strip $(1)),ash)
endef

ifeq ($(container),docker)
DOCKER_SHELL = $(SHELL)
else
DOCKER_SHELL = $(call docker_run,$(SHELL))
endif

######
# Sh #
######

ifneq ($(container),docker)
sh:
	$(call docker_run)
endif

###########
# Install #
###########

install: SHELL := $(DOCKER_SHELL)
install:
	go get

##########
# Update #
##########

update: SHELL := $(DOCKER_SHELL)
update:
	go get -u
	go mod tidy
