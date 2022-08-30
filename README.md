# eapollo
## 使用方式
```
go get github.com/ego-component/eapollo/conf

在main.go中引入
import (
    _ "github.com/ego-component/eapollo/conf"
)

在启动的时候调用    
go run main.gon --config=apollo://ip:port?appId=XXX&cluster=XXX&namespaceName=XXX&configKey=XXX&configType=toml&accesskeySecret=XXX&insecureSkipVerify=XXX&cacheDir=XXX
```