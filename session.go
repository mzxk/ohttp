package ohttp

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mzxk/oredis"
	"github.com/mzxk/oval"
)

var (
	expireTime   int64           = 3600 * 24                //标准的过期时间，24小时，这个时间会在用户key使用后重新刷新成24小时
	hostSalt     string          = sha(time.Now().String()) //生成用户key，value时的盐
	rds          *oredis.Oredis                             //redis客户端
	sessionCache *oval.ExpireMap = oval.NewExpire()         //一个带过期时间的内存map，目的是为了减少redis的调用次数，提高效率
)

type session struct {
	User       string //用户的唯一标识符，通常这是用户的bsonID，当然也可以ivi其他的东西
	Key        string //用户的key
	Value      string //用户的value
	ExpireTime int64  //过期时间
}

//InitSession 使用redis保存session,所以必须输入redis的用户名和密码，默认的，这使用localhost
func initSession(add, pwd string) {
	if add == "" {
		add = "127.0.0.1:6379"
	}
	rds = oredis.New(add, pwd)
}

//AddSession 生成一个用户key和value并返回
//同时保存到redis里
func AddSession(user string) (key, value string, err error) {
	key, value = getKV(user)
	ex := expireTime
	s := &session{
		User:       user,
		Key:        key,
		Value:      value,
		ExpireTime: time.Now().Unix() + ex,
	}
	err = saveSession(s)
	return
}

//DeleteSession 删除用户的当前的令牌
func DeleteSession(key string) {
	deleteSession(key)
}
func deleteSession(key string) {
	_, _ = rds.Do("del", key)
}

//每当令牌用户使用过，就提高他的过期时间
//为了提高效率，先调用内存验证，当时间超过1小时后，update到redis里。
//当然，这将带来额外的内存消耗，不过通常来说，redis服务器和本地内存不在一起。而上G的token是一个很可怕的网站，我的项目可能很难做到。
func updateSession(key string) {
	tm, load := sessionCache.Load(key)
	if !load {
		sessionCache.Store(key, int64(0), expireTime)
	}
	if load && time.Now().Unix()-tm.(int64) > 3600 {
		sessionCache.Expire(key, expireTime)
		_, err := rds.Do("expire", key, expireTime)
		if err != nil {
			log.Println("error", err)
		}

	}
}

//通过传入的key,从redis中读取用户的id和value
//很奇怪的这里我没有读取本地缓存，人懒～
func loadSession(key string) (user, value string) {
	if key == "" {
		return
	}
	m, err := redis.StringMap(rds.Do("hgetall", key))
	if err != nil {
		log.Println("GetSession", err)
		return
	}
	if m != nil {
		return m["u"], m["v"]
	}
	return
}

//保存token到redis里并设置过期时间
func saveSession(s *session) error {
	_, err := rds.Multi([][]interface{}{
		{"hmset", s.Key, "u", s.User, "v", s.Value},
		{"expire", s.Key, expireTime},
	})
	return err
}
func getKV(user string) (k, v string) {
	s := sha(user + time.Now().String() + hostSalt)
	return s[1:11], s[13:30]
}

//这是一个伪装，伪装自己是个sha1,避免彩虹表
func sha(s string) string {
	k := sha512.Sum512([]byte(s))
	return hex.EncodeToString(k[:])[58:98]
}
