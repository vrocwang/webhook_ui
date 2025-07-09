## Edit for Webhook

## 使用说明：
使用三个参数title,home,save,分别表示页面title，ui页面链接，保存链接
```shell
# 脚本已考虑了URL_PREFIX问题，无需考虑是否需要在链接前加URL_PREFIX
# 因为已经有默认值，如果不修改hooks，则不带参数也可运行
# 需要环境变量URL_PREFIX。含义见webhook项目
./edit_form -title "edit" -home "/ui" -save "/save"
```

## 编译
```shell
go build -ldflags "-w -s" -o edit_form .
```
