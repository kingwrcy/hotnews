package handler

import (
	"github.com/samber/do"
	"gorm.io/gorm"
)

type CommentHandler struct {
	injector *do.Injector
	db       *gorm.DB
}

func newCommentHandler(injector *do.Injector) (*CommentHandler, error) {
	return &CommentHandler{
		injector: injector,
		db:       do.MustInvoke[*gorm.DB](injector),
	}, nil
}

//func (h CommentHandler) Vote(c *gin.Context) {
//	cid := c.Query("cid")
//	action := c.Query("action")
//	var vote model.TbVote
//	userinfo := GetCurrentUser(c)
//	if userinfo == nil {
//		c.Redirect(302, "/u/login")
//		return
//	}
//
//	var comment model.TbComment
//	var exists int64
//	h.db.Model(&model.TbComment{}).Where("cid = ?", cid).First(&comment)
//
//	uid := userinfo.ID
//
//	if h.db.Model(&model.TbVote{}).Where("comment_id = ? and user_id = ?", comment.ID, uid).Count(&exists); exists == 0 {
//		var col string
//		if action == "u" {
//			vote.Action = "UP"
//			col = "upVote"
//		} else {
//			vote.Action = "Down"
//			col = "downVote"
//		}
//		vote.UserID = uid
//		vote.PostID = nil
//		vote.CommentID = &comment.ID
//
//		h.db.Transaction(func(tx *gorm.DB) error {
//			if err := tx.Save(&vote).Error; err != nil {
//				return err
//			}
//			if err := tx.Model(&model.TbComment{}).Where("id =?", comment.ID).Update(col, gorm.Expr(col+"+1")).Error; err != nil {
//				return err
//			}
//			return nil
//		})
//
//	}
//	refer := c.GetHeader("refer")
//	if refer == "" {
//		refer = "/"
//	}
//
//	c.Redirect(302, refer)
//}
