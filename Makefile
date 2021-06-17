all: forked-mp4ff-info forked-mp4ff-nallister forked-mp4ff-pslister forked-mp4ff-wvttlister

forked-mp4ff-info forked-mp4ff-nallister forked-mp4ff-pslister forked-mp4ff-wvttlister:
	go build -ldflags "-X github.com/jaypadia-frame/forked-mp4ff/mp4.commitVersion=$$(git describe --tags HEAD) -X github.com/jaypadia-frame/forked-mp4ff/mp4.commitDate=$$(git log -1 --format=%ct)" -o out/$@ cmd/$@/main.go

clean:
	rm -f out/*

install: all
	cp out/* $(GOPATH)/bin/

