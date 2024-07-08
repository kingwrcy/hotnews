package handler

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"github.com/spf13/cast"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/mail"
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
		c.HTML(200, "login.gohtml", gin.H{
			"msg":      "参数错误",
			"selected": "login",
		})
		return
	}
	var user model.TbUser
	if err := u.db.
		Where("username = ?", request.Username).
		First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {

		c.HTML(200, "login.gohtml", gin.H{
			"msg":      "登录失败，用户名或者密码不正确",
			"selected": "login",
		})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)) != nil {
		c.HTML(200, "login.gohtml", gin.H{
			"msg":      "登录失败，用户名或者密码不正确",
			"selected": "login",
		})
		return
	}
	if user.Status == "Banned" {
		c.HTML(200, "login.gohtml", gin.H{
			"msg":      "用户已被ban",
			"selected": "login",
		})
		return
	}

	cookieData := vo.Userinfo{
		Username:  user.Username,
		Role:      user.Role,
		ID:        user.ID,
		Email:     user.Email,
		EmailHash: user.EmailHash,
	}
	c.Redirect(301, "/")
	session := sessions.Default(c)
	session.Set("login", true)
	session.Set("userinfo", cookieData)
	_ = session.Save()
	return
}

func (u *UserHandler) ToLogin(c *gin.Context) {
	var settings model.TbSettings
	u.db.First(&settings)
	c.HTML(200, "login.gohtml", OutputCommonSession(u.injector, c, gin.H{
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

func (u *UserHandler) Asks(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	username := c.Param("username")
	p := c.DefaultQuery("p", "1")
	page := cast.ToInt(p)
	size := 10

	var user model.TbUser
	if err := u.db.Preload(clause.Associations).Where("username= ?", username).First(&user).Error; err == gorm.ErrRecordNotFound {
		c.HTML(200, "profile.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"selected": "mine",
			"msg":      "如果用户确定存在,可能他改名字了.",
		}))
		return
	}

	var inviteRecords []model.TbInviteRecord
	if userinfo != nil && userinfo.ID == user.ID {
		u.db.Model(&model.TbInviteRecord{}).Where("username = ?", user.Username).Find(&inviteRecords)
	}
	var invitedUsername string
	u.db.Model(&model.TbInviteRecord{}).Select("username").Where("invitedUsername = ?", user.Username).First(&invitedUsername)

	var total int64
	var posts []model.TbPost

	tx := u.db.Model(&model.TbPost{}).Preload(clause.Associations).
		Where("user_id = ? and status ='Active' and type = 'ask'", user.ID)
	tx.Count(&total)
	tx.Order("created_at desc").Offset((cast.ToInt(page) - 1) * size).Limit(size).
		Find(&posts)
	totalPage := total / cast.ToInt64(size)

	if total%cast.ToInt64(size) > 0 {
		totalPage = totalPage + 1
	}

	c.HTML(200, "profile.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"selected":        "mine",
		"user":            user,
		"sub":             "ask",
		"posts":           posts,
		"inviteRecords":   inviteRecords,
		"invitedUsername": invitedUsername,
		"totalPage":       totalPage,
		"total":           total,
		"hasNext":         cast.ToInt64(page) < totalPage,
		"hasPrev":         page > 1,
		"currentPage":     cast.ToInt(page),
	}))
}

func (u *UserHandler) Links(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	p := c.DefaultQuery("p", "1")
	page := cast.ToInt(p)
	size := 10

	username := c.Param("username")
	var user model.TbUser
	if err := u.db.Preload(clause.Associations).Where("username= ?", username).First(&user).Error; err == gorm.ErrRecordNotFound {
		c.HTML(200, "profile.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"selected": "mine",
			"msg":      "如果用户确定存在,可能他改名字了.",
		}))
		return
	}

	var inviteRecords []model.TbInviteRecord
	if userinfo != nil && userinfo.ID == user.ID {
		u.db.Model(&model.TbInviteRecord{}).Where("username = ?", user.Username).Scan(&inviteRecords)
	}
	var invitedUsername string
	u.db.Model(&model.TbInviteRecord{}).Select("username").Where("\"invitedUsername\" = ?", user.Username).First(&invitedUsername)

	var total int64
	var posts []model.TbPost
	tx := u.db.Model(&model.TbPost{}).Preload(clause.Associations).
		Where("user_id = ? and status ='Active' and type = 'link'", user.ID)

	tx.Count(&total)
	tx.Order("created_at desc").Offset((cast.ToInt(page) - 1) * size).Limit(size).
		Find(&posts)

	totalPage := total / (cast.ToInt64(size))

	if total%(cast.ToInt64(size)) > 0 {
		totalPage = totalPage + 1
	}

	c.HTML(200, "profile.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"selected":        "mine",
		"user":            user,
		"sub":             "link",
		"posts":           posts,
		"inviteRecords":   inviteRecords,
		"invitedUsername": invitedUsername,
		"totalPage":       totalPage,
		"total":           total,
		"hasNext":         cast.ToInt64(page) < totalPage,
		"hasPrev":         page > 1,
		"currentPage":     cast.ToInt(page),
	}))
}

