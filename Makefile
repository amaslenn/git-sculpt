app := git-sculpt

all: darwin linux windows

clean:
	rm -rf build/

build = GOOS=$(1) GOARCH=$(2) go build -o build/$(2)/$(1)/$(app)$(3)

darwin:
	$(call build,darwin,amd64,)

linux:
	$(call build,linux,amd64,)

windows:
	$(call build,windows,amd64,.exe)

