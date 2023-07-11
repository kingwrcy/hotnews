package vo

import (
	"time"
)

type NewPostRequest struct {
	Title   string `form:"title"`
	Link    string `form:"link"`
	TagIDs  []uint `form:"tagIDs[]"`
	Content string `form:"content"`
	Type    string `form:"type"`
}

type NewCommentRequest struct {
	Content         string `form:"content"`
	PostID          uint   `form:"post_id"`
	ParentCommentId uint   `form:"parent_comment_id"`
	PostPID         string `form:"post_pid"`
}

type QueryPostsRequest struct {
	Userinfo  *Userinfo
	Type      string
	Tags      []string
	Begin     *time.Time
	End       *time.Time
	Q         string
	OrderType string
	Page      int64
	Size      int64
	Domain    string
	PostPID   string
}
