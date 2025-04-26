module github.com/leminhnguyenai/personal-blog/app

go 1.23.4

require (
    github.com/leminhnguyenai/personal-blog/services/markdownparser v0.0.0
)

replace github.com/leminhnguyenai/personal-blog/services/markdownparser v0.0.0 => ../services/markdownparser
