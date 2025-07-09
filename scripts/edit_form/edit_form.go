package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"os"
)

func getEnvStr(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {

	// 定义命令行参数
	title := flag.String("title", "Edit Webhook Configuration", "Title for the Edit Page")
	homeUrl := flag.String("home", "/ui", "Home URL for the Webhook")
	saveUrl := flag.String("save", "/save", "URL for save form")

	flag.Parse()

	configFilePath := getEnvStr("HOOKS", "/app/hooks.yaml")

	// 定义要传递给模板的数据结构
	type TemplateData struct {
		ConfigContent string
		Title         string
		HomeUrl       string
		SaveUrl       string
	}

	var configContent string

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在时，提供一个空的 YAML 列表作为初始内容
			configContent = "[]\n"
			fmt.Fprintf(os.Stderr, "Config file not found at %s, starting with empty config.\n", configFilePath)
		} else {
			// 其他读取错误
			configContent = fmt.Sprintf("# Error loading config from %s: %v\n# Please check server logs.\n[]\n", configFilePath, err)
			fmt.Fprintf(os.Stderr, "Error reading config file from %s: %v\n", configFilePath, err)
			// 这里不直接退出，以便用户可以看到错误信息并在页面中尝试编辑或保存
		}
	} else {
		configContent = string(data) // 将读取到的字节数据直接转换为字符串
	}

	prefix := getEnvStr("URL_PREFIX", "hooks")
	if prefix != "" {
		prefix = "/" + prefix
	}

	templateData := TemplateData{
		ConfigContent: configContent,
		Title:         *title,
		HomeUrl:       prefix + *homeUrl,
		SaveUrl:       prefix + *saveUrl,
	}

	// HTML 模板
	htmlTemplate := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>编辑 Webhook 配置</title>
		<style>
			body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 20px; background-color: #f0f2f5; color: #333; line-height: 1.6; }
			.container { max-width: 800px; margin: 20px auto; background-color: #ffffff; padding: 30px 40px; border-radius: 10px; box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
			h1 { text-align: center; color: #2c3e50; margin-bottom: 30px; font-size: 2.2em; font-weight: 600; border-bottom: 2px solid #e0e0e0; padding-bottom: 15px; }
			form { margin-top: 30px; }
			textarea {
				width: 100%;
				height: 400px;
				padding: 15px;
				margin-bottom: 20px;
				border: 1px solid #ced4da;
				border-radius: 8px;
				font-family: 'Cascadia Code', 'Fira Code', monospace;
				font-size: 1em;
				box-sizing: border-box;
				resize: vertical; /* 允许垂直方向调整大小 */
				background-color: #f8f9fa;
				color: #495057;
			}
			.button-group { text-align: center; margin-top: 20px; }
			.button-group button {
				padding: 12px 28px;
				margin: 0 10px;
				border: none;
				border-radius: 6px;
				background-color: #007bff;
				color: white;
				font-size: 1.1em;
				cursor: pointer;
				transition: background-color 0.3s ease, transform 0.2s ease;
				box-shadow: 0 4px 8px rgba(0,123,255,0.2);
			}
			.button-group button:hover {
				background-color: #0056b3;
				transform: translateY(-2px);
			}
			.button-group button:active {
				transform: translateY(0);
				box-shadow: none;
			}
			.button-group button.cancel {
				background-color: #6c757d;
			}
			.button-group button.cancel:hover {
				background-color: #5a6268;
			}
			.hint { text-align: center; color: #6c757d; margin-top: 30px; font-size: 0.9em; padding: 15px; border: 1px solid #dee2e6; border-radius: 8px; background-color: #fff3cd; border-color: #ffeeba; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>编辑 Webhook 配置</h1>
			<form action="{{ .SaveUrl }}" method="POST">
				<textarea name="config" rows="20" cols="80">{{ .ConfigContent }}</textarea>
				<div class="button-group">
					<button type="submit">保存更改</button>
					<button type="button" class="cancel" onclick="location.href='{{ .HomeUrl }}'">取消并返回</button>
				</div>
			</form>
			<p class="hint">
				在此处修改您的 Webhook 配置 (YAML 格式)。
				点击 "保存更改" 将把新的配置发送到服务器。
				请确保 YAML 语法正确，否则可能导致 Webhook 服务无法正常启动。
			</p>
		</div>
	</body>
	</html>
	`

	tmpl, err := template.New("edit").Parse(htmlTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML template for edit form: %v\n", err)
		fmt.Print("<h1>Error: Internal UI rendering issue.</h1>")
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData) // 传入字符串内容
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template for edit form: %v\n", err)
		fmt.Print("<h1>Error: Internal UI rendering issue.</h1>")
		os.Exit(1)
	}

	fmt.Print(buf.String())
}
