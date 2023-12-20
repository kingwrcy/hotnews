package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"log"
	"time"
)

type IndexHandler struct {
	injector *do.Injector
	db       *gorm.DB
}

func NewIndexHandler(injector *do.Injector) (*IndexHandler, error) {
	return &IndexHandler{
		injector: injector,
		db:       do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (i *IndexHandler) Index(c *gin.Context) {
	userinfo := GetCurrentUser(c)

	begin := time.Now().AddDate(0, 0, -7)

	page := c.DefaultQuery("p", "1")

	c.HTML(200, "index.gohtml", OutputCommonSession(i.db, c, gin.H{
		"selected": "/",
	}, QueryPosts(i.db, vo.QueryPostsRequest{
		Userinfo:  userinfo,
		Begin:     &begin,
		OrderType: "index",
		Page:      cast.ToInt64(page),
		Size:      25,
	})))
}

func (i *IndexHandler) ToSearch(c *gin.Context) {
	c.HTML(200, "search.gohtml", OutputCommonSession(i.db, c, gin.H{
		"selected": "search",
	}))
}

func (i *IndexHandler) DoSearch(c *gin.Context) {
	var request vo.QueryPostsRequest
	c.Bind(&request)
	request.Size = 25
	if request.Page <= 0 {
		request.Page = 1
	}
	c.HTML(200, "search.gohtml", OutputCommonSession(i.db, c, gin.H{
		"selected":  "search",
		"condition": request,
	}, QueryPosts(i.db, request)))
}

func (i *IndexHandler) ToNew(c *gin.Context) {
	var tags []model.TbTag
	i.db.Model(&model.TbTag{}).Where("parent_id is null").Preload("Children").Find(&tags)
	c.HTML(200, "new.gohtml", OutputCommonSession(i.db, c, gin.H{
		"selected": "new",
		"tags":     tags,
	}))
}

func (i *IndexHandler) ToPost(c *gin.Context) {
	c.HTML(200, "post.gohtml", OutputCommonSession(i.db, c, gin.H{}))
}
func (i *IndexHandler) ToResetPwd(c *gin.Context) {
	c.HTML(200, "resetPwd.gohtml", OutputCommonSession(i.db, c, gin.H{}))
}
func (i *IndexHandler) ToBeInvited(c *gin.Context) {
	c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(i.db, c, gin.H{}))
}
func (i *IndexHandler) ToTags(c *gin.Context) {
	var tags []model.TbTag

	i.db.Model(&model.TbTag{}).Where("parent_id is null").Preload("Children").Find(&tags)
	c.HTML(200, "tags.gohtml", OutputCommonSession(i.db, c, gin.H{
		"tags":     tags,
		"selected": "tags",
	}))
}
func (i *IndexHandler) ToWaitApproved(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	var waitApprovedList []model.TbPost
	if userinfo != nil {
		if userinfo.Role == "admin" || userinfo.Role == "inspector" {
			i.db.Model(&model.TbPost{}).Preload("User").Preload("Tags").
				Where("status = 'WAIT_APPROVE'").Order("created_at desc").
				Find(&waitApprovedList)
			if len(waitApprovedList) == 0 {
				c.Redirect(302, "/")
				return
			}
		}
	} else {
		c.Redirect(302, "/u/login")
		return
	}

	c.HTML(200, "wait.gohtml", OutputCommonSession(i.db, c, gin.H{
		"posts":        waitApprovedList,
		"waitApproved": len(waitApprovedList),
		"selected":     "approve",
	}))
}

func (i *IndexHandler) History(c *gin.Context) {
	userinfo := GetCurrentUser(c)

	page := c.DefaultQuery("p", "1")

	c.HTML(200, "index.gohtml", OutputCommonSession(i.db, c, gin.H{
		"selected": "history",
	}, QueryPosts(i.db, vo.QueryPostsRequest{
		Userinfo: userinfo,
		Page:     cast.ToInt64(page),
		Size:     25,
	})))
}

func (i *IndexHandler) ToComments(c *gin.Context) {
	page := c.DefaultQuery("p", "1")
	size := 25
	var comments []model.TbComment
	var total int64
	var totalPage int64
	pageNumber := cast.ToInt(page)
	userinfo := GetCurrentUser(c)

	if userinfo != nil {
		subQuery := i.db.Table("tb_vote").Select("target_id").Where("tb_user_id = ? and type = 'COMMENT' and action ='UP'", userinfo.ID)

		i.db.Table("tb_comment c").Select("c.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted").Joins("LEFT JOIN (?) AS vote ON c.id = vote.target_id", subQuery).Preload("Post").
			Preload("User").Order("created_at desc").Limit(int(size)).Offset((pageNumber - 1) * size).Find(&comments)
	} else {
		i.db.Model(model.TbComment{}).Preload("Post").
			Preload("User").Order("created_at desc").Limit(int(size)).Offset((pageNumber - 1) * size).Find(&comments)
	}

	i.db.Model(model.TbComment{}).Count(&total)
	totalPage = total / int64(size)
	if total%int64(size) > 0 {
		totalPage = totalPage + 1
	}
	c.HTML(200, "comments.gohtml", OutputCommonSession(i.db, c, gin.H{
		"selected":    "comment",
		"comments":    comments,
		"totalPage":   totalPage,
		"hasNext":     pageNumber < int(totalPage),
		"hasPrev":     pageNumber > 1,
		"currentPage": pageNumber,
	}))
}

func (i *IndexHandler) Vote(c *gin.Context) {
	id := c.Query("id")
	action := c.Query("action")
	targetType := c.Query("type")
	var vote model.TbVote
	userinfo := GetCurrentUser(c)
	if userinfo == nil {
		c.Redirect(302, "/u/login")
		return
	}

	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}

	uid := userinfo.ID

	var exists int64
	var targetID uint
	var message model.TbMessage

	message.FromUserID = 999999999
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	message.Read = "N"

	if targetType == "POST" {
		var item model.TbPost
		i.db.Model(&model.TbPost{}).Where("pid = ?", id).First(&item)
		targetID = item.ID
		if item.UserID == uid {
			c.Redirect(302, refer)
			return
		}
		message.ToUserID = item.UserID
		message.Content = fmt.Sprintf("<a class='bLink' href='/u/profile/%s'>%s</a>给你的主题<a class='bLink' href='/p/%s'>%s</a>点赞了",
			userinfo.Username, userinfo.Username, item.Pid, item.Title)
	} else if targetType == "COMMENT" {
		var item model.TbComment
		i.db.Model(&model.TbComment{}).Preload("Post").Where("cid = ?", id).First(&item)
		targetID = item.ID
		if item.UserID == uid {
			log.Printf("comment item.UserID == uid ")

			c.Redirect(302, refer)
			return
		}
		message.ToUserID = item.UserID
		message.Content = fmt.Sprintf("<a class='bLink' href='/u/profile/%s'>%s</a>给你的<a class='bLink' href='/p/%s#c-%s'>评论</a>点赞了",
			userinfo.Username, userinfo.Username, item.Post.Pid, item.CID)
	}

	if i.db.Model(&model.TbVote{}).Where("target_id = ? and user_id = ?  and type = ?", targetID, uid, targetType).Count(&exists); exists == 0 {
		log.Printf("comment item.UserID == 0 ")
		var col string
		if action == "u" {
			vote.Action = "UP"
			col = "upVote"
		} else {
			vote.Action = "Down"
			col = "downVote"
		}
		vote.UserID = uid
		vote.TargetID = targetID
		vote.Type = targetType

		i.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&vote).Error; err != nil {
				return err
			}
			if targetType == "POST" {
				if err := tx.Model(&model.TbPost{}).Where("id =?", targetID).Update(col, gorm.Expr(col+"+1")).Error; err != nil {
					return err
				}
			} else if targetType == "COMMENT" {
				if err := tx.Model(&model.TbComment{}).Where("id =?", targetID).Update(col, gorm.Expr(col+"+1")).Error; err != nil {
					return err
				}
			}
			if err := tx.Save(&message).Error; err != nil {
				return err
			}
			return nil
		})

	}

	c.Redirect(302, refer)
}

