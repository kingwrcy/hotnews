package handler

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"math/rand"
)

func Setup(injector *do.Injector, engine *gin.Engine) {
	provideHandlers(injector)

	userHandler := do.MustInvoke[*UserHandler](injector)
	indexHandler := do.MustInvoke[*IndexHandler](injector)
	postHandler := do.MustInvoke[*PostHandler](injector)

	engine.GET("/", indexHandler.Index)
	engine.GET("/search", indexHandler.ToSearch)
	engine.GET("/new", indexHandler.ToNew)
	engine.GET("/s/:id", indexHandler.ToPost)
	engine.GET("/resetPwd", indexHandler.ToResetPwd)
	engine.GET("/tags", indexHandler.ToTags)

	userGroup := engine.Group("/u")
	userGroup.POST("/login", userHandler.Login)
	userGroup.GET("/login", userHandler.ToLogin)
	userGroup.GET("/logout", userHandler.Logout)
	userGroup.GET("/profile/:id", userHandler.ToProfile)
	userGroup.GET("/invite/:code", userHandler.ToProfile)

	postGroup := engine.Group("/p")
	postGroup.POST("/new", postHandler.Add)
}

func provideHandlers(injector *do.Injector) {
	do.Provide(injector, NewIndexHandler)
	do.Provide(injector, NewUserHandler)
	do.Provide(injector, NewPostHandler)
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

func GetCurrentUserID(c *gin.Context) uint {
	session := sessions.Default(c)
	login := session.Get("login")
	if login != nil {
		userinfo := session.Get("userinfo")
		if v, ok := userinfo.(vo.Userinfo); ok {
			return v.ID
		}
	}
	return 0
}

func OutputCommonSession(c *gin.Context, h gin.H) gin.H {
	session := sessions.Default(c)
	result := gin.H{}

	result["login"] = session.Get("login")
	result["userinfo"] = session.Get("userinfo")
	for k, v := range h {
		result[k] = v
	}
	return result
}
