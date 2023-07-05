package vo

type InspectRequest struct {
	Result      string `form:"result"`
	Reason      string `form:"reason"`
	PostID      uint   `form:"post_id"`
	CommentID   uint   `form:"comment_id"`
	Action      string `form:"action"`
	InspectType string `form:"inspect_type"`
}
