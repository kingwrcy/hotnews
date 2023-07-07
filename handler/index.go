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
			i.db.Model(&model.TbPost{}).Where("status = 'WAIT_APPROVE'").Count(&waitApproved)
		}
	}

	var posts []model.TbPost
	var total int64
	var totalPage int64
	size := 25
	page := c.DefaultQuery("p", "1")
	pageNumber := cast.ToInt(page)

	subQuery := i.db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'POST' and action ='UP'", userinfo.ID)

	i.db.Table("tb_post p").Select("p.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted").
		Joins("LEFT JOIN (?) AS vote ON p.id = vote.target_id", subQuery).Preload("User").Preload("Tags").
		Where("created_at >= now() - interval 7 day and status = 'Active'").
		Order("point desc,created_at desc").
		Offset((pageNumber - 1) * size).Limit(size).Find(&posts)
	i.db.Model(&model.TbPost{}).Where("created_at >= now() - interval 7 day and status = 'Active'").Count(&total)

	//if userinfo != nil {
	//	for index, p := range posts[:] {
	//		var item model.TbVote
	//		if err := i.db.Model(&model.TbVote{}).Where("post_id = ? and user_id = ?", p.ID, userinfo.ID).Limit(1).Find(&item).Error; err == nil {
	//			if item.Action == "UP" {
	//				posts[index].UpVoted = true
	//			} else if item.Action == "DOWN" {
	//				posts[index].DownVoted = true
	//			}
	//		}
	//	}
	//}
	totalPage = total / int64(size)
	if total%int64(size) > 0 {
		totalPage = totalPage + 1
	}

	c.HTML(200, "index.html", OutputCommonSession(c, gin.H{
		"selected":     "/",
		"waitApproved": waitApproved,
		"posts":        posts,
		"totalPage":    totalPage,
		"hasNext":      int64(pageNumber) < totalPage,
		"hasPrev":      int64(pageNumber) > 1,
		"currentPage":  pageNumber,
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

func (i *IndexHandler) History(c *gin.Context) {
	userinfo := GetCurrentUser(c)

	var posts []model.TbPost
	var total int64
	var totalPage int64
	size := 25
	page := c.DefaultQuery("p", "1")
	pageNumber := cast.ToInt(page)

	i.db.Model(&model.TbPost{}).Where("created_at >= now() - interval 7 day and status = 'Active'").Count(&total)

	if userinfo != nil {
		subQuery := i.db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'POST' and action ='UP'", userinfo.ID)

		i.db.Table("tb_post p").Select("p.*,IF(vote.target_id IS NOT NULL, 1, 0) AS UpVoted").
			Joins("LEFT JOIN (?) AS vote ON p.id = vote.target_id", subQuery).Preload("User").Preload("Tags").
			Where("created_at >= now() - interval 7 day and status = 'Active'").Order("created_at desc").
			Offset((pageNumber - 1) * size).Limit(size).Find(&posts)
	} else {
		i.db.Model(&model.TbPost{}).Preload("User").Preload("Tags").
			Where("created_at >= now() - interval 7 day and status = 'Active'").Order("created_at desc").
			Offset((pageNumber - 1) * size).Limit(size).Find(&posts)
	}

	totalPage = total / int64(size)
	if total%int64(size) > 0 {
		totalPage = totalPage + 1
	}

	c.HTML(200, "index.html", OutputCommonSession(c, gin.H{
		"selected":    "history",
		"posts":       posts,
		"totalPage":   totalPage,
		"hasNext":     int64(pageNumber) < totalPage,
		"hasPrev":     int64(pageNumber) > 1,
		"currentPage": pageNumber,
	}))
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
		subQuery := i.db.Table("tb_vote").Select("target_id").Where("user_id = ? and type = 'COMMENT' and action ='UP'", userinfo.ID)

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
	c.HTML(200, "comments.html", OutputCommonSession(c, gin.H{
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
	var exists int64
	var targetID uint
	if targetType == "POST" {
		var item model.TbPost
		i.db.Model(&model.TbPost{}).Where("pid = ?", id).First(&item)
		targetID = item.ID
	} else if targetType == "COMMENT" {
		var item model.TbComment
		i.db.Model(&model.TbComment{}).Where("cid = ?", id).First(&item)
		targetID = item.ID
	}

	uid := userinfo.ID

	if i.db.Model(&model.TbVote{}).Where("target_id = ? and user_id = ?  and type = ?", targetID, uid, targetType).Count(&exists); exists == 0 {
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
			return nil
		})

	}
	refer := c.GetHeader("Referer")
	if refer == "" {
		refer = "/"
	}

	c.Redirect(302, refer)
}
