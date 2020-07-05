package ohttp

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func parse(r *http.Request) map[string]string {
	s := strings.Split(r.URL.RawQuery, "&")
	result := map[string]string{}
	if len(s) == 0 {
		return result
	}
	for _, ss := range s {
		sss := strings.Split(ss, "=")
		if len(sss) == 2 {
			//for safe
			temp := template.HTMLEscapeString(sss[1])
			result[sss[0]] = temp
			// if chinese code
			if strings.Contains(temp, "%") {
				if temp2, err := url.QueryUnescape(temp); err == nil {
					result[sss[0]] = temp2
				}
			}

		} else if len(sss) == 1 {
			//just like aaa=&bbb=123 >> aaa
			result[sss[0]] = ""
		}
	}
	//if methhod POST wirte body to map body
	if r.Method == "POST" {
		bt, err := ioutil.ReadAll(r.Body)
		if err == nil {
			result["body"] = string(bt)
		}
	}
	//write ip to map
	result["ip"] = r.Header.Get("X-Forwarded-For") + "," + getIP(r.RemoteAddr)
	fmt.Println(result)
	return result
}
func getIP(s string) string {
	sArray := strings.Split(s, `:`)
	if len(sArray) == 2 {
		return sArray[0]
	}
	return s
}

//DoRespond return http
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
