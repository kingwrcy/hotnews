package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"gorm.io/gorm"
	"log"
)

type InspectHandler struct {
	injector *do.Injector
	db       *gorm.DB
}

func newInspectHandler(injector *do.Injector) (*InspectHandler, error) {
	return &InspectHandler{
		injector: injector,
		db:       do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (p InspectHandler) Inspect(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil || (userinfo.Role != "admin" && userinfo.Role != "inspector") {
		c.JSON(200, gin.H{
			"msg": "非法访问",
		})
		return
	}

	uid := userinfo.ID

	var request vo.InspectRequest
	var inspectLog model.TbInspectLog
	if err := c.Bind(&request); err != nil {
		c.JSON(200, gin.H{
			"msg": "参数错误",
		})
		return
	}
	log.Printf("%+v", request)

	if request.PostID == 0 && request.CommentID == 0 {
		c.JSON(200, gin.H{
			"msg": "参数错误",
		})
		return
	}
	status := "Active"

	inspectLog.InspectType = request.InspectType
	inspectLog.PostID = request.PostID
	inspectLog.Reason = request.Reason
	inspectLog.Result = request.Result
	if request.Result == "reject" {
		inspectLog.Action = "deleted " + request.InspectType
		status = "Rejected"
	}

	inspectLog.InspectorID = uid
	if request.PostID > 0 {
		var post model.TbPost
		if err := p.db.Model(&model.TbPost{}).Where("id = ?", request.PostID).First(&post).Error; err == nil {
			inspectLog.Title = "链接:" + post.Title
		}
	}

	err := p.db.Transaction(func(tx *gorm.DB) error {
		if request.Result == "reject" {
			err := tx.Save(&inspectLog).Error
			if err != nil {
				return err
			}
		}

		if request.PostID > 0 {
			err := tx.Model(model.TbPost{}).Where("id = ?", request.PostID).Update("status", status).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(200, gin.H{
			"msg": err.Error(),
		})
		return
	}
	c.Redirect(302, "/wait")
	return
}
