PKGDIR=./pkg
BASENAME=blocks-gcs-proxy
VERSION=`grep VERSION version.go | cut -f2 -d\"`
OS=linux
ARCH=amd64
UNFORMATTED=$(shell gofmt -l *.go)

all: build

setup:
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr

checksetup:
	go get golang.org/x/tools/cmd/goimports

check: checkfmt
	go vet *.go
	goimports -l *.go

checkfmt:
ifneq ($(UNFORMATTED),)
	@echo $(UNFORMATTED)
	exit 1
else
	@echo "gofmt -l *.go OK"
endif

build:
	mkdir -p ${PKGDIR}
	gox -output="${PKGDIR}/{{.Dir}}_${OS}_${ARCH}" -os="${OS}" -arch="${ARCH}"

release: build
	ghr -u groovenauts -r blocks-gcs-proxy --replace --draft ${VERSION} pkg

prerelease: build
	ghr -u groovenauts -r blocks-gcs-proxy --replace --draft --prerelease ${VERSION} pkg

version:
	echo ${VERSION}

clean:
	rm -rf ${PKGDIR}
