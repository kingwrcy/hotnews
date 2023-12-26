package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"gorm.io/gorm"
	"math/rand"
	"time"
)

func Setup(injector *do.Injector, engine *gin.Engine) {
	provideHandlers(injector)

	userHandler := do.MustInvoke[*UserHandler](injector)
	indexHandler := do.MustInvoke[*IndexHandler](injector)
	postHandler := do.MustInvoke[*PostHandler](injector)
	inspectHandler := do.MustInvoke[*InspectHandler](injector)
	_ = do.MustInvoke[*CommentHandler](injector)

	engine.GET("/", indexHandler.Index)
	engine.GET("/history", indexHandler.History)
	engine.GET("/search", indexHandler.ToSearch)
	engine.GET("/new", indexHandler.ToNew)
	engine.GET("/s/:pid", indexHandler.ToPost)
	engine.GET("/resetPwd", indexHandler.ToResetPwd)
	engine.GET("/tags", indexHandler.ToTags)
	//engine.GET("/wait", indexHandler.ToWaitApproved)
	engine.GET("/comments", indexHandler.ToComments)
	engine.GET("/vote", indexHandler.Vote)
	engine.GET("/moderations", indexHandler.Moderation)
	engine.GET("/d/:domainName", indexHandler.SearchByDomain)
	engine.POST("/search", indexHandler.DoSearch)
	engine.GET("/invite/:code", userHandler.ToInvited)
	engine.POST("/invite/:code", userHandler.DoInvited)
	engine.GET("/about", userHandler.ToAbout)
	engine.GET("/type/:type", postHandler.SearchByType)

	engine.POST("/inspect", inspectHandler.Inspect)

	userGroup := engine.Group("/u")
	userGroup.POST("/login", userHandler.Login)
	userGroup.GET("/login", userHandler.ToLogin)
	userGroup.GET("/logout", userHandler.Logout)
	userGroup.GET("/profile/:username", userHandler.Links)
	userGroup.GET("/profile/:username/asks", userHandler.Asks)
	userGroup.GET("/profile/:username/links", userHandler.Links)
	userGroup.GET("/profile/:username/comments", userHandler.Comments)
	userGroup.GET("/message/setAllRead", userHandler.SetAllRead)
	userGroup.GET("/message/setSingleRead", userHandler.SetSingleRead)
	userGroup.GET("/message", userHandler.ToMessage)

	//commentGroup := engine.Group("/c")
	//commentGroup.GET("/vote", commentHandler.Vote)

	postGroup := engine.Group("/p")
	postGroup.POST("/new", postHandler.Add)
	postGroup.GET("/:pid", postHandler.Detail)
	postGroup.GET("/:pid/edit", postHandler.ToEdit)
	postGroup.POST("/:pid/edit", postHandler.DoUpdate)
	postGroup.POST("/comment", postHandler.AddComment)

	tagGroup := engine.Group("/t")
	tagGroup.GET("/:tag", postHandler.SearchByTag)
	tagGroup.GET("/p/:tag", postHandler.SearchByParentTag)

}

func provideHandlers(injector *do.Injector) {
	do.Provide(injector, NewIndexHandler)
	do.Provide(injector, NewUserHandler)
	do.Provide(injector, NewPostHandler)
	do.Provide(injector, newInspectHandler)
	do.Provide(injector, newCommentHandler)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func GetCurrentUser(c *gin.Context) *vo.Userinfo {
	session := sessions.Default(c)
	login := session.Get("login")
	if login != nil {
		userinfo := session.Get("userinfo")
		if v, ok := userinfo.(vo.Userinfo); ok {
			return &v
		}
	}
	return nil
}

func OutputCommonSession(db *gorm.DB, c *gin.Context, h ...gin.H) gin.H {
	session := sessions.Default(c)
	result := gin.H{}
	start := c.GetInt64("executionTime")
	result["executionTime"] = time.Since(time.UnixMilli(start)).Milliseconds()
	result["login"] = session.Get("login")
	result["userinfo"] = session.Get("userinfo")
	for _, v := range h {
		for k1, v1 := range v {
			result[k1] = v1
		}
	}

	var total int64
	userinfo := GetCurrentUser(c)
	if userinfo != nil {
		db.Model(&model.TbMessage{}).Where("to_user_id = ? and `read` = 'N'", userinfo.ID).Count(&total)
		result["unReadMessageCount"] = total
	}
	return result
}
