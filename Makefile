ROOT = github.com/keydotcat/keycatd
ifeq ($(GIT_VERSION),)
	GIT_VERSION:=$(shell git describe --abbrev=8 --dirty --always --tags 2>/dev/null)
endif
SUF=
ifdef GOARCH
	SUF:=.$(GOARCH)
endif
ifdef GOOS
	SUF:=.$(GOOS)$(SUF)
ifeq ($(GOOS),windows)
	SUF:=$(SUF).exe
endif
endif
.PHONY: static web

all: dev

cmds:
	go get github.com/go-bindata/go-bindata/...
	go get github.com/acasajus/scaneo
	go get github.com/githubnemo/CompileDaemon

web:
	./scripts/build_web.sh

keycatd: bindir autogen
	go build -o bin/keycatd${SUF} ${ROOT}/cmd/keycatd
	
publish: web
	bash -c 'goreleaser release --rm-dist --release-notes <(./scripts/notes.sh)'

bindir:
	mkdir -p bindir

git-static: autogen
	mkdir -p data/version
	git log --date=iso  --pretty=format:'{ "commit": "%H", "date": "%ad"},' | perl -pe 'BEGIN{print "["}; END{print "]\n"}' | perl -pe 's/},]/}]/' > data/version/history
	echo $(GIT_VERSION) > data/version/current.server
	if [[ -e data/web ]]; then ( cd data/web; git describe --abbrev=8 --dirty --always --tags); else echo dev; fi > data/version/current.web

static: git-static
	go-bindata -prefix data/ -o static/data.go -pkg static data/...

dev-static: git-static
	go-bindata -debug -prefix data/ -o static/data.go -pkg static data/...

models/autogen.go: models/user.go models/team.go models/vault.go models/team_user.go models/vault_user.go models/invite.go models/token.go models/secret.go
	 scaneo -p models -u -o $@ $^

managers/autogen.go: managers/session_mgr.go
	 scaneo -p managers -u -o $@ $^

autogen: models/autogen.go managers/autogen.go

dev: autogen dev-static
	CompileDaemon -build 'go build -o bin/keycatd github.com/keydotcat/keycatd/cmd/keycatd' -command 'bin/keycatd' -color=true -directory=. -exclude-dir=bin -exclude-dir=web -exclude-dir=data/web -exclude-dir=scrips -exclude=tags -exclude-dir=vendor

test: test_db test_managers test_models test_api

test_db: autogen dev-static
	go test -v github.com/keydotcat/keycatd/db 

test_managers: autogen dev-static
	go test -v github.com/keydotcat/keycatd/managers 

test_models: autogen dev-static
	go test -v github.com/keydotcat/keycatd/models 

test_api: autogen dev-static
	go test -v github.com/keydotcat/keycatd/api

test_coverage: autogen dev-static
	go test -v -coverprofile db/cover.out -covermode atomic github.com/keydotcat/keycatd/db
	go test -v -coverprofile managers/cover.out -covermode atomic github.com/keydotcat/keycatd/managers 
	go test -v -coverprofile models/cover.out -covermode atomic github.com/keydotcat/keycatd/models 
	go test -v -coverprofile api/cover.out -covermode atomic github.com/keydotcat/keycatd/api
