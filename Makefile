.PHONY: build build-html build-cmd clean

build: build-html build-cmd

build-html:
	python3 web/build.py

build-cmd:
	cd cli && go build -ldflags="-s -w" -o ../output/tui ./tui

clean:
	rm -rf output
