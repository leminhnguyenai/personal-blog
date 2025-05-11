module github.com/leminhohoho/personal-blog/app

go 1.23.4

require (
	github.com/leminhohoho/personal-blog/pkgs/markdownparser v0.0.0
	github.com/leminhohoho/personal-blog/pkgs/filewatcher v0.0.0
	github.com/leminhohoho/personal-blog/pkgs/simplelog v0.0.0
    github.com/mattn/go-sqlite3 v1.14.28
)

replace github.com/leminhohoho/personal-blog/pkgs/markdownparser v0.0.0 => ../pkgs/markdownparser
replace github.com/leminhohoho/personal-blog/pkgs/filewatcher v0.0.0 => ../pkgs/filewatcher
replace github.com/leminhohoho/personal-blog/pkgs/simplelog v0.0.0 => ../pkgs/simplelog
