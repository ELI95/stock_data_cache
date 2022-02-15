### stock_data_cache

#### 访问方式
```
curl http://localhost:7295/cache/sina?key=https://hq.sinajs.cn/list=sz000001
```

#### 流程
- lru + timeout refresh + singleflight
- 若缓存不存在或者缓存过期，请求新浪接口成功后，更新lru
