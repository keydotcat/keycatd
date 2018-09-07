ROOT = github.com/keydotcat/server
SUF=
ifdef GOOS
       SUF=.$(GOOS)
ifeq ($(GOOS),windows)
       SUF=.windows.exe
endif
endif
ifdef GOARCH
       SUF=${SUF}.$(GOARCH)
endif

.PHONY: static web

cmds:
	go get github.com/jteeuwen/go-bindata/...
	go get github.com/acasajus/scaneo
	go get github.com/githubnemo/CompileDaemon

web:
	./build_web.sh

keycatd: web static bindir
	go build -o bin/keycatd${SUF} ${ROOT}/cmd/keycatd

bindir:
	mkdir -p bindir

git-static: autogen
	mkdir -p data/version
	git log --date=iso  --pretty=format:'{ "commit": "%H", "date": "%ad"},' | perl -pe 'BEGIN{print "["}; END{print "]\n"}' | perl -pe 's/},]/}]/' > data/version/history
	git describe --abbrev=8 --dirty --always --tags > data/version/current.server
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
	CompileDaemon -build 'go build -o bin/keycatd github.com/keydotcat/server/cmd/keycatd' -command 'bin/keycatd' -directory . -color=true -exclude=tags -exclude vendor -exclude .git

test: test_db test_managers test_models test_api

test_db: autogen
	go test -v github.com/keydotcat/server/db 

test_managers: autogen
	go test -v github.com/keydotcat/server/managers 

test_models: autogen
	go test -v github.com/keydotcat/server/models 

test_api: autogen
	go test -v github.com/keydotcat/server/api

test_coverage:
	go test -v -coverprofile db/cover.out -covermode atomic github.com/keydotcat/server/db
	go test -v -coverprofile managers/cover.out -covermode atomic github.com/keydotcat/server/managers 
	go test -v -coverprofile models/cover.out -covermode atomic github.com/keydotcat/server/models 
	go test -v -coverprofile api/cover.out -covermode atomic github.com/keydotcat/server/api
