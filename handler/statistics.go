package handler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/mileusna/useragent"
	"github.com/samber/do"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"html/template"
	"strings"
	"time"
)

type StatisticsHandler struct {
	injector *do.Injector
	db       *gorm.DB
}

func NewStatisticsHandler(injector *do.Injector) (*StatisticsHandler, error) {
	return &StatisticsHandler{
		injector: injector,
		db:       do.MustInvoke[*gorm.DB](injector),
	}, nil
}

func (s *StatisticsHandler) Query(c *gin.Context) {
	userinfo := GetCurrentUser(c)
	if userinfo == nil || userinfo.Role != "admin" {
		c.Redirect(302, "/")
		return
	}
	start, startExist := c.GetQuery("start")
	if !startExist {
		start = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}
	end, endExist := c.GetQuery("end")
	if !endExist {
		end = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}
	var countMapList []map[string]interface{}
	var countryMapList []map[string]interface{}
	var referMapList []map[string]interface{}
	s.db.Model(&model.TbStatistics{}).Select(" COUNT(1) as total, to_char(created_at, 'YYYY-MM-DD') as day").
		Where("created_at between ? and ?", start, end).
		Group("day").Order("day asc").Scan(&countMapList)
	s.db.Model(&model.TbStatistics{}).Select("country as name,count(1) as value").
		Where("created_at between ? and ?", start, end).Group("name").
		Order("name").Scan(&countryMapList)
	s.db.Model(&model.TbStatistics{}).Select("SUBSTRING(refer FROM '(^[a-z]+://[^/]+)') AS name,count(1) as value").
		Where("created_at between ? and ?", start, end).
		Group("name").
		Scan(&referMapList)

	var yv []int
	var xv []string
	for _, m := range countMapList {
		temp := cast.ToInt(m["total"])
		xv = append(xv, cast.ToString(m["day"]))
		yv = append(yv, temp)
	}

	xBuf, _ := json.Marshal(xv)
	yBuf, _ := json.Marshal(yv)
	countryMapListJson, _ := json.Marshal(countryMapList)
	referMapListJson, _ := json.Marshal(referMapList)

	c.HTML(200, "statistics.gohtml", OutputCommonSession(s.injector, c, gin.H{
		"selected":    "statistics",
		"referData":   template.JS(referMapListJson),
		"startDate":   start,
		"endDate":     end,
		"xAxis":       template.JS(xBuf),
		"yAxis":       template.JS(yBuf),
		"countryData": template.JS(countryMapListJson),
	}))
}

func (s *StatisticsHandler) Hit(c *gin.Context) {
	path, pathExist := c.GetQuery("path")
	ref, refExist := c.GetQuery("ref")
	var stat model.TbStatistics
	xForwardFor := c.GetHeader("X-Forwarded-For")
	userAgent := c.GetHeader("User-Agent")
	if !pathExist || !refExist || path == "" || xForwardFor == "" || userAgent == "" {
		return
	}
	arr := strings.Split(xForwardFor, ",")
	if len(arr) == 0 {
		return
	}

	ua := useragent.Parse(userAgent)
	if ua.Bot {
		return
	}
	if path == "index" {
		path = "/"
	}
	sha := sha256.New()
	sha.Write([]byte(fmt.Sprintf("%s%s", arr[0], time.Now().Format("20060102"))))
	stat.IP = arr[0]
	stat.IPHash = fmt.Sprintf("%x", sha.Sum(nil))

	var count int64
	s.db.Model(&model.TbStatistics{}).Where("ip_hash = ?", stat.IPHash).Count(&count)
	if count >= 1 {
		c.String(200, "ok")
		return
	}

	stat.Target = path
	stat.UpdatedAt = time.Now()
	stat.CreatedAt = time.Now()
	stat.Desktop = ua.Desktop
	stat.Mobile = ua.Mobile
	stat.Tablet = ua.Tablet
	stat.Device = ua.Device
	stat.Refer = ref
	stat.Country = c.GetHeader("CF-IPCountry")
	s.db.Save(&stat)
	c.String(200, "ok")
}
