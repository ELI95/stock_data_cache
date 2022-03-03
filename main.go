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
			v, err := cache.RequestSina(key, time.Second*5)
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
			case <-time.After(time.Hour):
				g.SaveCache()
			}
		}
	}()
	go func() {
		for {
			select {
			case <-time.After(time.Minute * 10):
				g.SendTimeoutCache(100)
			}
		}
	}()
	go func() {
		for {
			<-time.After(time.Second)
			empty, _ := g.RemoteUpdateCache()
			if empty {
				<-time.After(time.Second * 10)
			}
		}
	}()

	addr := "0.0.0.0:7296"
	peers := cache.NewHTTPPool(addr)
	log.Println("server is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
