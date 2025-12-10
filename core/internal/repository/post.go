package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// PostRepository handles database operations for posts
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) (*domain.Post, error)
	GetAll(ctx context.Context) ([]*domain.Post, error)
	GetByID(ctx context.Context, id int64) (*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) (*domain.Post, error)
	Delete(ctx context.Context, id int64) error
	IncrementLikes(ctx context.Context, id int64) error
	DecrementLikes(ctx context.Context, id int64) error
}

// SQLPostRepository implements PostRepository using SQL
type SQLPostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *sql.DB) PostRepository {
	return &SQLPostRepository{db: db}
}

// Create creates a new post
func (r *SQLPostRepository) Create(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	query := `
		INSERT INTO posts (user_id, content, image_url, likes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		post.UserID,
		post.Content,
		post.ImageURL,
		post.Likes,
		post.CreatedAt,
		post.UpdatedAt,
	).Scan(&post.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

// GetAll retrieves all posts with author information
func (r *SQLPostRepository) GetAll(ctx context.Context) ([]*domain.Post, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.likes, p.created_at, p.updated_at,
			u.id, u.username, u.email, u.full_name, u.bio, u.avatar_url
		FROM posts p
		LEFT JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		var post domain.Post
		var author domain.User

		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.Likes,
			&post.CreatedAt, &post.UpdatedAt,
			&author.ID, &author.Username, &author.Email, &author.FullName,
			&author.Bio, &author.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		post.Author = &author
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, nil
}

// GetByID retrieves a post by ID
func (r *SQLPostRepository) GetByID(ctx context.Context, id int64) (*domain.Post, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.content, p.image_url, p.likes, p.created_at, p.updated_at,
			u.id, u.username, u.email, u.full_name, u.bio, u.avatar_url
		FROM posts p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.id = $1
	`

	var post domain.Post
	var author domain.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID, &post.UserID, &post.Content, &post.ImageURL, &post.Likes,
		&post.CreatedAt, &post.UpdatedAt,
		&author.ID, &author.Username, &author.Email, &author.FullName,
		&author.Bio, &author.AvatarURL,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	post.Author = &author
	return &post, nil
}

// Update updates an existing post
func (r *SQLPostRepository) Update(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	query := `
		UPDATE posts
		SET content = $1, image_url = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		post.Content,
		post.ImageURL,
		post.UpdatedAt,
		post.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return post, nil
}

// Delete deletes a post
func (r *SQLPostRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// IncrementLikes increments the like count of a post
func (r *SQLPostRepository) IncrementLikes(ctx context.Context, id int64) error {
	query := `UPDATE posts SET likes = likes + 1 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment likes: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

// DecrementLikes decrements the like count of a post
func (r *SQLPostRepository) DecrementLikes(ctx context.Context, id int64) error {
	query := `UPDATE posts SET likes = GREATEST(likes - 1, 0) WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to decrement likes: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}