func (u *UserHandler) ToMessage(c *gin.Context) {

	var messages []model.TbMessage
	var total int64
	userinfo := GetCurrentUser(c)
	page := cast.ToInt(c.DefaultQuery("p", "1"))
	size := 25

	u.db.Where("to_user_id = ?", userinfo.ID).Count(&total)
	u.db.Where("to_user_id = ?", userinfo.ID).Limit(size).Offset((page - 1) * size).
		Order("created_at desc").Find(&messages)

	c.HTML(200, "message.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"selected": "message",
		"messages": messages,
		"total":    total,
	}))
}

func (u *UserHandler) SetAllRead(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil {
		c.Redirect(302, "/u/login")
		return
	}
	u.db.Model(&model.TbMessage{}).Where("to_user_id = ? and read = 'N'", userinfo.ID).Update("read", "Y")
	u.ToMessage(c)
}

func (u *UserHandler) SetSingleRead(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil {
		c.Redirect(302, "/u/login")
		return
	}

	if id, ok := c.GetQuery("id"); ok {
		log.Printf("get id %+v", id)
		u.db.Model(&model.TbMessage{}).Where("id = ? and to_user_id = ? and read = 'N'", id, userinfo.ID).Update("read", "Y")
	}
	u.ToMessage(c)
}

func (u *UserHandler) Comments(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	p := c.DefaultQuery("p", "1")
	page := cast.ToInt(p)
	size := 10

	username := c.Param("username")
	var user model.TbUser
	if err := u.db.Preload(clause.Associations).Where("username= ?", username).First(&user).Error; err == gorm.ErrRecordNotFound {
		c.HTML(200, "profile.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"selected": "mine",
			"msg":      "如果用户确定存在,可能他改名字了.",
		}))
		return
	}

	var inviteRecords []model.TbInviteRecord
	if userinfo != nil && userinfo.ID == user.ID {
		u.db.Model(&model.TbInviteRecord{}).Where("username = ?", user.Username).Find(&inviteRecords)
	}

	var invitedUsername string
	var total int64

	u.db.Model(&model.TbInviteRecord{}).Select("username").Where("\"invitedUsername\" = ?", user.Username).First(&invitedUsername)
	var comments []model.TbComment
	tx := u.db.Model(&model.TbComment{}).
		Preload("User").
		Where("user_id = ? ", user.ID)
	tx.Count(&total)
	tx.Order("created_at desc").Offset((cast.ToInt(page) - 1) * size).Limit(size).
		Find(&comments)

	totalPage := total / (cast.ToInt64(size))

	if total%(cast.ToInt64(size)) > 0 {
		totalPage = totalPage + 1
	}

	c.HTML(200, "profile.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"selected":        "mine",
		"user":            user,
		"sub":             "comments",
		"comments":        comments,
		"inviteRecords":   inviteRecords,
		"invitedUsername": invitedUsername,
		"totalPage":       totalPage,
		"total":           total,
		"hasNext":         cast.ToInt64(page) < totalPage,
		"hasPrev":         page > 1,
		"currentPage":     cast.ToInt(page),
	}))
}

