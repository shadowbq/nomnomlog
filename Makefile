export GO15VENDOREXPERIMENT=1

include packaging/Makefile.packaging

.PHONY: reportVersion depend dependPackage clean test build tarball
.DEFAULT: build

GOLD_FLAGS="-X main.Version=$(PACKAGE_VERSION)"

X86_PLATFORMS := windows linux
X64_PLATFORMS := windows linux
ARM_PLATFORMS := linux
CGO_PLATFORMS := darwin

ifeq '$(findstring ;,$(PATH))' ';'
	DETECTED_OS := Windows
else
	DETECTED_OS := $(shell uname 2>/dev/null || echo Unknown)
	DETECTED_OS := $(patsubst CYGWIN%,Cygwin,$(DETECTED_OS))
	DETECTED_OS := $(patsubst MSYS%,MSYS,$(DETECTED_OS))
	DETECTED_OS := $(patsubst MINGW%,MSYS,$(DETECTED_OS))
endif

BUILD_PAIRS := $(foreach p,$(X86_PLATFORMS), $(p)/i386 )
BUILD_PAIRS += $(foreach p,$(X64_PLATFORMS), $(p)/amd64 )
BUILD_PAIRS += $(foreach p,$(ARM_PLATFORMS), $(p)/armhf )
BUILD_PAIRS += $(foreach p,$(ARM_PLATFORMS), $(p)/armh64 )

BUILD_DOCS := README.md LICENSE example_config.yml

ifeq ($(DETECTED_OS),Darwin)
	BUILD_PAIRS += $(foreach p,$(CGO_PLATFORMS), $(p)/amd64 )
endif

package: $(BUILD_PAIRS)

reportVersion: 
	@echo "\033[32mProduct Version $(PACKAGE_VERSION)"

build: reportVersion depend clean test
	@echo
	@echo "\033[32mBuilding ----> \033[m"

	gox -ldflags=$(GOLD_FLAGS) -os="$(X86_PLATFORMS)" -arch="386" -output "build/{{.OS}}/i386/nomnomlog/nomnomlog"
	gox -ldflags=$(GOLD_FLAGS) -os="$(X64_PLATFORMS)" -arch="amd64" -output "build/{{.OS}}/amd64/nomnomlog/nomnomlog"
	
	gox -ldflags=$(GOLDFLAGS) -os="$(ARM_PLATFORMS)" -arch="arm" -output "build/{{.OS}}/armhf/nomnomlog/nomnomlog"
	gox -ldflags=$(GOLDFLAGS) -os="$(ARM_PLATFORMS)" -arch="arm64" -output "build/{{.OS}}/armh64/nomnomlog/nomnomlog"
# Mac OS X - daemon_darwin.go:6:10: fatal error: mach-o/dyld.h 
ifeq ($(DETECTED_OS),Darwin)
	gox -ldflags=$(GOLD_FLAGS) -cgo -os="$(CGO_PLATFORMS)" -arch="amd64" -output "build/{{.OS}}/amd64/nomnomlog/nomnomlog"
endif
	

