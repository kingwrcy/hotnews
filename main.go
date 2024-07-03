package main

import (
	"embed"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kingwrcy/hn/handler"
	"github.com/kingwrcy/hn/middleware"
	"github.com/kingwrcy/hn/model"
	"github.com/kingwrcy/hn/provider"
	"github.com/kingwrcy/hn/task"
	"github.com/kingwrcy/hn/vo"
	"github.com/samber/do"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	} else if duration < 24*time.Hour*365 {
		return fmt.Sprintf("%d天前", duration/(24*time.Hour))
	} else {
		return fmt.Sprintf("%d年前", duration/(24*time.Hour*365))
	}
}

//go:embed templates
var templatesFS embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	injector := do.New()

	do.Provide(injector, provider.NewAppConfig)
	do.Provide(injector, provider.NewRepository)

	db := do.MustInvoke[*gorm.DB](injector)
	config := do.MustInvoke[*provider.AppConfig](injector)
	err = db.AutoMigrate(&model.TbMessage{},
		&model.TbUser{}, &model.TbInviteRecord{},
		&model.TbPost{}, &model.TbInspectLog{},
		&model.TbComment{}, &model.TbTag{}, &model.TbStatistics{},
		&model.TbVote{})
	if err != nil {
		log.Printf("升级数据库异常,启动失败.%s", err)
		return
	}
	gob.Register(vo.Userinfo{})
	engine := gin.Default()

	//store, _ := redis.NewStore(10, "tcp", config.RedisAddress, "", []byte(config.CookieSecret))
	store := cookie.NewStore([]byte(config.CookieSecret))

	engine.Use(sessions.Sessions("c", store))
	engine.Use(middleware.CostHandler())

	if os.Getenv("GIN_MODE") == "release" {
		ts, _ := fs.Sub(templatesFS, "templates")
		engine.HTMLRender = loadTemplates(ts)
		//s, _ := fs.Sub(staticFS, "static")
		//engine.StaticFS("/static", http.FS(s))
	} else {
		engine.HTMLRender = loadLocalTemplates("./templates")
		//engine.Static("/static", "./static")
	}

	handler.Setup(injector, engine)

	go task.StartPostTask(injector)

	log.Printf("启动http服务,端口:%d,监听请求中...", config.Port)
	engine.Run(fmt.Sprintf(":%d", config.Port))
}

func templateFun() template.FuncMap {
	return template.FuncMap{
		"StringsJoin": strings.Join,
		"timeAgo":     timeAgo,
		"unEscapeHTML": func(content string) template.HTML {
			return template.HTML(content)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"dateFormat": func(date time.Time, format string) string {
			return date.Format(format)
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}

func loadLocalTemplates(templatesDir string) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	layouts, err := filepath.Glob(templatesDir + "/layouts/*.gohtml")
	if err != nil {
		panic(err.Error())
	}
	includes, err := filepath.Glob(templatesDir + "/includes/*.gohtml")
	if err != nil {
		panic(err.Error())
	}

	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)

		r.AddFromFilesFuncs(filepath.Base(include), templateFun(), files...)
	}
	return r
}

func loadTemplates(templatesDir fs.FS) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	layouts, err := fs.Glob(templatesDir, "layouts/*.gohtml")
	if err != nil {
		panic(err.Error())
	}
	includes, err := fs.Glob(templatesDir, "includes/*.gohtml")
	if err != nil {
		panic(err.Error())
	}

	// Generate our templates map from our layouts/ and includes/ directories
	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		templateContents := make([]string, len(files))

		for _, f := range files {
			open, err := templatesDir.Open(f)
			if err != nil {
				panic(err)
			}
			buffer, err := io.ReadAll(open)
			if err != nil {
				panic(err)
			}
			templateContents = append(templateContents, string(buffer))
			open.Close()
		}
		r.AddFromStringsFuncs(filepath.Base(include), templateFun(), templateContents...)
	}
	return r
}
