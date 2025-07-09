## UI for Webhook

## 使用说明：
使用三个参数title,edit,upload,分别表示页面title，编辑页面链接，上传页面链接
```shell
# 脚本已考虑了URL_PREFIX问题，无需考虑是否需要在链接前加URL_PREFIX
# 因为已经有默认值，如果不修改hooks，则不带参数也可运行
# 需要环境变量HOOKS、URL_PREFIX。含义见webhook项目
./ui -title "abc" -edit "/edit_form" -upload "/upload_form"
```

## 编译
```shell
go build -ldflags "-w -s" -o ui .
```
