BIN := varnam-ibus-engine
INSTALL_PREFIX := /usr/local
VERSION := $(shell echo $$(git describe --abbrev=0 --tags || echo "latest") | sed s/v//)
RELEASE_NAME := varnam-ibus-engine-${VERSION}-${shell uname -m}
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
	$(shell mkdir component)
	./${BIN} -s ml-inscript -lang ml -xml component/varnam-ml-inscript.xml -prefix ${INSTALL_PREFIX}

	$(shell SCHEMES=("ml" "ta" "hi" "te" "ka" "bn" "ne"); for s in $${SCHEMES[@]}; do echo $s; ./${BIN} -s $$s -lang $$s -xml component/varnam-$$s.xml -prefix ${INSTALL_PREFIX}; done)

ubuntu-14:
	CGO_CFLAGS="-w" go build -tags "pango_1_36,gtk_3_10,glib_2_40,cairo_1_13,gdk_pixbuf_2_30" -ldflags "-s -w" -o ${BIN} .
	$(MAKE) ibus-xml
	$(MAKE) install-script

# Won't work
ubuntu-18:
	go build -tags "pango_1_42,gtk_3_22,glib_2_66,cairo_1_15" -ldflags "-s -w" -o ${BIN} .
	$(MAKE) ibus-xml
	$(MAKE) install-script

# Won't work
ubuntu-20:
	go build -ldflags "-s -w" -o ${BIN} .
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
