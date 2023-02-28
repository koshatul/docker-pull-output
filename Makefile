GO_MATRIX += darwin/amd64 darwin/arm64
GO_MATRIX += linux/amd64 linux/arm64
GO_MATRIX += windows/amd64

APP_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH ?= $(shell git show -s --format=%h)

GO_DEBUG_ARGS   ?= -v -ldflags "-X main.version=$(GO_APP_VERSION)+debug -X main.gitHash=$(GIT_HASH) -X main.buildDate=$(APP_DATE)"
GO_RELEASE_ARGS ?= -v -ldflags "-X main.version=$(GO_APP_VERSION) -X main.gitHash=$(GIT_HASH) -X main.buildDate=$(APP_DATE) -s -w"

_GO_GTE_1_14 := $(shell expr `go version | cut -d' ' -f 3 | tr -d 'a-z' | cut -d'.' -f2` \>= 14)
ifeq "$(_GO_GTE_1_14)" "1"
_MODFILEARG := -modfile tools.mod
endif

-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/ext/na4ma4/lib/golangci-lint/v1/Makefile
-include .makefiles/ext/na4ma4/lib/goreleaser/v1/Makefile

.makefiles/ext/na4ma4/%: .makefiles/Makefile
	@curl -sfL https://raw.githubusercontent.com/na4ma4/makefiles-ext/main/v1/install | bash /dev/stdin "$@"

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"

.PHONY: install
install: artifacts/build/release/$(GOHOSTOS)/$(GOHOSTARCH)/docker-pull-output
	install "$(<)" /usr/local/bin/

.PHONY: run
run: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/docker-pull-output
	$< $(RUN_ARGS)

.PHONY: upx
upx: $(patsubst artifacts/build/%,artifacts/upx/%.upx,$(_GO_RELEASE_TARGETS_ALL))

artifacts/upx/%.upx: artifacts/build/%
	-@mkdir -p "$(@D)"
	-$(RM) -f "$(@)"
	upx -o "$@" "$<"

test:: artifacts/test/testoutput.log

.DELETE_ON_ERROR: artifacts/test/testoutput.log
artifacts/test/testoutput.log: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/docker-pull-output test/testoutput.txt
	-@mkdir -p "$(@D)"
	cat "test/testoutput.txt" | "$(<)" 2>&1 | sed -e 's/.* level=//' | tee "$(@)" > /dev/null
	diff "$(@)" "test/testoutput.txt.run" > /dev/null

######################
# Linting
######################

ci:: lint
