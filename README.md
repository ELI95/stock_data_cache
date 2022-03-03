### stock_data_cache

#### 访问方式
```
curl http://localhost:7295/cache/sina?key=https://hq.sinajs.cn/list=sz000001
```

#### 流程
- lru + singleflight
- 若缓存命中, 返回数据
- 若缓存未命中, 加入待更新channel，返回500，
- 若缓存过期，加入待更新channel，返回过期数据
- master定时保存缓存文件
- master定时检查过期缓存
- slave更新缓存
