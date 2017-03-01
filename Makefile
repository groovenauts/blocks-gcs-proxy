PKGDIR=./pkg
BASENAME=magellan-gcs-proxy
VERSION=`grep VERSION version.go | cut -f2 -d\"`
OS=linux
ARCH=amd64

all: build

setup:
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr

build:
	mkdir -p ${PKGDIR}
	gox -output="${PKGDIR}/{{.Dir}}_${OS}_${ARCH}" -os="${OS}" -arch="${ARCH}"

release: build
	ghr -u groovenauts --replace --draft ${VERSION} pkg

prerelease: build
	ghr -u groovenauts --replace --draft --prerelease ${VERSION} pkg

version:
	echo ${VERSION}

clean:
	rm -rf ${PKGDIR}
