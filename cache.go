package ohttp

import (
	"sync"
	"sync/atomic"
	"time"
)

type cacheStruct struct {
	time int64
	lock int32
	date interface{}
}

var cache *sync.Map

// var onceClearCache sync.Once

// func clearCache() {
// 	const cacheing = 120
// 	var temp sync.Map
// 	cache = &temp
// 	go func() {
// 		for {
// 			time.Sleep(cacheing * time.Second)
// 			var temp2 sync.Map
// 			now := time.Now().Unix()
// 			cache.Range(func(key, value interface{}) bool {
// 				c, ok := value.(*cacheStruct)
// 				if ok || now-c.time < cacheing {
// 					temp2.Store(key, value)
// 				}
// 				return false
// 			})
// 			cache = &temp2
// 		}
// 	}()
// }

//Cache 输入名字,过期时间(秒),一致的id,如果 id相同,那么即使超时了也不重新读取 需要运行的函数,如果超时,那么运行函数并把返回值缓存到名字里
func Cache(name string, seccnd int64, f func() interface{}) interface{} {
	// onceClearCache.Do(clearCache)
	if seccnd <= 0 {
		seccnd = 1
	}
RE:
	cc, _ := cache.LoadOrStore(name, &cacheStruct{})

	c, ok := cc.(*cacheStruct)
	if !ok {
		cache.Delete(name)
		goto RE
	}
	if f == nil {
		return c.date
	}

	if time.Now().Unix()-c.time > seccnd {
		if i := atomic.AddInt32(&c.lock, 1); i == 1 {
			result := f()

			if result != nil {
				c.time = time.Now().Unix()
				c.date = result

			}
			atomic.StoreInt32(&c.lock, 0)
			return c.date
		}
		time.Sleep(50 * time.Millisecond)
		goto RE

	}
	return c.date
}