func (i *IndexHandler) Moderation(c *gin.Context) {
	page := c.DefaultQuery("p", "1")
	size := 25
	var logs []model.TbInspectLog
	var total int64
	var totalPage int64
	pageNumber := cast.ToInt(page)

	i.db.Model(&model.TbInspectLog{}).Preload("Inspector").Preload("Post").Limit(size).Offset((pageNumber - 1) * size).Find(&logs)
	i.db.Model(&model.TbInspectLog{}).Count(&total)

	totalPage = total / int64(size)
	if total%int64(size) > 0 {
		totalPage = totalPage + 1
	}
	c.HTML(200, "moderation.gohtml", OutputCommonSession(i.db, c, gin.H{
		"logs":        logs,
		"totalPage":   totalPage,
		"hasNext":     pageNumber < int(totalPage),
		"hasPrev":     pageNumber > 1,
		"currentPage": pageNumber,
	}))
}

func (i *IndexHandler) SearchByDomain(c *gin.Context) {
	userinfo := GetCurrentUser(c)

	domainName := c.Param("domainName")

	page := c.DefaultQuery("p", "1")

	c.HTML(200, "index.gohtml", OutputCommonSession(i.db, c, gin.H{
		"selected": "history",
	}, QueryPosts(i.db, vo.QueryPostsRequest{
		Userinfo:  userinfo,
		Domain:    domainName,
		OrderType: "",
		Page:      cast.ToInt64(page),
		Size:      25,
	})))
}
