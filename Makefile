.PHONY: build build-html build-cmd run-web run-cli demo clean

build: build-html build-cmd

build-html:
	python3 web/build.py

build-cmd:
	cd cli && go generate ./tui && go build -ldflags="-s -w" -o ../output/tui ./tui

run-web: build-html
	open output/index.html

run-cli: build-cmd
	./output/tui

demo: build-cmd
	vhs demo/demo.tape

clean:
	rm -rf output
