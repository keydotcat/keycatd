GIT_VERSION = $(shell git describe --abbrev=8 --dirty --always 2>/dev/null)

.PHONY: static

cmds:
	go get github.com/jteeuwen/go-bindata/...
	go get github.com/acasajus/scaneo
	go get github.com/githubnemo/CompileDaemon

git-static: models/autogen.go
	mkdir -p data/version
	git log --pretty=format:'{ "commit": "%H", "date": "%aI"},' | perl -pe 'BEGIN{print "["}; END{print "]\n"}' | perl -pe 's/},]/}]/' > data/version/history
	echo ${GIT_VERSION} > data/version/current

static: git-static
	go-bindata -prefix data/ -o static/data.go -pkg static data/...

dev-static: git-static
	go-bindata -debug -prefix data/ -o static/data.go -pkg static data/...

models/autogen.go: models/user.go models/team.go models/vault.go models/team_user.go models/vault_user.go models/invite.go models/token.go models/secret.go
	 scaneo -p models -u -o $@ $^

autoreload: models/autogen.go dev-static
	CompileDaemon -build 'go build -o bin/keycatd github.com/keydotcat/backend/cmd/keycatd' -command 'bin/keycatd' -directory . -color=true -exclude=tags -exclude vendor -exclude .git
