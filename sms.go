package ohttp

import (
	"errors"
	"fmt"
	"net/url"
)

//Sms 一个发送短信的接口
type Sms interface {
	//发送 手机号码 ， 模版id ， code代码
	Send(phone, id, code string) error
}

//NewSms 发送短信的接口，现在支持"juhe"和"maixun"
//当错误的短信提供商时，直接panic，既然调用了就必须是正常的。
//submail暂时不写了
func NewSms(name, key, value string) Sms {
	switch name {
	case "juhe":
		return &juheSms{key}
	case "maixun":
		return &maixun{key, value}
	default:
		panic("wrongSmsName")
	}
	return nil
}

type maixun struct {
	key   string
	value string
}

func (t *maixun) Send(phone, id, code string) error {
	vals := url.Values{}
	vals.Add("account", t.key)
	vals.Add("pswd", t.value)
	vals.Add("mobile", phone)
	//TODO 修改签名
	vals.Add("msg", fmt.Sprintf("【方舟】您的验证码是%v", code))
	vals.Add("needstatus", "true")
	vals.Add("product", id)
	vals.Add("resptype", "json")
	rst, err := HTTP("http://www.weiwebs.cn/msg/HttpBatchSendSM",
		nil).Header(map[string]string{"Content-Type": "application/x-www-form-urlencoded"}).Post([]byte(vals.Encode()))
	fmt.Println(err)
	fmt.Println(rst.String())
	fmt.Println(rst.URL)
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
