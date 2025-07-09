## Save configure for Webhook

## 使用说明：
使用两个参数home,config-content,分别表示ui页面链接，配置文件内容
```shell
# 脚本已考虑了URL_PREFIX问题，无需考虑是否需要在链接前加URL_PREFIX
# 
# 需要环境变量HOOKS、URL_PREFIX。含义见webhook项目
./save -home "/ui" -config-content "<页面传入>"
```

## 编译
```shell
go build -ldflags "-w -s" -o save .
```