func (u *UserHandler) ToInvited(c *gin.Context) {
	var settings model.TbSettings

	u.db.First(&settings)

	if settings.Content.RegMode == "shutdown" {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"selected": "/",
		}))
		return
	}

	code := c.Param("code")
	if code == "" {
		c.Redirect(200, "/")
		return
	}
	if settings.Content.RegMode == "hotnews" {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"selected": "/",
			"code":     code,
		}))
		return
	}
	var invited model.TbInviteRecord
	err := u.db.Where("code = ? and \"invalidAt\" >= now() and status = 'ENABLE'", code).First(&invited).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"codeIsInvalid": true,
			"msg":           "邀请码已使用/已过期/无效",
		}))
		return
	}

	c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"selected": "/",
		"invited":  invited,
		"code":     code,
	}))
}

func (u *UserHandler) ToAbout(c *gin.Context) {
	c.HTML(200, "about.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"selected": "/",
	}))
}

func (u *UserHandler) DoInvited(c *gin.Context) {
	var settings model.TbSettings
	u.db.First(&settings)

	if settings.Content.RegMode == "shutdown" {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"codeIsInvalid": true,
			"msg":           "目前不开放注册",
		}))
		return
	}

	code := c.Param("code")
	if code == "" {
		c.Redirect(200, "/")
		return
	}

	var invited model.TbInviteRecord
	var user model.TbUser
	if settings.Content.RegMode == "invite" {
		err := u.db.Where("code = ? and \"invalidAt\" >= now() and status = 'ENABLE'", code).First(&invited).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
				"codeIsInvalid": true,
				"msg":           "邀请码已使用/已过期/无效",
			}))
			return
		}
	}

	var request vo.RegisterRequest
	if err := c.Bind(&request); err != nil {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg": "参数无效", "code": code,
		}))
		return
	}
	if len(request.Username) < 3 {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg": "用户名长度必须大于3位", "code": code,
		}))
		return
	}
	if len(request.Password) < 5 {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg": "密码长度必须大于5位", "code": code,
		}))
		return
	}
	if request.Password != request.RepeatPassword {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg": "两次密码不一致", "code": code,
		}))
		return
	}
	if _, ok := mail.ParseAddress(request.Email); ok != nil {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg": "邮箱格式不正确", "code": code,
		}))
		return
	}
	user.Username = request.Username
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg": "系统异常", "code": code,
		}))
		return
	}
	user.Password = string(hashedPwd)
	user.Bio = request.Bio
	user.Email = request.Email
	user.Status = "Active"
	user.CommentCount = 0
	user.PostCount = 0

	hash := sha256.New()
	hash.Write([]byte(user.Email))
	user.EmailHash = fmt.Sprintf("%x", hash.Sum(nil))

	var totalUsers int64
	u.db.Table("tb_user").Where("id <> 999999999").Count(&totalUsers)
	if totalUsers == 0 {
		user.Role = "admin"
	}

	var inviteRecords []model.TbInviteRecord
	var count = 0
	for count < 3 {
		count++
		inviteRecords = append(inviteRecords, model.TbInviteRecord{
			Username:  user.Username,
			Code:      RandStringBytesMaskImpr(10),
			InvalidAt: time.Now().Add(30 * 24 * time.Hour),
			Status:    "ENABLE",
		})
	}
	err = u.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&user).Error
		if err != nil {
			return err
		}
		if settings.Content.RegMode == "invite" {
			err = tx.Model(&invited).Where("id=?", invited.ID).Updates(model.TbInviteRecord{
				InvitedUsername:  request.Username,
				InvitedUserEmail: request.Email,
				Status:           "DISABLE",
			}).Error
			if err != nil {
				return err
			}
		}
		return tx.Save(&inviteRecords).Error
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg":  "用户名已经存在了,换一个吧",
			"code": code,
		}))
		return
	} else if err != nil {
		c.HTML(200, "toBeInvited.gohtml", OutputCommonSession(u.injector, c, gin.H{
			"msg": "系统异常",
		}))
		return
	}
	c.HTML(200, "login.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"msg": "注册成功,去登录吧",
	}))
}

func (u *UserHandler) ToList(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.Redirect(302, "/")
		return
	}
	var users []model.TbUser
	u.db.Where("ID <> 999999999").Order("id desc").Find(&users)
	c.HTML(200, "users.gohtml", OutputCommonSession(u.injector, c, gin.H{
		"selected": "users",
		"users":    users,
	}))

}
