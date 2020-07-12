# ohttp

这个包进行了net/http的易用的封装，包含client端和server端，以及一个简单的内存缓存。并且用简单易用的方式维持session和进行签名
当然，这几乎是一个耦合性很高自用的包

###最简单的例子

```cgo
package main

import "github.com/mzxk/ohttp"

func main() {
	h := ohttp.New()
	h.Add("/foo", bar)
	h.Run(":6666")
}
func bar(p map[string]string) (interface{}, error) {
	return p["hi"], nil
}
```
调用 curl http://localhost:6666/foo?hi=nihao
返回值是
```cgo

{"m":"ok","r":"nihao","t":1594542588}
```
