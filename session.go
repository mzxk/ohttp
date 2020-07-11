/*
	Sign的方式：
		1：登陆，获取后端返回的key和value。
				例如：key = foo , value = bar
		2：在正常的请求上url，增加两个参数：
			nonce	时间戳 精确到秒
			key		login后后端返回的key，value之一的key
				例如：
					原始请求是http://localhost/user/changePwd?oldpwd=old&newpwd=new
					增加后变成http://localhost/user/changePwd?oldpwd=old&newpwd=new&key=foo&nonce=158754624
		3：计算sign
			使用整个requestURI 不包含域名和端口的部分 加上 value 进行sha512计算 ， 然后截取[58:98]
				例如：
					needsign:="/user/changePwd?oldpwd=old&newpwd=new&key=foo&nonce=158754624"+"bar"
					sign=sha512(needsign)[58:98]
					"bf11206c32dc27301f681da41fc3570937c1e2cd"
			然后在请求request里增加一个header，名字为key,值为上面的sign

		请注意：
			如果需要加密的字符串里含有中文字符,那么一定要在url被转码成%ab的格式后在进行签名，直接使用中文拼接进行的签名是无法通过检查的
*/

package ohttp

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/mzxk/oredis"
	"github.com/mzxk/oval"
)

var (
	expireTime   int64  = 3600 * 24
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
func AddSession(user string, exTime ...int64) (key, value string, err error) {
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
	err = saveSession(s)
	return
}

//DeleteSession del session
func DeleteSession(key string) {
	deleteSession(key)
}
func deleteSession(key string) {
	c := rds.Get()
	defer c.Close()
	_, _ = c.Do("del", key)
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
		_, err := c.Do("expire", key, expireTime)
		if err != nil {
			log.Println("error", err)
		}

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
	_ = c.Send("multi")
	_ = c.Send("hmset", s.Key, "u", s.User, "v", s.Value)
	_ = c.Send("expire", s.Key, expireTime)
	_, err := c.Do("exec")
	return err
}
func getKV(user string) (k, v string) {
	s := sha(user + time.Now().String() + hostSalt)
	return s[1:11], s[13:30]
}
func sha(s string) string {
	k := sha512.Sum512([]byte(s))
	return hex.EncodeToString(k[:])[58:98]
}
