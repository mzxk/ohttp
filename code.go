package ohttp

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"
)

var nonceSeed = rand.NewSource(time.Now().Unix() / 2)

func CodeGet(key string, time, limit int64) (string, error) {
	code := fmt.Sprintf("%06d", rand.New(nonceSeed).Intn(999999))
	c := rds.Get()
	defer c.Close()
	_ = c.Send("hmset", key, "limit", limit, "code", code)
	_ = c.Send("expire", key, time)
	err := c.Flush()
	return code, err
}

//现在忽略了所有的错误，未来将针对错误进行某些处理
func CodeCheck(key, code string) bool {
	return codeCheck(key, code) == nil
}
func codeCheck(key, code string) error {
	rd := rds.Get()
	defer rd.Close()
	//首先确认这个code是否存在，如果不存在，返回错误
	codeStore, err := redis.String(rd.Do("hget", key, "code"))
	if err != nil {
		return err
	}
	//确认code是否相同，如果相同，返回删除这个的错误，通常这里都是nil
	if codeStore == code {
		_, err = rd.Do("del", key)
		return err
	}
	//到这里代表不相同，那么就减少重新验证的次数
	i, err := redis.Int(rd.Do("hincrby", key, "limit", -1))
	if err != nil {
		//通常的，这里发生任何错误，都要删除这个key，因为理论上来说，如果不是键重复或者网络链接失败，都是不会发生错误都
		_, _ = rd.Do("del", key)
		return err
	}
	//到这里代表重试次数已经超过限制了，那么需要删除这个key，通常的
	if i < 0 {
		_, _ = rd.Do("del", key)
		return errors.New("outLimit")
	}
	//到这里代表还有重试次数，但是验证码是错误的
	return errors.New("codeWrong")
}
