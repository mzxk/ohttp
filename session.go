package ohttp

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/mzxk/oredis"
	"github.com/mzxk/oval"
)

var (
	expireTime   int64  = 7200
	hostSalt     string = sha(time.Now().String())
	rds          *oredis.Oredis
	sessionCache *oval.ExpireMap = oval.NewExpire()
)

type session struct {
	User       string
	Key        string
	Value      string
	ExpireTime int64
}

//InitSession if use session ,must input redis url and pwd
func initSession(add, pwd string) {
	if add == "" {
		add = "127.0.0.1:6379"
	}
	rds = oredis.New(add, pwd)
}

//AddSession newSession
func AddSession(user string, exTime ...int64) (key, value string) {
	key, value = getKV(user)
	ex := expireTime
	if len(exTime) >= 1 {
		ex = exTime[0]
	}
	s := &session{
		User:       user,
		Key:        key,
		Value:      value,
		ExpireTime: time.Now().Unix() + ex,
	}
	saveSession(s)
	return
}

//DeleteSession del session
func DeleteSession(key string) {
	deleteSession(key)
}
func deleteSession(key string) {
	c := rds.Get()
	defer c.Close()
	c.Do("del", key)
}
func updateSession(key string) {
	tm, load := sessionCache.Load(key)
	if !load {
		sessionCache.Store(key, 0, expireTime)
	}
	if load && time.Now().Unix()-tm.(int64) > 3600 {
		sessionCache.Expire(key, expireTime)
		c := rds.Get()
		defer c.Close()
		c.Do("expire", key, expireTime)

	}
}
func loadSession(key string) (user, value string) {
	if key == "" {
		return
	}
	c := rds.Get()
	defer c.Close()
	m, err := redis.StringMap(c.Do("hgetall", key))
	if err != nil {
		log.Println("GetSession", err)
		return
	}
	if m != nil {
		return m["u"], m["v"]
	}
	return
}
func saveSession(s *session) error {
	c := rds.Get()
	defer c.Close()
	c.Send("multi")
	c.Send("hmset", s.Key, "u", s.User, "v", s.Value)
	c.Send("expire", s.Key, expireTime)
	_, err := c.Do("exec")
	return err
}
func getKV(user string) (k, v string) {
	s := sha(user + time.Now().String() + hostSalt)
	return s[1:11], s[13:30]
}
func sha(s string) string {
	k := sha1.Sum([]byte(s))
	return hex.EncodeToString(k[:])
}
