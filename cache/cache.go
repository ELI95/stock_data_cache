package cache

import (
	"fmt"
	"golang.org/x/sync/singleflight"
	"os"
	"stock_data_cache/utils"
	"sync"
	"time"
)

const FilePath = "/tmp/cache.gob"

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	b []byte
}

// Len returns the view's length
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice returns a copy of the data as a byte slice.
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

type cache struct {
	mu         sync.Mutex
	lru        *Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	sg        *singleflight.Group
}

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		sg:        &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		fmt.Printf("cache hit, key: %s\n", key)
		return v, nil
	}

	fmt.Printf("cache miss, key: %s\n", key)
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	v, err, _ := g.sg.Do(key, func() (interface{}, error) {
		return g.getLocally(key)
	})

	if err == nil {
		return v.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) UpdateCache(num int) {
	defer utils.TimeTrack(time.Now(), "UpdateCache")

	keys := make([]string, 0)

	g.mainCache.mu.Lock()
	ele := g.mainCache.lru.ll.Front()
	for {
		if ele == nil {
			break
		}
		kv := ele.Value.(*entry)
		keys = append(keys, kv.key)
		if ele == g.mainCache.lru.ll.Back() {
			break
		}
		if len(keys) == num {
			break
		}
		ele = ele.Next()
	}
	g.mainCache.mu.Unlock()

	var succeed int
	kvs := make(map[string]string)
	for _, key := range keys {
		v, err := RequestSina(key)
		if err != nil {
			fmt.Printf("request sina failed, err: %s\n", err.Error())
			continue
		}
		value := ByteView{b: cloneBytes([]byte(v))}
		g.mainCache.add(key, value)
		kvs[key] = v
		succeed++
	}
	fmt.Printf("update cache done, total: %d, succeed: %d\n", len(keys), succeed)

	g.SaveCache(kvs)
}

func (g *Group) SaveCache(kvs map[string]string) {
	if err := utils.Save(FilePath, kvs); err != nil {
		fmt.Printf("save cache failed, error: %s\n", err.Error())
	}
	fmt.Println("save cache done")
}

func (g *Group) LoadCache() {
	if _, err := os.Stat(FilePath); os.IsNotExist(err) {
		fmt.Printf("file not exist, file: %s\n", FilePath)
		return
	}

	kvs := make(map[string]string)
	if err := utils.Load(FilePath, &kvs); err != nil {
		fmt.Printf("load cache failed, error: %s\n", err.Error())
		return
	}

	for k, v := range kvs {
		g.mainCache.add(k, ByteView{b: cloneBytes([]byte(v))})
	}

	fmt.Println("load cache done")
}
