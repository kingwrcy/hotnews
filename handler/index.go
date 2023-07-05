package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/samber/do"
	"github.com/spf13/cast"
	"gorm.io/gorm"
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
	var waitApproved int64
	if userinfo != nil {
		if userinfo.Role == "admin" || userinfo.Role == "inspector" {
			i.db.Model(&model.TbPost{}).Where("status = 'WAIT_APPROVE'").Order("created_at desc").Count(&waitApproved)
		}
	}

	var posts []model.TbPost
	var total int64
	page := c.Param("p")
	pageNumber := cast.ToInt(page)
	if pageNumber == 0 {
		pageNumber = 1
	}
	i.db.Model(&model.TbPost{}).Preload("User").Preload("Tags").
		Where("created_at >= now() - interval 7 day and status = 'Active'").
		Offset((pageNumber - 1) * 20).Limit(20).Find(&posts)
	i.db.Model(&model.TbPost{}).Where("created_at >= now() - interval 7 day and status = 'Active'").Count(&total)

	c.HTML(200, "index.html", OutputCommonSession(c, gin.H{
		"selected":     "/",
		"waitApproved": waitApproved,
		"posts":        posts,
		"total":        total,
	}))
}

func (i *IndexHandler) ToSearch(c *gin.Context) {
	c.HTML(200, "search.html", OutputCommonSession(c, gin.H{
		"selected": "search",
	}))
}

func (i *IndexHandler) ToNew(c *gin.Context) {
	var tags []model.TbTag
	i.db.Model(&model.TbTag{}).Find(&tags)
	c.HTML(200, "new.html", OutputCommonSession(c, gin.H{
		"selected": "new",
		"tags":     tags,
	}))
}

func (i *IndexHandler) ToPost(c *gin.Context) {
	c.HTML(200, "post.html", OutputCommonSession(c, gin.H{}))
}
func (i *IndexHandler) ToResetPwd(c *gin.Context) {
	c.HTML(200, "resetPwd.html", OutputCommonSession(c, gin.H{}))
}
func (i *IndexHandler) ToBeInvited(c *gin.Context) {
	c.HTML(200, "toBeInvited.html", OutputCommonSession(c, gin.H{}))
}
func (i *IndexHandler) ToTags(c *gin.Context) {
	var tags []model.TbTag
	i.db.Model(model.TbTag{}).Find(&tags)
	c.HTML(200, "tags.html", OutputCommonSession(c, gin.H{
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

	c.HTML(200, "wait.html", OutputCommonSession(c, gin.H{
		"posts":        waitApprovedList,
		"waitApproved": len(waitApprovedList),
		"selected":     "approve",
	}))
}
