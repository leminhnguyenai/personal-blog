all:
	$(MAKE) compress
	go build -o main main.go
	./main -p test.md

compress:
	uglifyjs static/scripts/src.js --output static/scripts/src.min.js
	gzip -k static/scripts/src.min.js || true
	uglifycss static/icons.css --output static/icons.min.css 
	gzip -k static/icons.min.css || true
	tailwindcss -i static/input.css -o static/output.css --minify
	gzip -k static/output.css || true
