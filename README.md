### stock_data_cache

#### 访问方式
```
curl http://localhost:7295/cache/sina?key=https://hq.sinajs.cn/list=sz000001
```

#### 流程
- lru + timeout refresh + singleflight
- 若缓存命中, 返回数据
- 若缓存未命中, 发起请求成功后，更新缓存，返回数据
- 若缓存过期，对应链表节点MoveToFront，先返回旧数据, 新协程发起请求成功后，更新缓存，并将全部缓存持久化到磁盘
- 定时更新部分最常使用缓存，并将全部缓存持久化到磁盘
- 重启时加载缓存文件，更新部分最常使用缓存
