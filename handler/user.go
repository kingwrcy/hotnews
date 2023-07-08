package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
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
	err := c.Bind(&request)
	if err != nil {
		log.Printf("error2 is %s", err)
		c.HTML(200, "login.html", gin.H{
			"msg":      "参数错误",
			"selected": "login",
		})
		return
	}
	log.Printf("req is %+v", request)

	var user model.TbUser
	if err := u.db.
		Where("username = ?", request.Username).
		First(&user).Error; err == gorm.ErrRecordNotFound {

		log.Printf("error is %s", err)
		c.HTML(200, "login.html", gin.H{
			"msg":      "登录失败，用户名或者密码不正确 not exists",
			"selected": "login",
		})
		return
	}
	log.Printf("req is %+v", user)
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)) != nil {
		c.HTML(200, "login.html", gin.H{
			"msg":      "登录失败，用户名或者密码不正确",
			"selected": "login",
		})
		return
	}
	if user.Status == "Banned" {
		c.HTML(200, "login.html", gin.H{
			"msg":      "用户已被ban",
			"selected": "login",
		})
		return
	}

	cookieData := vo.Userinfo{
		Username: user.Username,
		Role:     user.Role,
		ID:       user.ID,
		Email:    user.Email,
	}
	c.Redirect(301, "/")
	session := sessions.Default(c)
	session.Set("login", true)
	session.Set("userinfo", cookieData)
	_ = session.Save()
	return
}

func (u *UserHandler) ToLogin(c *gin.Context) {
	c.HTML(200, "login.html", OutputCommonSession(c, gin.H{
		"selected": "login",
	}))
}
func (u *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1})
	session.Save()
	c.Redirect(302, "/")
}
func (u *UserHandler) ToProfile(c *gin.Context) {
	username := c.Param("username")
	var user model.TbUser
	if err := u.db.Preload(clause.Associations).Where("username= ?", username).First(&user).Error; err == gorm.ErrRecordNotFound {
		c.HTML(200, "profile.html", OutputCommonSession(c, gin.H{
			"selected": "mine",
			"msg":      "如果用户确定存在,可能他改名字了.",
		}))
		return
	}
	c.HTML(200, "profile.html", OutputCommonSession(c, gin.H{
		"selected": "mine",
		"user":     user,
	}))
}
func (u *UserHandler) DoInvited(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.Redirect(200, "/")
		return
	}

	var invited model.TbInviteRecord
	var user model.TbUser
	u.db.Where("code = ? and invalidAt >= now() and status = 'ENABLE'", code).First(&invited)
	if &invited == nil {
		c.HTML(200, "toBeInvited.html", gin.H{
			"msg": "邀请码已使用/已过期/无效",
		})
		return
	}
	var request vo.RegisterRequest
	if err := c.Bind(&request); err != nil {
		c.HTML(200, "toBeInvited.html", gin.H{
			"msg": "参数无效",
		})
		return
	}
	user.Username = request.Username
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(200, "toBeInvited.html", gin.H{
			"msg": "系统异常",
		})
		return
	}
	user.Password = string(hashedPwd)
	user.Bio = request.Bio
	user.Email = request.Email
	user.Status = "Active"
	user.CommentCount = 0
	user.PostCount = 0

	var inviteRecords []model.TbInviteRecord
	var count = 0
	for count < 3 {
		count++
		inviteRecords = append(inviteRecords, model.TbInviteRecord{
			Username:  user.Username,
			Code:      RandStringBytesMaskImpr(10),
			InvalidAt: time.Now().Add(3 * 24 * time.Hour),
			Status:    "ENABLE",
		})
	}
	err = u.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&user).Error
		if err != nil {
			return err
		}
		err = tx.Model(&invited).Where("id=?", invited.ID).Updates(model.TbInviteRecord{
			InvitedUsername:  request.Username,
			InvitedUserEmail: request.Email,
			Status:           "DISABLE",
		}).Error
		if err != nil {
			return err
		}
		err = tx.Save(&inviteRecords).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.HTML(200, "toBeInvited.html", gin.H{
			"msg": "系统异常",
		})
		return
	}
	c.HTML(200, "toBeInvited.html", gin.H{
		"msg": "注册成功,去登录吧",
	})
}
