package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
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
	c.HTML(200, "index.html", gin.H{})
}

func (i *IndexHandler) ToSearch(c *gin.Context) {
	c.HTML(200, "search.html", gin.H{})
}

func (i *IndexHandler) ToNew(c *gin.Context) {
	c.HTML(200, "new.html", gin.H{})
}

func (i *IndexHandler) ToPost(c *gin.Context) {
	c.HTML(200, "post.html", gin.H{})
}
func (i *IndexHandler) ToResetPwd(c *gin.Context) {
	c.HTML(200, "resetPwd.html", gin.H{})
}
