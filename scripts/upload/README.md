## Upload for Webhook

## 使用说明：
使用两个参数home,file-name,分别表示ui页面链接，文件名
```shell
# 脚本已考虑了URL_PREFIX问题，无需考虑是否需要在链接前加URL_PREFIX
#
# 需要环境变量URL_PREFIX。含义见webhook项目。需要环境变量UPLOAD_DEST_DIR，上传目录；UPLOADED_FILE_PATH，临时路径
./updata -home "/ui" -file-name "<页面传入>"
```

## 编译
```shell
go build -ldflags "-w -s" -o updata .
```
