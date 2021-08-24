BIN := varnam-ibus-engine
INSTALL_PREFIX := /usr/local
VERSION := $(shell git describe --abbrev=0 --tags | sed s/v//)
RELEASE_NAME := varnam-ibus-engine-${VERSION}
IBUS_COMPONENT_INSTALL_LOC := "/usr/share/ibus/component"

install-script:
	cp install.sh.in install.sh
	sed -i "s#@INSTALL_PREFIX@#${INSTALL_PREFIX}#g" install.sh
	sed -i "s#@IBUS_COMPONENT_INSTALL_LOC@#${IBUS_COMPONENT_INSTALL_LOC}#g" install.sh

	chmod +x install.sh

install:
	./install.sh

ibus-xml: SHELL := /bin/bash
ibus-xml:
	mkdir -p component
	./${BIN} -s ml-inscript -lang ml -xml component/varnam-ml-inscript.xml -prefix ${INSTALL_PREFIX}

	$(shell SCHEMES=("ml" "ta" "hi" "te" "ka" "bn"); for s in $${SCHEMES[@]}; do echo $s; ./${BIN} -s $$s -lang $$s -xml component/varnam-$$s.xml -prefix ${INSTALL_PREFIX}; done)

ubuntu-18:
	go build -tags pango_1_42,gtk_3_22 -o ${BIN} .
	$(MAKE) ibus-xml
	$(MAKE) install-script

ubuntu-20:
	go build -o ${BIN} .
	$(MAKE) ibus-xml
	$(MAKE) install-script

release:
	mkdir -p ${RELEASE_NAME} ${RELEASE_NAME}/icons ${RELEASE_NAME}/component
	cp ${BIN} ${RELEASE_NAME}/
	cp install.sh ${RELEASE_NAME}/
	cp icons/*.png ${RELEASE_NAME}/icons/
	cp component/*.xml ${RELEASE_NAME}/component

	zip -r ${RELEASE_NAME}.zip ${RELEASE_NAME}/*

clean:
	rm "component/*.xml"
	rmdir component
