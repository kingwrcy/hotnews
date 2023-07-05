package main

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/handler"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/provider"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"html/template"
	"path/filepath"
	"time"

	"gorm.io/gorm"
	"log"
)

func timeAgo(target time.Time) string {
	duration := time.Now().Sub(target)
	if duration < time.Second {
		return "刚刚"
	} else if duration < time.Minute {
		return fmt.Sprintf("%d秒前", duration/time.Second)
	} else if duration < time.Hour {
		return fmt.Sprintf("%d分钟前", duration/time.Minute)
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d小时前", duration/time.Hour)
	} else {
		return fmt.Sprintf("%d天前", duration/(24*time.Hour))
	}
}

func main() {
	injector := do.New()

	do.Provide(injector, provider.NewAppConfig)
	do.Provide(injector, provider.NewRepository)

	db := do.MustInvoke[*gorm.DB](injector)
	config := do.MustInvoke[*provider.AppConfig](injector)
	err := db.AutoMigrate(&model.TbUser{}, &model.TbInviteRecord{}, &model.TbPost{}, &model.TbInspectLog{}, &model.TbComment{})
	if err != nil {
		log.Printf("升级数据库异常,启动失败.%s", err)
		return
	}
	gob.Register(vo.Userinfo{})
	engine := gin.Default()
	store := cookie.NewStore([]byte(config.CookieSecret))

	engine.Use(sessions.Sessions("c", store))
	engine.HTMLRender = loadTemplates("./templates")
	engine.Static("/static", "./static")

	handler.Setup(injector, engine)

	log.Printf("启动http服务,端口:%d,监听请求中...", config.Port)
	engine.Run(fmt.Sprintf(":%d", config.Port))
}

func loadTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	layouts, err := filepath.Glob(templatesDir + "/layouts/*.html")
	if err != nil {
		panic(err.Error())
	}

	includes, err := filepath.Glob(templatesDir + "/includes/*.html")
	if err != nil {
		panic(err.Error())
	}

	// Generate our templates map from our layouts/ and includes/ directories
	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		r.AddFromFilesFuncs(filepath.Base(include), template.FuncMap{
			"timeAgo": timeAgo,
		}, files...)
	}
	return r
}
