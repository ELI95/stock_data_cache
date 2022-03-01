package cron

import (
	"github.com/robfig/cron"
	"stock_data_cache/cache"
)

func RunCrontabJob() {
	c := cron.New()
	_ = c.AddFunc("0 0 12,15 * * 1-5", func() {
		g := cache.GetGroup(cache.Sina)
		g.UpdateCache(100, cache.ExpireMinutes)
		g.SaveCache()
	})
	c.Start()
}
