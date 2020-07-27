/*
这是一个内存缓存的函数，唯一的调用就是Cache函数
输入为缓存的标识符，过期时间（最小1妙），以及一个单一返回值的interface{}的匿名函数f()interface{}.
他会在每120秒清理一次过期的缓存
同时避免了缓存穿透的问题
*/
package ohttp

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"
)

type cacheStruct struct {
	time int64
	lock int32
	date interface{}
}

var cache *sync.Map

var onceClearCache sync.Once

func clearCache() {
	const caching = 120
	go func() {
		for {
			time.Sleep(caching * time.Second)
			now := time.Now().Unix()
			cache.Range(func(key, value interface{}) bool {
				c, ok := value.(*cacheStruct)
				if ok && now-c.time > caching {
					cache.Delete(key)
				}
				return true
			})
		}
	}()
}

//Cache 输入名字,过期时间(秒),以一个匿名函数f（），这里的name是一个唯一标识，函数f请不要返回nil，如果有可能返回nil可能会造成性能问题。
func Cache(name string, seccnd int64, f func() interface{}) interface{} {
	onceClearCache.Do(clearCache) //每两分钟清理一次内存
	if seccnd <= 0 {              //缓存时间必须大于0
		seccnd = 1
	}
RE:
	//无论如何先读取或写入一个空结构，统一一下结构体。这里是唯一的写入，可以保证结构的一致。
	cc, _ := cache.LoadOrStore(name, &cacheStruct{})
	//理论上来说，这里应该用不到ok的结构。但是需要时间的测试
	c := cc.(*cacheStruct)

	if f == nil {
		return c.date
	}

	//判断当前内容是否过期
	if time.Now().Unix()-c.time > seccnd {
		//这里是为了避免过期时大量的访问落盘，仅可以一个请求访问，其他请求等待50ms后重新获取值。
		//c.lock==1时代表有锁
		if i := atomic.AddInt32(&c.lock, 1); i == 1 {
			result := f()
			//当f（）结果正常时，写入新的时间以及结果
			if result != nil {
				c.time = time.Now().Unix()
				c.date = result

			}
			//无论f（）执行是否正确，都取消锁
			atomic.StoreInt32(&c.lock, 0)
			return c.date
		}
		time.Sleep(50 * time.Millisecond)
		goto RE

	}
	return c.date
}

//RedisSet 一个简易的缓存,同时忽略了所有错误。请不要塞重要的内容在里面，塞即使小概率出错也无伤大雅的东西
func RedisSet(key, txt string, expire int64) {
	_, _ = rds.Multi([][]interface{}{
		{"select", 1},
		{"set", key, txt, "ex", expire},
	})
}

//RedisGet 一个简易的缓存,同时忽略了所有错误。请不要塞重要的内容在里面，塞即使小概率出错也无伤大雅的东西
func RedisGet(key string) string {
	s, _ := redis.String(rds.Do("get", key))
	return s
}
