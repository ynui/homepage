.PHONY: build build-html build-cmd run-web run-cli clean

build: build-html build-cmd

build-html:
	python3 web/build.py

build-cmd:
	cd cli && go generate ./tui && go build -ldflags="-s -w" -o ../output/tui ./tui

run-web:
	open output/index.html

run-cli:
	./output/tui

clean:
	rm -rf output
