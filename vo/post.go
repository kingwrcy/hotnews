package vo

type NewPostRequest struct {
	Title   string `form:"title"`
	Link    string `form:"link"`
	TagIDs  []uint `form:"tagIDs"`
	Content string `form:"content"`
	Type    string `form:"type"`
}
