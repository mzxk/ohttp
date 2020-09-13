package ohttp

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//这是一个非常关键的函数，优化他的效率可以优化整个效率。当然服务器资源都是够的，这套代码这辈子很难到达瓶颈，所以应该不会优化。。。。
//他解析了uri字符串，把所有的参数写入到一个map里
//如果不是get方法，还多写入了一个body的map
//还写入了一个ip，为了安全，写入所有可以获取的ip，以，分割
//调用了go原生的方法来避免奇怪字符的问题
//同时会把中文的url编码返回成中文
func parse(r *http.Request) map[string]string {
	s := strings.Split(r.URL.RawQuery, "&")
	result := map[string]string{}
	if len(s) == 0 {
		return result
	}
	for _, ss := range s {
		sss := strings.Split(ss, "=")
		if len(sss) == 2 {

			temp := template.HTMLEscapeString(sss[1]) //避免注入
			result[sss[0]] = temp
			// 翻译成中文字符
			if strings.Contains(temp, "%") {
				if temp2, err := url.QueryUnescape(temp); err == nil {
					result[sss[0]] = strings.TrimSpace(temp2)
				}
			}

		} else if len(sss) == 1 {
			//just like aaa=&bbb=123 >> aaa
			result[sss[0]] = ""
		}
	}
	//if methhod not POST wirte body to map body
	if r.Method != "GET" {
		lens := s2i(r.Header.Get("Content-Length"))
		if lens > 0 {
			bt, err := ioutil.ReadAll(r.Body)
			if err == nil {
				result["body"] = string(bt)
			}
		}
	}
	//write ip to map
	result["ip"] = r.Header.Get("X-Forwarded-For") + "," + getIP(r.RemoteAddr)
	return result
}
func getIP(s string) string {
	sArray := strings.Split(s, `:`)
	if len(sArray) == 2 {
		return sArray[0]
	}
	return s
}

//DoRespond 返回给前端的返回函数
//当err不是nil的时候，仅仅返回err，忽略i里所有的东西
//返回值m如果正常，一直是ok，如果不正常，返回不正常的原因
//返回值i是需要讨论的返回值结构
//返回值t是服务器unix时间
func doRespond(w http.ResponseWriter, i interface{}, err error) {
	rsp := &Respond{}
	if err != nil {
		rsp.Message = err.Error()
		i = nil
	} else {
		rsp.Message = "ok"
	}
	rsp.Result = i
	rsp.Time = time.Now().Unix()
	e := json.NewEncoder(w).Encode(rsp)

	if e != nil {
		log.Println("WrongWhenWriteRespond", e.Error())
	}
}

// Respond 返回值接口
type Respond struct {
	Message string      `json:"m"`
	Result  interface{} `json:"r"`
	Time    int64       `json:"t"`
}

func s2i(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		i = 0
	}
	return i
}
func mac(s, key string) string {
	bey := []byte(key)
	mac := hmac.New(sha512.New, bey)
	mac.Write([]byte(s))
	return hex.EncodeToString(mac.Sum(nil))[58:98]
}
