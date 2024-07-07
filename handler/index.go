package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"log"
	"strconv"
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
	topics := QueryPosts(i.db, vo.QueryPostsRequest{
		Userinfo:  userinfo,
		Begin:     &begin,
		OrderType: "index",
		Page:      cast.ToInt64(page),
		Size:      25,
	})
	if list, ok := topics["posts"].([]model.TbPost); ok && len(list) == 0 {
		c.Redirect(301, "/history")
		return
	}

	c.HTML(200, "index.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"selected": "/",
	}, topics))
}

func (i *IndexHandler) ToSearch(c *gin.Context) {
	c.HTML(200, "search.gohtml", OutputCommonSession(i.injector, c, gin.H{
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
	c.HTML(200, "search.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"selected":  "search",
		"condition": request,
	}, QueryPosts(i.db, request)))
}

func (i *IndexHandler) ToNew(c *gin.Context) {
	var tags []model.TbTag
	i.db.Model(&model.TbTag{}).Preload("Parent").Where("parent_id is null").Preload("Children").Find(&tags)
	c.HTML(200, "new.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"selected": "new",
		"tags":     tags,
	}))
}

func (i *IndexHandler) ToPost(c *gin.Context) {
	c.HTML(200, "post.gohtml", OutputCommonSession(i.injector, c, gin.H{}))
}
func (i *IndexHandler) ToResetPwd(c *gin.Context) {
	c.HTML(200, "resetPwd.gohtml", OutputCommonSession(i.injector, c, gin.H{}))
}

func (i *IndexHandler) ToAddTag(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.Redirect(302, "/tags")
		return
	}
	var parentTags []model.TbTag
	i.db.Find(&parentTags, "parent_id is null")

	c.HTML(200, "tagEdit.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"parents":  parentTags,
		"selected": "tags",
	}))
}
func (i *IndexHandler) ToEditTag(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.Redirect(302, "/tags")
		return
	}
	id := cast.ToString(c.Param("id"))
	var tag model.TbTag
	i.db.First(&tag, "id = ?", id)

	var parentTags []model.TbTag
	i.db.Find(&parentTags, "parent_id is null")

	c.HTML(200, "tagEdit.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"tag":      tag,
		"parentID": cast.ToInt(tag.ParentID),
		"parents":  parentTags,
		"selected": "tags",
	}))
}

func (i *IndexHandler) SaveTag(c *gin.Context) {
	var request vo.EditTagVo

	if err := c.Bind(&request); err != nil {
		c.Redirect(302, "/tags")
		return
	}
	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.Redirect(302, "/tags")
		return
	}
	showInHot := "Y"
	showInAll := "Y"
	var pid *uint
	if request.ShowInHot != "on" {
		showInHot = "N"
	}
	if request.ShowInAll != "on" {
		showInAll = "N"
	}
	if cast.ToInt(request.ParentID) > 0 {
		id := cast.ToUint(request.ParentID)
		pid = &id
	} else {
		pid = nil
	}
	log.Printf("request.ParentID is %+v", cast.ToInt(request.ParentID))
	if cast.ToInt(request.ID) == 0 {
		i.db.Save(&model.TbTag{
			Name:      request.Name,
			Desc:      request.Desc,
			ParentID:  pid,
			CssClass:  request.CssClass,
			ShowInHot: showInHot,
			ShowInAll: showInAll,
		})
	} else {
		i.db.Model(&model.TbTag{}).Where("id = ?", request.ID).
			Updates(map[string]interface{}{
				"name":        request.Name,
				"desc":        request.Desc,
				"parent_id":   pid,
				"css_class":   request.CssClass,
				"show_in_hot": showInHot,
				"show_in_all": showInAll,
			})
	}

	c.Redirect(302, "/tags")
}

