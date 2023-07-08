package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"github.com/spf13/cast"
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

	userinfo := GetCurrentUser(c)
	var uid uint = 0
	if userinfo != nil {
		uid = userinfo.ID
	}

	var rootComments []model.TbComment
	if len(posts) > 0 {
		if userinfo != nil {
			subQuery := p.db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'COMMENT' and action ='UP'", uid)

			p.db.Table("tb_comment c").Select("c.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted").Joins("LEFT JOIN (?) AS vote ON c.id = vote.target_id", subQuery).
				Preload("User").Where("post_id = ? and parent_comment_id is null", posts[0].ID).Find(&rootComments)

		} else {
			p.db.Table("tb_comment c").Select("c.*").
				Preload("User").Where("post_id = ? and parent_comment_id is null", posts[0].ID).
				Find(&rootComments)

		}

		buildCommentTree(&rootComments, p.db, uid)
		posts[0].Comments = rootComments
	}
	c.HTML(200, "post.html", OutputCommonSession(c, gin.H{
		"posts":    posts,
		"selected": "detail",
	}))
}

func buildCommentTree(comments *[]model.TbComment, db *gorm.DB, uid uint) {
	subQuery := db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'COMMENT' and action ='UP'", uid)
	for i := range *comments {
		var children []model.TbComment
		if uid > 0 {
			db.Table("tb_comment c").Select("c.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted").
				Joins("LEFT JOIN (?) AS vote ON c.id = vote.target_id", subQuery).Preload("User").Where("post_id = ? and parent_comment_id = ?", (*comments)[i].PostID, (*comments)[i].ID).Find(&children)
		} else {
			db.Model(&model.TbComment{}).
				Joins("LEFT JOIN (?) AS vote ON c.id = vote.target_id", subQuery).Preload("User").Where("post_id = ? and parent_comment_id = ?", (*comments)[i].PostID, (*comments)[i].ID).Find(&children)
		}
		(*comments)[i].Comments = children
		buildCommentTree(&(*comments)[i].Comments, db, uid)
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
	status := "Active"

	host := ""
	if request.Type == "link" {
		urlParsed, _ := url.Parse(request.Link)
		host = urlParsed.Host
		if strings.HasPrefix(host, "www") {
			_, host, _ = strings.Cut(host, "www")
		}
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
	c.Redirect(302, "/p/"+post.Pid)
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

func (p PostHandler) SearchByTag(c *gin.Context) {
	userinfo := GetCurrentUser(c)

	var posts []model.TbPost
	var total int64
	var totalPage int64
	size := 25
	page := c.DefaultQuery("p", "1")
	pageNumber := cast.ToInt(page)

	tagName := c.Param("tag")

	if userinfo != nil {
		subQuery := p.db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'POST' and action ='UP'", userinfo.ID)

		p.db.Table("tb_post p").Select("p.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted").
			Joins("LEFT JOIN (?) AS vote ON p.id = vote.target_id", subQuery).
			InnerJoins(",tb_post_tag pt,tb_tag t").
			Preload("User").Preload("Tags").
			Where("status = 'Active' and t.id = pt.tb_tag_id and pt.tb_post_id = p.id and t.name = ?", tagName).
			Order("created_at desc").
			Offset((pageNumber - 1) * size).Limit(size).Find(&posts)
	} else {
		p.db.Table("tb_post p").Preload("User").Preload("Tags").
			InnerJoins(",tb_post_tag pt,tb_tag t").
			Preload("User").Preload("Tags").
			Where("status = 'Active' and t.id = pt.tb_tag_id and pt.tb_post_id = p.id and t.name = ?", tagName).
			Order("created_at desc").
			Offset((pageNumber - 1) * size).Limit(size).Find(&posts)
	}
	p.db.Table("tb_post p").InnerJoins(",tb_post_tag pt,tb_tag t").
		Where("status = 'Active' and t.id = pt.tb_tag_id and pt.tb_post_id = p.id and t.name = ?", tagName).
		Count(&total)

	totalPage = total / int64(size)
	if total%int64(size) > 0 {
		totalPage = totalPage + 1
	}

	c.HTML(200, "index.html", OutputCommonSession(c, gin.H{
		"posts":       posts,
		"totalPage":   totalPage,
		"hasNext":     int64(pageNumber) < totalPage,
		"hasPrev":     int64(pageNumber) > 1,
		"currentPage": pageNumber,
	}))
}
func (p PostHandler) SearchByType(c *gin.Context) {
	userinfo := GetCurrentUser(c)

	var posts []model.TbPost
	var total int64
	var totalPage int64
	size := 25
	page := c.DefaultQuery("p", "1")
	pageNumber := cast.ToInt(page)

	typeName := c.Param("type")

	if userinfo != nil {
		subQuery := p.db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'POST' and action ='UP'", userinfo.ID)

		p.db.Table("tb_post p").Select("p.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted").
			Joins("LEFT JOIN (?) AS vote ON p.id = vote.target_id", subQuery).
			Preload("User").Preload("Tags").
			Where("status = 'Active' and type = ? ", typeName).
			Order("created_at desc").
			Offset((pageNumber - 1) * size).Limit(size).Find(&posts)
	} else {
		p.db.Table("tb_post p").Preload("User").Preload("Tags").
			Preload("User").Preload("Tags").
			Where("status = 'Active' and type = ? ", typeName).
			Order("created_at desc").
			Offset((pageNumber - 1) * size).Limit(size).Find(&posts)
	}
	p.db.Model(&model.TbPost{}).InnerJoins(",tb_post_tag pt,tb_tag t").
		Where("status = 'Active' and type = ? ", typeName).
		Count(&total)

	totalPage = total / int64(size)
	if total%int64(size) > 0 {
		totalPage = totalPage + 1
	}

	c.HTML(200, "index.html", OutputCommonSession(c, gin.H{
		"posts":       posts,
		"totalPage":   totalPage,
		"hasNext":     int64(pageNumber) < totalPage,
		"hasPrev":     int64(pageNumber) > 1,
		"currentPage": pageNumber,
	}))
}
