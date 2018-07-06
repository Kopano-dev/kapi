PACKAGE_NAME = kapi-playground

# Variables

VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2>/dev/null | sed 's/^v//' || \
			cat $(CURDIR)/../.version 2> /dev/null || echo 0.0.0-unreleased)

.PHONY: all
all:

pubjs/umd/pubs.js:

.PHONY: dist
dist: pubsjs/umd/pubs.js ; $(info building dist tarball ...)
	@mkdir -p "dist/${PACKAGE_NAME}-${VERSION}"
	@mkdir -p "dist/${PACKAGE_NAME}-${VERSION}/pubsjs/umd"
	cp -avf pubsjs/umd/pubs.js "dist/${PACKAGE_NAME}-${VERSION}/pubsjs/umd/"
	@cd dist && \
	cp -avf ../*.html "${PACKAGE_NAME}-${VERSION}" && \
	cp -avf ../*.css "${PACKAGE_NAME}-${VERSION}" && \
	cp -avf ../*.js "${PACKAGE_NAME}-${VERSION}" && \
	tar --owner=0 --group=0 -czvf ${PACKAGE_NAME}-${VERSION}.tar.gz "${PACKAGE_NAME}-${VERSION}" && \
	cd ..
