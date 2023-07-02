package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func Setup(injector *do.Injector, engine *gin.Engine) {
	provideHandlers(injector)

	userHandler := do.MustInvoke[*UserHandler](injector)
	indexHandler := do.MustInvoke[*IndexHandler](injector)

	engine.GET("/", indexHandler.Index)
	engine.GET("/search", indexHandler.ToSearch)
	engine.GET("/new", indexHandler.ToNew)
	engine.GET("/s/:id", indexHandler.ToPost)
	engine.GET("/resetPwd", indexHandler.ToResetPwd)

	userGroup := engine.Group("/u")
	userGroup.POST("/login", userHandler.Login)
	userGroup.GET("/login", userHandler.ToLogin)
	userGroup.GET("/profile/:id", userHandler.ToProfile)

}

//
//// 定义一个全局翻译器T
//var trans ut.Translator
//
//// InitTrans 初始化翻译器
//func InitTrans(locale string) (err error) {
//	// 修改gin框架中的Validator引擎属性，实现自定制
//	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
//
//		zhT := zh.New() // 中文翻译器
//		enT := en.New() // 英文翻译器
//
//		// 第一个参数是备用（fallback）的语言环境
//		// 后面的参数是应该支持的语言环境（支持多个）
//		// uni := ut.New(zhT, zhT) 也是可以的
//		uni := ut.New(enT, zhT, enT)
//
//		// locale 通常取决于 http 请求头的 'Accept-Language'
//		var ok bool
//		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
//		trans, ok = uni.GetTranslator(locale)
//		if !ok {
//			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
//		}
//
//		// 注册翻译器
//		switch locale {
//		case "en":
//			err = enTranslations.RegisterDefaultTranslations(v, trans)
//		case "zh":
//			err = zhTranslations.RegisterDefaultTranslations(v, trans)
//		default:
//			err = enTranslations.RegisterDefaultTranslations(v, trans)
//		}
//		return
//	}
//	return
//}

func provideHandlers(injector *do.Injector) {
	do.Provide(injector, NewIndexHandler)
	do.Provide(injector, NewUserHandler)
}

func handleParamError(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(&obj); err != nil {
		c.HTML(200, "error", gin.H{
			"error": err.Error(),
		})
		return err
	}
	return nil
}

func success(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"code": 0,
		"data": data,
	})
}

func fail(c *gin.Context, msg string) {
	failWithCode(c, 2, msg)
}
func failWithCode(c *gin.Context, code int, msg string) {
	c.JSON(200, gin.H{
		"code": code,
		"msg":  msg,
	})
	c.Abort()
}
