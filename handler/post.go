package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/url"
	"strings"
)

type PostHandler struct {
	injector *do.Injector
	db       *gorm.DB
}

func NewPostHandler(injector *do.Injector) (*PostHandler, error) {
	return &PostHandler{
		injector: injector,
		db:       do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (p PostHandler) Detail(c *gin.Context) {
	var posts []model.TbPost
	p.db.Model(model.TbPost{}).Preload("Comments.User").
		Preload(clause.Associations).Where("pid = ? ", c.Param("pid")).First(&posts)
	var rootComments []model.TbComment
	if len(posts) > 0 {
		p.db.Model(&model.TbComment{}).Preload("User").Where("post_id = ? and parent_comment_id is null", posts[0].ID).Find(&rootComments)

		buildCommentTree(&rootComments, p.db)
		posts[0].Comments = rootComments
	}
	c.HTML(200, "post.html", OutputCommonSession(c, gin.H{
		"posts":    posts,
		"selected": "detail",
	}))
}

func buildCommentTree(comments *[]model.TbComment, db *gorm.DB) {
	for i := range *comments {
		var children []model.TbComment
		db.Preload("User").Where("post_id = ? and parent_comment_id = ?", (*comments)[i].PostID, (*comments)[i].ID).Find(&children)
		(*comments)[i].Comments = children
		buildCommentTree(&(*comments)[i].Comments, db)
	}
}

func (p PostHandler) Add(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil {
		c.Redirect(302, "/u/login")
		return
	}
	uid := userinfo.ID

	var request vo.NewPostRequest
	if err := c.Bind(&request); err != nil {
		c.HTML(200, "new.html", OutputCommonSession(c, gin.H{
			"msg":      "参数异常",
			"selected": "new",
		}))
		return
	}
	log.Printf("params:%+v", request)
	if len(request.TagIDs) == 0 || len(request.TagIDs) > 5 {
		c.HTML(200, "new.html", OutputCommonSession(c, gin.H{
			"msg":      "标签最少1个,最多5个",
			"selected": "new",
		}))
		return
	}
	if request.Type == "" {
		c.HTML(200, "new.html", OutputCommonSession(c, gin.H{
			"msg":      "类型必填",
			"selected": "new",
		}))
		return
	}

	var tags []model.TbTag
	for _, v := range request.TagIDs {
		tags = append(tags, model.TbTag{
			Model: gorm.Model{ID: v},
		})
	}
	var user model.TbUser

	p.db.Model(model.TbUser{}).Where("id=?", uid).First(&user)
	status := "WAIT_APPROVE"
	if request.Type == "ask" {
		status = "Active"
	}

	host := ""
	if request.Type == "link" {
		urlParsed, _ := url.Parse(request.Link)
		host = urlParsed.Host
	}

	post := model.TbPost{
		Title:        strings.Trim(request.Title, " "),
		Link:         strings.Trim(request.Link, " "),
		Status:       status,
		Content:      strings.Trim(request.Content, " "),
		UpVote:       0,
		DownVote:     0,
		Type:         request.Type,
		Tags:         tags,
		User:         model.TbUser{Model: gorm.Model{ID: uid}},
		Domain:       host,
		Pid:          RandStringBytesMaskImpr(8),
		CommentCount: 0,
	}

	err := p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&post).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.TbUser{Model: gorm.Model{ID: uid}}).Update("postCount", user.PostCount+1).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.HTML(200, "new.html", OutputCommonSession(c, gin.H{
			"msg":      "系统错误",
			"selected": "new",
		}))
		return
	}
	if request.Type == "ask" {
		c.Redirect(302, "/p/"+post.Pid)
		return
	}
	c.HTML(200, "new.html", OutputCommonSession(c, gin.H{
		"msg":      "提交成功,等待审核",
		"selected": "new",
	}))
	return
}

func (p PostHandler) AddComment(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil {
		c.Redirect(302, "/u/login")
		return
	}
	uid := userinfo.ID
	var comment model.TbComment
	var request vo.NewCommentRequest
	err := c.Bind(&request)
	if err != nil {
		c.Redirect(302, "/")
		return
	}
	comment.PostID = request.PostID
	if request.ParentCommentId == 0 {
		comment.ParentCommentID = nil
	} else {
		comment.ParentCommentID = &request.ParentCommentId
	}

	comment.Content = request.Content
	comment.UserID = uid
	comment.UpVote = 0
	comment.DownVote = 0
	comment.CID = RandStringBytesMaskImpr(8)
	var redirectUrl = "/p/" + request.PostPID + "#c-" + comment.CID

	err = p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&comment).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.TbPost{}).Where("id = ?", request.PostID).Update("commentCount", gorm.Expr("commentCount + 1")).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.Redirect(302, "/")
		return
	}
	c.Redirect(302, redirectUrl)
}

func (p PostHandler) Vote(c *gin.Context) {
	pid := c.Query("pid")
	action := c.Query("action")
	var vote model.TbVote
	userinfo := GetCurrentUser(c)
	if userinfo == nil {
		c.Redirect(302, "/u/login")
		return
	}

	var post model.TbPost
	var exists int64
	p.db.Model(&model.TbPost{}).Where("pid = ?", pid).First(&post)

	uid := userinfo.ID

	if p.db.Model(&model.TbVote{}).Where("post_id = ? and user_id = ?", post.ID, uid).Count(&exists); exists == 0 {
		log.Printf("ahhaha")
		var col string
		if action == "u" {
			vote.Action = "UP"
			col = "upVote"
		} else {
			vote.Action = "Down"
			col = "downVote"
		}
		vote.UserID = uid
		vote.PostID = &post.ID
		vote.CommentID = nil

		p.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&vote).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.TbPost{}).Where("id =?", post.ID).Update(col, gorm.Expr(col+"+1")).Error; err != nil {
				return err
			}
			return nil
		})

	}
	refer := c.GetHeader("refer")
	if refer == "" {
		refer = "/"
	}

	c.Redirect(302, refer)
}
