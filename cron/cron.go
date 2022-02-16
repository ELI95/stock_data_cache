package cron

import (
	"github.com/robfig/cron"
	"stock_data_cache/cache"
)

func RunCrontabJob() {
	c := cron.New()
	_ = c.AddFunc("0 0,30 9-11,13-15 * * 1-5", func() {
		g := cache.GetGroup(cache.Sina)
		g.UpdateCache(1000, cache.ExpireMinutes)
		g.SaveCache()
	})
	c.Start()
}