clean:
	@echo
	@echo "\033[32mCleaning Build ----> \033[m"
	$(RM) -rf pkg/*
	$(RM) -rf build/*
	$(RM) -rf tmp/*


test:
	@echo
	@echo "\033[32mTesting ----> \033[m"
	go list ./... | grep -v /vendor/ | xargs -I {} go test -v -race {}

dependPackage:
	@echo
	@echo "\033[32mChecking Package Dependencies ----> \033[m"
	@type zip >/dev/null 2>&1|| { \
	  echo "\033[1;33mZIP is required to package this application\033[m"; \
	  echo "\033[1;33mIf you are using homebrew on OSX, run\033[m"; \
	  echo "Recommend: $$ brew install go --cross-compile-all"; \
	  exit 1; \
	}

	@gem list | grep fpm >/dev/null 2>&1 || { \
	  echo "\033[1;33mfpm is not installed. See https://github.com/jordansissel/fpm\033[m"; \
	  echo "Recommend: $$ gem install fpm"; \
	  exit 1; \
	}

	@gem list | grep ronn >/dev/null 2>&1 || { \
	  echo "\033[1;33mronn is not installed. See https://github.com/rtomayko/ronn\033[m"; \
	  echo "Recommend: $$ gem install ronn"; \
	  exit 1; \
	}

	@type rpmbuild >/dev/null 2>&1 || { \
	  echo "\033[1;33mRecommend: rpmbuild is not installed. See the package for your distribution\033[m"; \
	  exit 1; \
	}

depend:
	@echo
	@echo "\033[32mChecking Build Dependencies ----> \033[m"

ifndef PACKAGE_VERSION
	@echo "\033[1;33mPACKAGE_VERSION is not set. In order to build a package I need PACKAGE_VERSION=n\033[m"
	exit 1;
endif

ifndef GOPATH
	@echo "\033[1;33mGOPATH is not set. This means that you do not have go setup properly on this machine\033[m"
	@echo "$$ mkdir ~/gocode";
	@echo "$$ echo 'export GOPATH=~/gocode' >> ~/.bash_profile";
	@echo "$$ echo 'export PATH=\"\$$GOPATH/bin:\$$PATH\"' >> ~/.bash_profile";
	@echo "$$ source ~/.bash_profile";
	exit 1;
endif

	@type go >/dev/null 2>&1|| { \
	  echo "\033[1;33mGo is required to build this application\033[m"; \
	  echo "\033[1;33mIf you are using homebrew on OSX, run\033[m"; \
	  echo "Recommend: $$ brew install go --cross-compile-all"; \
	  exit 1; \
	}

	@type govendor >/dev/null 2>&1|| { \
	  echo "\033[1;33mgovendor is not installed. See https://github.com/kardianos/govendor\033[m"; \
	  echo "Recommend: $$ go get -u github.com/kardianos/govendor"; \
	  exit 1; \
	}


	@type semver-bump >/dev/null 2>&1 || { \
	  echo "\033[1;33msemver-bump is not installed. See https://github.com/giantswarm/semver-bump\033[m"; \
	  echo "Recommend: $$ go get github.com/giantswarm/semver-bump"; \
	  exit 1; \
	}

	@type gox >/dev/null 2>&1 || { \
	  echo "\033[1;33mGox is not installed. See https://github.com/mitchellh/gox\033[m"; \
	  echo "Recommend: $$ go get github.com/mitchellh/gox"; \
	  exit 1; \
	}

$(BUILD_PAIRS): dependPackage build
	@echo
	@echo "\033[32mPackaging ----> $@\033[m"
	$(eval PLATFORM := $(strip $(subst /, ,$(dir $@))))
	$(eval ARCH := $(notdir $@))
	mkdir -p pkg || echo
	mkdir -p build/$@/nomnomlog
	cp $(BUILD_DOCS) build/$@/nomnomlog
	@pwd
	@mkdir -p pkg/tmp/usr/share/man/man5
	ronn -W -r README.md --pipe > pkg/tmp/usr/share/man/man5/nomnomlog.5 2>/dev/null
	@file pkg/tmp/usr/share/man/man5/nomnomlog.5

	if [ "$(PLATFORM)" = "linux" ]; then\
		mkdir -p pkg/tmp/etc/init.d;\
		mkdir -p pkg/tmp/usr/local/bin;\
		cp -f example_config.yml pkg/tmp/etc/nomnomlog-config.yml;\
		cp -f packaging/linux/nomnomlog.initd pkg/tmp/etc/init.d/nomnomlog;\
		cp -f build/$@/nomnomlog/nomnomlog pkg/tmp/usr/local/bin;\
		(cd pkg && \
		fpm \
		  -s dir \
		  -C tmp \
		  -t deb \
		  -n $(PACKAGE_NAME) \
		  -v $(PACKAGE_VERSION) \
		  --vendor $(PACKAGE_VENDOR) \
		  --license $(PACKAGE_LICENSE) \
		  -a $(ARCH) \
		  -m $(PACKAGE_CONTACT) \
		  --description $(PACKAGE_DESCRIPTION) \
		  --url $(PACKAGE_URL) \
		  --before-remove ../packaging/linux/deb/prerm \
		  --after-install ../packaging/linux/deb/postinst \
		  --config-files etc/nomnomlog-config.yml \
		  --config-files etc/init.d/nomnomlog usr/local/bin/nomnomlog etc/nomnomlog-config.yml etc/init.d/nomnomlog \
		  usr/share/man && \
		fpm \
		  -s dir \
		  -C tmp \
		  -t rpm \
		  -n $(PACKAGE_NAME) \
		  -v $(PACKAGE_VERSION) \
		  --vendor $(PACKAGE_VENDOR) \
		  --license $(PACKAGE_LICENSE) \
		  -a $(ARCH) \
		  -m $(PACKAGE_CONTACT) \
		  --description $(PACKAGE_DESCRIPTION) \
		  --url $(PACKAGE_URL) \
		  --before-remove ../packaging/linux/rpm/preun \
		  --after-install ../packaging/linux/rpm/post \
		  --config-files etc/nomnomlog-config.yml \
		  --config-files etc/init.d/nomnomlog \
		  --rpm-os linux usr/local/bin/nomnomlog etc/nomnomlog-config.yml etc/init.d/nomnomlog \
		  usr/share/man );\
		rm -R -f pkg/tmp;\
	fi

	if [ "$(PLATFORM)" = "windows" ]; then \
		cd build/$@ && echo `pwd` && zip -r ../../../pkg/nomnomlog_$(PLATFORM)_$(ARCH).zip nomnomlog;\
	else \
		cd build/$@ && echo `pwd` && tar -cvzf ../../../pkg/nomnomlog_$(PLATFORM)_$(ARCH).tar.gz nomnomlog;\
	fi

	@echo "Done Packaging"

	
