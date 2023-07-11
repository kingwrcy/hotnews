package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"github.com/spf13/cast"
	"gorm.io/gorm"
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
	userinfo := GetCurrentUser(c)

	var posts []model.TbPost
	result := QueryPosts(p.db, vo.QueryPostsRequest{
		Userinfo: userinfo,
		Page:     1,
		Size:     25,
		PostPID:  cast.ToString(c.Param("pid")),
	})

	var uid uint = 0
	if userinfo != nil {
		uid = userinfo.ID
	}
	posts = result["posts"].([]model.TbPost)

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
		if strings.HasPrefix(host, "www.") {
			_, host, _ = strings.Cut(host, "www.")
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
	page := c.DefaultQuery("p", "1")

	tagName := c.Param("tag")

	c.HTML(200, "index.html", OutputCommonSession(c, gin.H{
		"selected": "/",
	}, QueryPosts(p.db, vo.QueryPostsRequest{
		Userinfo:  userinfo,
		Tags:      []string{tagName},
		OrderType: "index",
		Page:      cast.ToInt64(page),
		Size:      25,
	})))
}
func (p PostHandler) SearchByType(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	page := c.DefaultQuery("p", "1")
	typeName := c.Param("type")
	c.HTML(200, "index.html", OutputCommonSession(c, gin.H{
		"selected": "/",
	}, QueryPosts(p.db, vo.QueryPostsRequest{
		Userinfo:  userinfo,
		Type:      typeName,
		OrderType: "index",
		Page:      cast.ToInt64(page),
		Size:      25,
	})))
}

func QueryPosts(db *gorm.DB, request vo.QueryPostsRequest) gin.H {

	tx := db.Table("tb_post p").Where("status = 'Active'")
	if request.Type != "" {
		tx.Where("type = ?", request.Type)
	}
	if request.Begin != nil {
		tx.Where("created_at >= ?", request.Begin)
	}
	if request.End != nil {
		tx.Where("created_at <= ?", request.End)
	}
	if request.Domain != "" {
		tx.Where("domain = ?", request.Domain)
	}
	if request.PostPID != "" {
		tx.Where("pid = ?", request.PostPID)
	}
	if request.Userinfo != nil {
		subQuery := db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'POST' and action ='UP'", request.Userinfo.ID)

		tx.Select("p.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted")
		tx.Joins("LEFT JOIN (?) AS vote ON p.id = vote.target_id", subQuery)
	}
	if len(request.Tags) > 0 {
		tx.InnerJoins(",tb_post_tag pt,tb_tag t")
		tx.Where("status = 'Active' and t.id = pt.tb_tag_id and pt.tb_post_id = p.id and t.name in (?)", request.Tags)
	}

	var total int64

	tx.Count(&total)

	if request.OrderType == "index" {
		tx.Order("point desc,created_at desc")
	} else {
		tx.Order("created_at desc")
	}

	var posts []model.TbPost
	tx.Preload("Tags").Preload("User").Limit(int(request.Size)).Offset(int((request.Page - 1) * request.Size)).Find(&posts)

	totalPage := total / request.Size
	if total%request.Size > 0 {
		totalPage = totalPage + 1
	}
	return gin.H{
		"posts":       posts,
		"totalPage":   totalPage,
		"hasNext":     request.Page < totalPage,
		"hasPrev":     request.Page > 1,
		"currentPage": request.Page,
	}
}
