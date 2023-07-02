package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type UserHandler struct {
	injector *do.Injector
	db       *gorm.DB
}

func NewUserHandler(injector *do.Injector) (*UserHandler, error) {
	return &UserHandler{
		injector: injector,
		db:       do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (u *UserHandler) Login(c *gin.Context) {
	var request vo.LoginRequest
	err := handleParamError(c, &request)
	if err != nil {
		return
	}
	var user model.TbUser
	if u.db.Model(model.TableNameTbUser).
		Where("username = ? and password = ?", request.Username, request.Password).
		First(&user); &user != nil {
		c.Redirect(200, "/")
		return
	}
	c.HTML(200, "/login", gin.H{
		"error": "登录失败，用户名或者密码不正确",
	})
}

func (u *UserHandler) ToLogin(c *gin.Context) {
	c.HTML(200, "login.html", gin.H{})
}
func (u *UserHandler) ToProfile(c *gin.Context) {
	c.HTML(200, "profile.html", gin.H{})
}
