package main

import (
	"fmt"
	"log"
	"net/http"
	"stock_data_cache/cache"
	"stock_data_cache/cron"
)

func main() {
	cache.NewGroup(cache.Sina, 2<<26, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("request sina", key)
			v, err := cache.RequestSina(key)
			if err != nil {
				return nil, fmt.Errorf("request sina failed, err: %s", err.Error())
			}
			return []byte(v), nil
		}))
	g := cache.GetGroup(cache.Sina)
	g.LoadCache()
	cron.RunCrontabJob()

	addr := "0.0.0.0:7295"
	peers := cache.NewHTTPPool(addr)
	log.Println("server is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
