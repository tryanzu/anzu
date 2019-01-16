package events

const (
	POSTS_NEW       = "posts:new"
	POSTS_COMMENT   = "posts:comment"
	POST_VIEW       = "posts:view"
	POSTS_REACHED   = "posts:reached"
	RECENT_ACTIVITY = "activity:recent"

	COMMENT_DELETE          = "comments:delete"
	COMMENT_UPDATE          = "comments:update"
	COMMENT_UPVOTE          = "comments:upvote"
	COMMENT_VOTE            = "comments:vote"
	COMMENT_UPVOTE_REMOVE   = "comments:upvote.remove"
	COMMENT_DOWNVOTE        = "comments:downvote"
	COMMENT_DOWNVOTE_REMOVE = "comments:downvote.remove"

	RAW_EMIT = "transmit:emit"
)
