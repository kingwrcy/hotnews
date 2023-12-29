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
	Top     string `form:"top"`
}

type NewCommentRequest struct {
	Content         string `form:"content"`
	PostID          uint   `form:"post_id"`
	ParentCommentId uint   `form:"parent_comment_id"`
	PostPID         string `form:"post_pid"`
}

type QueryPostsRequest struct {
	Userinfo  *Userinfo
	Type      string     `form:"type"`
	Tags      []string   `form:"tags"`
	Begin     *time.Time `form:"begin"`
	End       *time.Time `form:"end"`
	Q         string     `form:"q"`
	OrderType string     `form:"orderType"`
	Page      int64      `form:"page"`
	Size      int64      `form:"size"`
	Domain    string     `form:"domain"`
	PostPID   string     `form:"post_pid"`
}
