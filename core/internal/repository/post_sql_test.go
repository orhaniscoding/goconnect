package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLPostRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()
	now := time.Now()

	post := &domain.Post{
		UserID:    1,
		Content:   "test content",
		ImageURL:  nil,
		Likes:     0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO posts (user_id, content, image_url, likes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(post.UserID, post.Content, post.ImageURL, post.Likes, post.CreatedAt, post.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	created, err := repo.Create(ctx, post)
	require.NoError(t, err)
	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, post.Content, created.Content)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLPostRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT p.id, p.user_id, p.content, p.image_url, p.likes, p.created_at, p.updated_at, u.id, u.username, u.email, u.full_name, u.bio, u.avatar_url FROM posts p LEFT JOIN users u ON p.user_id = u.id WHERE p.id = $1`

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "content", "image_url", "likes", "created_at", "updated_at",
		"id", "username", "email", "full_name", "bio", "avatar_url",
	}).AddRow(
		1, 1, "content", nil, 0, now, now,
		"1", "user1", "email@test.com", "User One", "bio", nil,
	)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(1).
		WillReturnRows(rows)

	post, err := repo.GetByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)
	assert.Equal(t, "content", post.Content)
	assert.Equal(t, "user1", *post.Author.Username)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLPostRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()

	query := `SELECT p.id, p.user_id, p.content, p.image_url, p.likes, p.created_at, p.updated_at, u.id, u.username, u.email, u.full_name, u.bio, u.avatar_url FROM posts p LEFT JOIN users u ON p.user_id = u.id WHERE p.id = $1`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(ctx, 999)
	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLPostRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()
	now := time.Now()

	query := `SELECT p.id, p.user_id, p.content, p.image_url, p.likes, p.created_at, p.updated_at, u.id, u.username, u.email, u.full_name, u.bio, u.avatar_url FROM posts p LEFT JOIN users u ON p.user_id = u.id ORDER BY p.created_at DESC`

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "content", "image_url", "likes", "created_at", "updated_at",
		"id", "username", "email", "full_name", "bio", "avatar_url",
	}).AddRow(
		1, 1, "content", nil, 0, now, now,
		"1", "user1", "email@test.com", "User One", "bio", nil,
	)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(rows)

	posts, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, int64(1), posts[0].ID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLPostRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()
	now := time.Now()

	post := &domain.Post{
		ID:        1,
		Content:   "updated content",
		UpdatedAt: now,
	}

	query := `UPDATE posts SET content = $1, image_url = $2, updated_at = $3 WHERE id = $4`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(post.Content, post.ImageURL, post.UpdatedAt, post.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	updated, err := repo.Update(ctx, post)
	require.NoError(t, err)
	assert.Equal(t, post.Content, updated.Content)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLPostRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()

	query := `DELETE FROM posts WHERE id = $1`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(ctx, 1)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLPostRepository_IncrementLikes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()

	query := `UPDATE posts SET likes = likes + 1 WHERE id = $1`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.IncrementLikes(ctx, 1)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLPostRepository_DecrementLikes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostRepository(db)
	ctx := context.Background()

	query := `UPDATE posts SET likes = GREATEST(likes - 1, 0) WHERE id = $1`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DecrementLikes(ctx, 1)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
