package task

import (
	"github.com/jasonlvhit/gocron"
	"github.com/kingwrcy/hn/model"
	"github.com/samber/do"
	"gorm.io/gorm"
	"log"
	"math"
	"time"
)

func StartPostTask(i *do.Injector) {
	s := gocron.NewScheduler()
	s.Every(1).Minute().Do(func(i *do.Injector) {
		log.Printf("start refresh last 7 days post points.")
		db := do.MustInvoke[*gorm.DB](i)

		var posts []model.TbPost
		db.Model(&model.TbPost{}).Where("created_at >= now() - interval 7 day and status = 'Active'").Scan(&posts)
		g := 1.80
		for _, post := range posts {
			p := float64(post.UpVote - 1)
			if p < 0 {
				p = 0
			}
			t := time.Now().Sub(post.CreatedAt).Hours()
			point := p / math.Pow(t+2, g)

			db.Model(&model.TbPost{}).Where("id= ?", post.ID).Update("point", point)
		}
		log.Printf("end of refresh last 7 days post points.")
	}, i)
	s.Start()
}
