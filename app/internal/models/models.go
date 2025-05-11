package models

import (
	"context"
)

type Blog struct {
	Name        string
	Path        string
	HTMLContent string
	ModTime     string
	Tags        []string
}

type BlogRepository interface {
	GetAllPosts(ctx context.Context) ([]*Blog, error)
	GetPostByID(ctx context.Context, filename string) (*Blog, error)
	AddPost(ctx context.Context, blog Blog) error
	UpdatePost(ctx context.Context, blog Blog) error
	DeletePost(ctx context.Context, filename string) error
}
