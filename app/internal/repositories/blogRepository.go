package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/leminhohoho/personal-blog/app/internal/models"
)

type SQLiteBlogRepository struct {
	db *sql.DB
}

func NewSQLiteBlogRepository(db *sql.DB) *SQLiteBlogRepository {
	return &SQLiteBlogRepository{
		db: db,
	}
}

func (b *SQLiteBlogRepository) GetPost(ctx context.Context, filename string) (*models.Blog, error) {
	errChan := make(chan error)
	respChan := make(chan models.Blog)

	go func() {
		rows, err := b.db.Query(`
        SELECT name, file_path, content, last_modified_time FROM posts
        WHERE name = ?
        LIMIT 1
        `)
		if err != nil {
			errChan <- err
		}

		for rows.Next() {
			var blog models.Blog

			if err := rows.Scan(&blog.Name, &blog.Path, &blog.HTMLContent, &blog.ModTime); err != nil {
				errChan <- err
			}

			respChan <- blog
		}

		errChan <- fmt.Errorf("No blog is found with the filename: %s\n", filename)
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("Stopped by context\n")
	case err := <-errChan:
		return nil, err
	case resp := <-respChan:
		return &resp, nil
	}
}

func (b *SQLiteBlogRepository) AddPost(ctx context.Context, blog models.Blog) error {
	errChan := make(chan error)

	go func() {
		_, err := b.db.Exec(`
            INSERT INTO posts(file_path, name, content, last_modified_time)
            VALUES (?,?,?,?)
        `, blog.Path, blog.Name, blog.HTMLContent, blog.ModTime)
		if err != nil {
			errChan <- err
		}

		errChan <- nil
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("Stopped by context\n")
	case err := <-errChan:
		return err
	}

}

func (b *SQLiteBlogRepository) UpdatePost(ctx context.Context, blog models.Blog) error {
	errChan := make(chan error)

	go func() {
		_, err := b.db.Exec(`
            UPDATE posts
            SET file_path = ?, content = ?, last_modified_time = ?
            WHERE name = ?
        `, blog.Path, blog.HTMLContent, blog.ModTime, blog.Name)
		if err != nil {
			errChan <- err
		}

		errChan <- nil
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("Stopped by context\n")
	case err := <-errChan:
		return err
	}
}

func (b *SQLiteBlogRepository) DeletePost(ctx context.Context, filename string) error {
	errChan := make(chan error)

	go func() {
		_, err := b.db.Exec("DELETE FROM posts WHERE name = ?", filename)
		if err != nil {
			errChan <- err
		}

		errChan <- nil
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("Stopped by context\n")
	case err := <-errChan:
		return err
	}
}
