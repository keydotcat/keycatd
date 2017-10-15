GIT_VERSION = $(shell git describe --abbrev=8 --dirty --always 2>/dev/null)

.PHONY: static

static:
	mkdir -p data/version
	git log --pretty=format:'{ "commit": "%H", "date": "%aI"},' | perl -pe 'BEGIN{print "["}; END{print "]\n"}' | perl -pe 's/},]/}]/' > data/version/history
	echo ${GIT_VERSION} > data/version/current
	go-bindata -o static/data.go -pkg static data/**
