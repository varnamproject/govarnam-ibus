BIN := govarnam-ibus
INSTALL_PREFIX := /usr/local
VERSION := $(shell git describe --abbrev=0 --tags | sed s/v//)
RELEASE_NAME := govarnam-ibus-${VERSION}
IBUS_COMPONENT_INSTALL_LOC := "/usr/share/ibus/component"

build-install-script:
	cp install.sh.in install.sh
	sed -i "s#@INSTALL_PREFIX@#${INSTALL_PREFIX}#g" install.sh
	sed -i "s#@IBUS_COMPONENT_INSTALL_LOC@#${IBUS_COMPONENT_INSTALL_LOC}#g" install.sh

	chmod +x install.sh

install:
	./install.sh

ibus-xml:
	./${BIN} -xml govarnam.xml -prefix ${INSTALL_PREFIX}

build-ubuntu18:
	go build -tags pango_1_42,gtk_3_22 -o ${BIN} .
	$(MAKE) ibus-xml
	$(MAKE) build-install-script

build-ubuntu20:
	go build -o ${BIN} .
	$(MAKE) ibus-xml
	$(MAKE) build-install-script

release:
	mkdir -p ${RELEASE_NAME}
	cp ${BIN} ${RELEASE_NAME}/
	cp install.sh ${RELEASE_NAME}/
	cp *.png ${RELEASE_NAME}/
	cp *.xml ${RELEASE_NAME}/

	zip -r ${RELEASE_NAME}.zip ${RELEASE_NAME}/*
