package ohttp

import (
	"errors"
	"fmt"
)

//Sms 一个发送短信的接口
type Sms interface {
	//发送 手机号码 ， 模版id ， code代码
	Send(phone, id, code string) error
}

//NewSms 发送短信的接口，现在支持"juhe"和"submail"
//submail暂时不写了
func NewSms(name, key string) Sms {
	switch name {
	case "juhe":
		return &juheSms{key}
	}
	return nil
}

type juheSms struct {
	key string
}

func (t *juheSms) Send(phone, id, code string) error {
	rst, err := HTTP("http://v.juhe.cn/sms/send",
		map[string]interface{}{
			"mobile":    phone,
			"tpl_id":    id,
			"tpl_value": "#code#=" + code,
			"key":       t.key,
		}).Get()
	if err != nil {
		return err
	}
	var result map[string]interface{}
	err = rst.JSON(&result)
	if err != nil {
		return err
	}
	errCode := fmt.Sprint(result["error_code"])
	if errCode == "" || errCode == "0" {
		return nil
	}
	return errors.New(fmt.Sprint(result["reason"]))
}
