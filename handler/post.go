package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"gorm.io/gorm"
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

func (p PostHandler) Add(c *gin.Context) {
	uid := GetCurrentUserID(c)
	if uid == 0 {
		c.Redirect(302, "/login")
		return
	}

	var request vo.NewPostRequest
	if err := c.Bind(&request); err != nil {
		c.HTML(200, "new.html", gin.H{
			"msg": "参数异常",
		})
		return
	}
	if len(request.TagIDs) == 0 || len(request.TagIDs) > 5 {
		c.HTML(200, "new.html", gin.H{
			"msg": "标签最少1个,最多5个",
		})
		return
	}
	if request.Type == "" {
		c.HTML(200, "new.html", gin.H{
			"msg": "类型必填",
		})
		return
	}

	var tags []model.TbTag
	for _, v := range request.TagIDs {
		tags = append(tags, model.TbTag{
			Model: gorm.Model{ID: v},
		})
	}
	post := model.TbPost{
		Title:    request.Title,
		Link:     request.Link,
		Status:   "WAIT_APPROVE",
		Content:  request.Content,
		UpVote:   0,
		DownVote: 0,
		Type:     request.Type,
		Tags:     tags,
		User:     model.TbUser{Model: gorm.Model{ID: uid}},
	}

	err := p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(post).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.TbUser{Model: gorm.Model{ID: uid}}).Update("postCount", "postCount+1").Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.HTML(200, "new.html", gin.H{
			"msg": "系统错误",
		})
		return
	}
	c.HTML(200, "new.html", gin.H{
		"msg": "提交成功,等待审核",
	})
	return
}
