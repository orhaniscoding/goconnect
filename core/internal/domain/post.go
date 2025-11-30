package domain

import "time"

// Post represents a social media post
type Post struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	ImageURL  *string   `json:"image_url,omitempty" db:"image_url"`
	Likes     int       `json:"likes" db:"likes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Author    *User     `json:"author,omitempty"`
}

// CreatePostRequest represents a request to create a new post
type CreatePostRequest struct {
	UserID   int64   `json:"-"`
	Content  string  `json:"content" binding:"required,min=1,max=5000"`
	ImageURL *string `json:"image_url,omitempty"`
}

// UpdatePostRequest represents a request to update a post
type UpdatePostRequest struct {
	ID       int64   `json:"-"`
	UserID   int64   `json:"-"`
	Content  string  `json:"content" binding:"required,min=1,max=5000"`
	ImageURL *string `json:"image_url,omitempty"`
}

// PostWithAuthor represents a post with author information
type PostWithAuthor struct {
	Post
	AuthorUsername string  `json:"author_username" db:"author_username"`
	AuthorEmail    string  `json:"author_email" db:"author_email"`
	AuthorFullName *string `json:"author_full_name,omitempty" db:"author_full_name"`
	AuthorBio      *string `json:"author_bio,omitempty" db:"author_bio"`
	AuthorAvatar   *string `json:"author_avatar_url,omitempty" db:"author_avatar_url"`
}
