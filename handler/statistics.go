package handler

import (
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kingwrcy/hn/model"
	"github.com/mileusna/useragent"
	"github.com/samber/do"
	"github.com/spf13/cast"
	"gorm.io/gorm"
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
	start, startExist := c.GetQuery("start")
	if !startExist {
		start = time.Now().Format("2006-01-02")
	}
	end, endExist := c.GetQuery("end")
	if !endExist {
		end = time.Now().Add(1 * time.Hour * 24).Format("2006-01-02")
	}
	var total int64
	var countryMapList []map[string]interface{}
	var referMapList []map[string]interface{}
	s.db.Model(&model.TbStatistics{}).Where("date(created_at) between date(?) and date(?)", start, end).Count(&total)
	s.db.Model(&model.TbStatistics{}).Select("country,count(1) as total,DATE_FORMAT(created_at,'%Y-%m-%d') as day").
		Where("date(created_at) between date(?) and date(?)", start, end).Group("country,DATE_FORMAT(created_at,'%Y-%m-%d')").
		Order("country").Scan(&countryMapList)
	s.db.Model(&model.TbStatistics{}).Select("refer,count(1) as total,DATE_FORMAT(created_at,'%Y-%m-%d') as day").
		Where("date(created_at) between date(?) and date(?)", start, end).Group("refer,DATE_FORMAT(created_at,'%Y-%m-%d')").
		Order("refer").Scan(&referMapList)

	finalCountryMapList := map[string][]map[string]interface{}{}
	finalReferMapList := map[string][]map[string]interface{}{}

	for _, item := range countryMapList {
		day := cast.ToString(item["day"])
		if _, ok := finalCountryMapList[day]; !ok {
			finalCountryMapList[day] = []map[string]interface{}{}
		}
		finalCountryMapList[day] = append(finalCountryMapList[day], item)
	}
	for _, item := range referMapList {
		day := cast.ToString(item["day"])
		if _, ok := finalReferMapList[day]; !ok {
			finalReferMapList[day] = []map[string]interface{}{}
		}
		finalReferMapList[day] = append(finalReferMapList[day], item)
	}

	c.HTML(200, "statistics.gohtml", OutputCommonSession(s.db, c, gin.H{
		"selected":       "statistics",
		"total":          total,
		"countryMapList": finalCountryMapList,
		"referMapList":   finalReferMapList,
		"startDate":      start,
		"endDate":        end,
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
	if count == 0 {
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