func (i *IndexHandler) AddTag(c *gin.Context) {
	var request vo.EditTagVo
	if err := c.Bind(&request); err != nil {
		c.JSON(403, nil)
		return
	}
	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.JSON(403, nil)
		return
	}
	var tag model.TbTag
	tag.Name = request.Name
	tag.Desc = request.Desc
	if request.ParentID != nil {
		tag.Parent = &model.TbTag{
			Model: gorm.Model{
				ID: *request.ParentID,
			},
		}
	}
	i.db.Create(&tag)
	c.JSON(200, nil)
}

func (i *IndexHandler) ToTags(c *gin.Context) {
	var tags []model.TbTag

	i.db.Model(&model.TbTag{}).Where("parent_id is null").Preload("Children").Find(&tags)
	c.HTML(200, "tags.gohtml", OutputCommonSession(i.injector, c, gin.H{
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

	c.HTML(200, "wait.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"posts":        waitApprovedList,
		"waitApproved": len(waitApprovedList),
		"selected":     "approve",
	}))
}

func (i *IndexHandler) History(c *gin.Context) {
	userinfo := GetCurrentUser(c)

	page := c.DefaultQuery("p", "1")

	c.HTML(200, "index.gohtml", OutputCommonSession(i.injector, c, gin.H{
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

		i.db.Table("tb_comment c").Select("c.*,CASE WHEN vote.target_id IS NOT NULL THEN 1 ELSE 0  END AS UpVoted").Joins("LEFT JOIN (?) AS vote ON c.id = vote.target_id", subQuery).Preload("Post").
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
	c.HTML(200, "comments.gohtml", OutputCommonSession(i.injector, c, gin.H{
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

	if i.db.Model(&model.TbVote{}).Where("target_id = ? and tb_user_id = ?  and type = ?", targetID, uid, targetType).Count(&exists); exists == 0 {
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
				if err := tx.Model(&model.TbPost{}).Where("id =?", targetID).Update(col, gorm.Expr(fmt.Sprintf("\"%s\"", col)+"+1")).Error; err != nil {
					return err
				}
			} else if targetType == "COMMENT" {
				if err := tx.Model(&model.TbComment{}).Where("id =?", targetID).Update(col, gorm.Expr(fmt.Sprintf("\"%s\"", col)+"+1")).Error; err != nil {
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
	c.HTML(200, "moderation.gohtml", OutputCommonSession(i.injector, c, gin.H{
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

	c.HTML(200, "index.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"selected": "history",
	}, QueryPosts(i.db, vo.QueryPostsRequest{
		Userinfo:  userinfo,
		Domain:    domainName,
		OrderType: "",
		Page:      cast.ToInt64(page),
		Size:      25,
	})))
}

func (i *IndexHandler) ToSettings(c *gin.Context) {

	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.Redirect(302, "/")
		return
	}

	var settings model.TbSettings
	var saveSettingsRequest vo.SaveSettingsRequest
	if errors.Is(i.db.First(&settings).Error, gorm.ErrRecordNotFound) {
		saveSettingsRequest.RegMode = "hotnews"
		c.HTML(200, "settings.gohtml", OutputCommonSession(i.injector, c, gin.H{
			"selected": "settings",
		}))
		return
	}

	c.HTML(200, "settings.gohtml", OutputCommonSession(i.injector, c, gin.H{
		"selected": "settings",
	}))
}

func (i *IndexHandler) SaveSettings(c *gin.Context) {
	var request vo.SaveSettingsRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(403, nil)
		return
	}
	var settings model.TbSettings
	i.db.First(&settings)

	settings.Content = model.SaveSettingsRequest(request)
	i.db.Save(&settings)

	c.Redirect(302, "/settings")

}

func (i *IndexHandler) RemoveTag(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.Redirect(302, "/tags")
		return
	}
	tagId, _ := strconv.Atoi(c.Param("tagId"))
	var tag model.TbTag
	i.db.Preload("Posts").First(&tag, "id = ?", tagId)
	i.db.Delete(&tag.Posts)
	i.db.Delete(&tag)
	c.Redirect(302, "/tags")
}
