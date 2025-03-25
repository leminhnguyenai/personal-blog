all:
	$(MAKE) compress
	go build -o main main.go
	./main -p test.md

compress:
	uglifyjs static/scripts/src.js --output static/scripts/src.min.js
	tailwindcss -i static/input.css -o static/output.css --minify
