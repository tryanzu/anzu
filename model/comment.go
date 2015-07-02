package model

type CommentForm struct {
	Content string `json:"content" binding:"required"`
}
