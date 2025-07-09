package main

import (
	"flag"
	"fmt"
	"io"            // 导入 io 包，替代部分 ioutil 功能
	"os"            // 导入 os 包，替代部分 ioutil 功能
	"path/filepath" // 用于字符串处理

	"gopkg.in/yaml.v2" // 用于 YAML 语法验证
)

// getEnvStr 获取环境变量字符串，如果不存在则使用默认值
func getEnvStr(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// renderResponse 用于向客户端返回 HTML 响应
func renderResponse(w io.Writer, title, message, backUrl string) { // 将 w 的类型从 http.ResponseWriter 改为 io.Writer
	// 对于 webhook execute-command，直接向 os.Stdout 写入 HTML
	// 所以这里的 w 是 os.Stdout
	htmlTemplate := fmt.Sprintf(`
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>%s</title>
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 20px; background-color: #f0f2f5; color: #333; line-height: 1.6; display: flex; flex-direction: column; justify-content: center; align-items: center; min-height: 100vh; text-align: center;}
            .container { max-width: 600px; margin: 20px auto; background-color: #ffffff; padding: 30px 40px; border-radius: 10px; box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
            h1 { color: #2c3e50; margin-bottom: 20px; font-size: 2em; }
            p { font-size: 1.1em; color: #555; margin-bottom: 25px; }
            button {
                padding: 12px 28px;
                border: none;
                border-radius: 6px;
                background-color: #007bff;
                color: white;
                font-size: 1.1em;
                cursor: pointer;
                transition: background-color 0.3s ease, transform 0.2s ease;
                box-shadow: 0 4px 8px rgba(0,123,255,0.2);
            }
            button:hover {
                background-color: #0056b3;
                transform: translateY(-2px);
            }
            button:active {
                transform: translateY(0);
                box-shadow: none;
            }
            .success { color: #28a745; }
            .error { color: #dc3545; }
            .error-detail { /* 添加错误详情样式 */
                display: block;
                background-color: #f8d7da;
                color: #721c24;
                border: 1px solid #f5c6cb;
                padding: 10px;
                margin-top: 10px;
                border-radius: 5px;
                text-align: left;
                white-space: pre-wrap; /* 保持换行 */
                word-break: break-all; /* 允许长单词换行 */
            }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>%s</h1>
            <p>%s</p>
            <button onclick="location.href='%s'">返回</button>
        </div>
    </body>
    </html>
    `, title, title, message, backUrl)
	fmt.Fprint(w, htmlTemplate)
}

func main() {
	var homeUrl string
	var content string

	flag.StringVar(&homeUrl, "home", "/ui", "Home URL for the Webhook")
	flag.StringVar(&content, "config-content", "", "Contents of the config")
	flag.Parse()

	configFilePath := getEnvStr("HOOKS", "/etc/webhook/hooks.yaml")

	prefix := getEnvStr("URL_PREFIX", "hooks")
	if prefix != "" {
		prefix = "/" + prefix
	}
	homeUrl = prefix + homeUrl

	if content == "" {
		renderResponse(os.Stdout, "错误", "未在请求中找到 'config' 字段内容。<br>请确认webhook配置正确传递了'-config'参数。", homeUrl)
		os.Exit(1)
	}

	// 1. YAML 语法验证
	var temp interface{}
	err := yaml.Unmarshal([]byte(content), &temp)
	if err != nil {
		// 使用 pre 标签来保留错误信息的格式，并添加 error-detail 类
		renderResponse(os.Stdout, "保存失败", fmt.Sprintf("YAML 语法错误，请检查: <span class='error-detail'><pre>%v</pre></span>", err), homeUrl)
		os.Exit(1)
	}

	// 2. 写入文件操作
	dir := filepath.Dir(configFilePath)
	// os.CreateTemp 替代 ioutil.TempFile
	tmpFile, err := os.CreateTemp(dir, "hooks-temp-*.yaml")
	if err != nil {
		renderResponse(os.Stdout, "保存失败", fmt.Sprintf("无法创建临时文件: %v", err), homeUrl)
		os.Exit(1)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		renderResponse(os.Stdout, "保存失败", fmt.Sprintf("无法写入临时文件: %v", err), homeUrl)
		return
	}

	// 确保所有数据都已写入磁盘
	err = tmpFile.Sync()
	if err != nil {
		renderResponse(os.Stdout, "保存失败", fmt.Sprintf("无法同步临时文件到磁盘: %v", err), homeUrl)
		return
	}

	// 关闭临时文件，否则在 Windows 上 os.Rename 可能会失败
	tmpFile.Close()

	// 原子性替换原文件
	err = os.Rename(tmpFile.Name(), configFilePath)
	if err != nil {
		renderResponse(os.Stdout, "保存失败", fmt.Sprintf("无法替换配置文件: %v<br>请检查文件权限。", err), homeUrl)
		os.Exit(1)
	}

	// 3. 返回成功响应
	renderResponse(os.Stdout, "保存成功", "Webhook 配置已成功更新！<br>Webhook 服务可能需要重启才能加载新配置。", homeUrl)
}
