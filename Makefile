all:
	$(MAKE) compress
	go build -o main main.go
	./main -p test.md

compress:
	uglifyjs static/scripts/src.js --output static/scripts/src.min.js
	uglifycss static/icons.css --output static/icons.min.css
	tailwindcss -i static/input.css -o static/output.css --minify
