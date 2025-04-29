module github.com/leminhohoho/personal-blog/app

go 1.23.4

require (
    github.com/leminhohoho/personal-blog/services/markdownparser v0.0.0
) 

replace github.com/leminhohoho/personal-blog/services/markdownparser v0.0.0 => ../services/markdownparser
