package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"stock_data_cache/cache"
	"time"
)

func main() {
	cache.NewGroup(cache.Sina, 2<<26, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("request sina", key)
			v, err := cache.RequestSina(key)
			if err != nil {
				msg := fmt.Sprintf("request sina failed, error: %s", err.Error())
				fmt.Println(msg)
				return nil, errors.New(msg)
			}
			return []byte(v), nil
		}))
	g := cache.GetGroup(cache.Sina)
	g.LoadCache()
	go func() {
		for {
			select {
			case <- time.After(time.Hour):
				g.SaveCache()
			}
		}
	}()
	go func() {
		for {
			empty, _ := g.RemoteUpdateCache()
			if empty {
				<- time.After(time.Minute)
			}
		}
	}()

	addr := "0.0.0.0:7295"
	peers := cache.NewHTTPPool(addr)
	log.Println("server is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
