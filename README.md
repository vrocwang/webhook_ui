***注意：本项目主要是为了研究[webhook](https://github.com/soulteary/webhook.git)项目的用法，不保证代码质量，不保证性能，一般情况下也不会更新***

## 项目介绍
基于[webhook](https://github.com/soulteary/webhook.git)特性创建的web页面。
### 功能列表
- [x] ui: ui页面，展示配置信息，修改配置、上传可执行文件按钮
- [x] edit_form: 编辑配置页面，展示yaml样式配置信息，保存配置、取消按钮
- [x] upload_form: 上传可执行文件页面，上传可执行文件、返回按钮，上传目录当前存在的文件
- [x] upload: 上传可执行文件，上传成功后返回ui

## 使用说明
***前提条件：docker，docker-compose需要安装好***
* 1、编译scripts下的各个脚本
* 2、基于编译生成文件名称及参数修改config/hooks.yaml
* 3、执行docker-compose up -d，启动项目
* 4、访问http://<ip>:8002/ui，访问ui页面