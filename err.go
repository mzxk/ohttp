package ohttp

import "errors"

var errs = struct {
	signExpire    error
	outLimit      error
	unknownMethod error
	signWrong     error
	signTimeWrong error
	codeWrong     error
}{
	errors.New("登录已过期"),
	errors.New("接口访问超过限制"),
	errors.New("未知路由"),
	errors.New("签名不正确"),
	errors.New("请检查系统时间"),
	errors.New("验证码错误"),
}
