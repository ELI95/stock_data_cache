package cron

import (
	"github.com/robfig/cron"
	"stock_data_cache/cache"
)

func RunCrontabJob() {
	c := cron.New()
	_ = c.AddFunc("0,30 9-11,13-15 * * *", func() {
		g := cache.GetGroup(cache.Sina)
		g.UpdateCache(1000)
	})
	c.Start()
}
