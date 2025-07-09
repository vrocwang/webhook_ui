package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"os"

	"gopkg.in/yaml.v2" // 需要安装：go get gopkg.in/yaml.v2
)

// 定义与 webhook 配置文件匹配的 Go 结构体
type Config []Hook

type Header struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}

type ResponseHeaders []Header

type Argument struct {
	Source       string `yaml:"source,omitempty" json:"source,omitempty"`
	Name         string `yaml:"name,omitempty" json:"name,omitempty"`
	EnvName      string `yaml:"envname,omitempty" json:"envname,omitempty"`
	Base64Decode bool   `yaml:"base64decode,omitempty" json:"base64decode,omitempty"`
}

type Rules struct {
	And   *AndRule   `yaml:"and,omitempty" json:"and,omitempty"`
	Or    *OrRule    `yaml:"or,omitempty" json:"or,omitempty"`
	Not   *NotRule   `yaml:"not,omitempty" json:"not,omitempty"`
	Match *MatchRule `yaml:"match,omitempty" json:"match,omitempty"`
}

type AndRule []Rules

type OrRule []Rules

type NotRule Rules

type MatchRule struct {
	Type      string   `yaml:"type,omitempty" json:"type,omitempty"`
	Regex     string   `yaml:"regex,omitempty" json:"regex,omitempty"`
	Secret    string   `yaml:"secret,omitempty" json:"secret,omitempty"`
	Value     string   `yaml:"value,omitempty" json:"value,omitempty"`
	Parameter Argument `yaml:"parameter,omitempty" json:"parameter,omitempty"` // Note: webhook's match.parameter is singular
	IPRange   string   `yaml:"ip-range,omitempty" json:"ip-range,omitempty"`
}

type Hook struct {
	ID                                  string          `yaml:"id,omitempty" json:"id,omitempty"`
	ExecuteCommand                      string          `yaml:"execute-command,omitempty" json:"execute-command,omitempty"`
	CommandWorkingDirectory             string          `yaml:"command-working-directory,omitempty" json:"command-working-directory,omitempty"`
	ResponseMessage                     string          `yaml:"response-message,omitempty" json:"response-message,omitempty"`
	ResponseHeaders                     ResponseHeaders `yaml:"response-headers,omitempty" json:"response-headers,omitempty"`
	CaptureCommandOutput                bool            `yaml:"include-command-output-in-response,omitempty" json:"include-command-output-in-response,omitempty"`
	StreamCommandOutput                 bool            `yaml:"stream-command-output,omitempty" json:"stream-command-output,omitempty"`
	CaptureCommandOutputOnError         bool            `yaml:"include-command-output-in-response-on-error,omitempty" json:"include-command-output-in-response-on-error,omitempty"`
	PassEnvironmentToCommand            []Argument      `yaml:"pass-environment-to-command,omitempty" json:"pass-environment-to-command,omitempty"`
	PassArgumentsToCommand              []Argument      `yaml:"pass-arguments-to-command,omitempty" json:"pass-arguments-to-command,omitempty"`
	PassFileToCommand                   []Argument      `yaml:"pass-file-to-command,omitempty" json:"pass-file-to-command,omitempty"`
	JSONStringParameters                []Argument      `yaml:"parse-parameters-as-json,omitempty" json:"parse-parameters-as-json,omitempty"`
	TriggerRule                         *Rules          `yaml:"trigger-rule,omitempty" json:"trigger-rule,omitempty"`
	TriggerRuleMismatchHttpResponseCode int             `yaml:"trigger-rule-mismatch-http-response-code,omitempty" json:"trigger-rule-mismatch-http-response-code,omitempty"`
	TriggerSignatureSoftFailures        bool            `yaml:"trigger-signature-soft-failures,omitempty" json:"trigger-signature-soft-failures,omitempty"`
	IncomingPayloadContentType          string          `yaml:"incoming-payload-content-type,omitempty" json:"incoming-payload-content-type,omitempty"`
	SuccessHttpResponseCode             int             `yaml:"success-http-response-code,omitempty" json:"success-http-response-code,omitempty"`
	HTTPMethods                         []string        `yaml:"http-methods,omitempty" json:"http-methods,omitempty"`
}

// getEnvStr 从环境变量获取字符串，如果不存在则返回默认值
func getEnvStr(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// 定义命令行参数
	title := flag.String("title", "Webhook Configuration", "Title for the configuration UI")
	editUrl := flag.String("edit", "/edit_form", "URL for the edit configuration form")
	uploadUrl := flag.String("upload", "/upload_form", "URL for the upload configuration form")

	flag.Parse()

	configFilePath := getEnvStr("HOOKS", "/app/hooks.yaml")
	// 1. 读取 webhook 配置文件
	data, err := os.ReadFile(configFilePath)
	// 2. 解析 YAML 到 Go 结构体
	var config Config
	if err != nil {
		if os.IsNotExist(err) {
			config = []Hook{}
			fmt.Fprintf(os.Stderr, "Config file not found, starting with empty config.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			fmt.Print("<h1>Error: Could not load Webhook configuration.</h1><p>Please check server logs.</p>")
			os.Exit(1)
		}
	} else {
		// 2. 解析 YAML 到 Go 结构体
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshaling config: %v\n", err)
			fmt.Print("<h1>Error: Could not parse Webhook configuration.</h1><p>Invalid YAML format.</p>")
			os.Exit(1)
		}
	}

	// 定义要传递给模板的数据结构
	type TemplateData struct {
		Hooks     Config
		Title     string
		EditUrl   string
		UploadUrl string
		URLPrefix string // 确保 URLPrefix 被传递
	}

	prefix := getEnvStr("URL_PREFIX", "hooks")
	if prefix != "" {
		prefix = "/" + prefix
	}

	templateData := TemplateData{
		Hooks:     config,
		Title:     *title,
		EditUrl:   prefix + *editUrl,
		UploadUrl: prefix + *uploadUrl,
		URLPrefix: prefix, // 将 prefix 传递给模板
	}

	// 3. 定义 HTML 模板
	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary-color: #007bff; /* 蓝色 */
            --primary-dark: #0056b3;
            --secondary-color: #6c757d; /* 灰色 */
            --background-light: #f8f9fa; /* 浅背景 */
            --background-medium: #e9ecef; /* 中等背景 */
            --background-dark: #343a40; /* 深色背景 (用于代码块) */
            --text-primary: #212529; /* 主要文本色 */
            --text-secondary: #495057; /* 次要文本色 */
            --text-light: #f8f9fa; /* 浅色文本 (用于深色背景上的文本) */
            --code-text: #d63384; /* 代码高亮色 (粉色/洋红色) */
            --border-color: #dee2e6; /* 边框色 */
            --shadow-light: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
            --shadow-medium: 0 0.5rem 1rem rgba(0, 0, 0, 0.1);
            --accent-green: #28a745; /* 绿色 (可选用于成功提示) */
            --accent-red: #dc3545; /* 红色 (可选用于错误提示) */
        }

        body {
            font-family: 'Inter', 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: var(--background-light);
            color: var(--text-primary);
            line-height: 1.6;
            display: flex;
            justify-content: center;
            align-items: flex-start;
            min-height: 100vh;
        }
        .container {
            max-width: 1000px;
            width: 100%;
            margin: 20px auto;
            background-color: #ffffff;
            padding: 40px;
            border-radius: 12px;
            box-shadow: var(--shadow-medium);
            box-sizing: border-box;
        }
        h1 {
            text-align: center;
            color: var(--primary-color);
            margin-bottom: 30px;
            font-size: 2.5em;
            font-weight: 700;
            border-bottom: 3px solid var(--primary-color);
            padding-bottom: 20px;
            letter-spacing: -0.5px;
        }
        h2 {
            color: var(--text-primary);
            margin-top: 40px;
            margin-bottom: 25px;
            font-size: 1.8em;
            font-weight: 600;
            border-bottom: 1px solid var(--border-color);
            padding-bottom: 10px;
        }
        .hook-list {
            list-style: none;
            padding: 0;
        }
        .hook-item {
            background-color: #ffffff;
            border: 1px solid var(--border-color);
            margin-bottom: 20px;
            padding: 25px;
            border-radius: 10px;
            transition: all 0.2s ease-in-out;
            box-shadow: var(--shadow-light);
            display: flex;
            flex-direction: column;
            gap: 10px; /* Spacing between key-value pairs */
        }
        .hook-item:hover {
            transform: translateY(-5px);
            box-shadow: var(--shadow-medium);
        }
        .hook-item > div {
            display: flex;
            justify-content: space-between; /* 键左对齐，值右对齐 */
            align-items: flex-start;
            flex-wrap: wrap; /* Allow wrapping for long values */
            gap: 10px; /* 键和值之间的最小间距 */
        }
        .hook-item strong {
            color: var(--text-secondary); /* 统一键的颜色为次要文本色 */
            font-weight: 600;
            /* min-width: 280px; */ /* 不再需要固定最小宽度，让其自然宽度 */
            flex-shrink: 0;
            text-align: left; /* 键左对齐 */
            padding-right: 0; /* 移除右内边距 */
        }
        .hook-item code {
            background-color: var(--background-medium); /* 统一 code 标签的背景色 */
            padding: 4px 8px;
            border-radius: 5px;
            font-family: 'Fira Code', 'Cascadia Code', monospace;
            color: var(--code-text); /* 统一 code 标签的文本颜色 */
            font-size: 0.95em;
            word-break: break-all;
            flex-grow: 1; /* Allow code to take up remaining space */
            text-align: right; /* 值右对齐 */
        }
        .hook-item pre {
            display: block;
            margin-top: 10px;
            background-color: var(--background-dark);
            color: var(--text-light);
            padding: 15px;
            border-radius: 8px;
            overflow-x: auto;
            line-height: 1.5;
            white-space: pre-wrap;
            word-break: break-all;
            font-family: 'Fira Code', 'Cascadia Code', monospace;
            font-size: 0.9em;
            max-height: 200px; /* Limit height for long messages */
            flex-grow: 1; /* 让 pre 也弹性填充 */
            text-align: left; /* pre 保持左对齐 */
        }
        .actions-buttons {
            text-align: center;
            margin-top: 50px;
            margin-bottom: 40px;
            display: flex;
            justify-content: center;
            gap: 20px; /* Spacing between buttons */
        }
        .actions-buttons button {
            padding: 15px 35px;
            border: none;
            border-radius: 8px;
            background-color: var(--primary-color);
            color: white;
            font-size: 1.2em;
            cursor: pointer;
            transition: all 0.3s ease;
            box-shadow: 0 4px 12px rgba(0,123,255,0.25);
            font-weight: 500;
        }
        .actions-buttons button:hover {
            background-color: var(--primary-dark);
            transform: translateY(-3px);
            box-shadow: 0 6px 16px rgba(0,123,255,0.35);
        }
        .actions-buttons button:active {
            transform: translateY(0);
            box-shadow: none;
        }
        .hint {
            text-align: center;
            color: var(--secondary-color);
            margin-top: 40px;
            font-size: 0.95em;
            padding: 20px;
            border: 1px solid var(--border-color);
            border-radius: 10px;
            background-color: #e6f7ff; /* Lighter blue for hint */
            box-shadow: var(--shadow-light);
        }
        .hint strong {
            color: var(--primary-dark);
        }

        /* Nested list for parameters and headers - 保持不变 */
        .nested-list {
            margin: 5px 0 5px 0px; 
            list-style: none; 
            padding-left: 0;
            font-size: 0.9em;
            border-left: 2px solid var(--border-color); 
            padding-left: 15px; 
            width: 100%; 
        }
        .nested-list li {
            margin-bottom: 5px;
            border: none;
            padding: 0;
            background: none;
            box-shadow: none;
            line-height: 1.4;
            display: flex; 
            align-items: flex-start;
        }
        .nested-list li .nested-param-item {
            display: flex; 
            align-items: flex-start;
            margin-bottom: 2px;
            word-break: break-all;
            color: var(--text-primary); 
            flex-grow: 1; 
        }
        .nested-list li .nested-param-item strong {
            color: var(--text-secondary); 
            min-width: 120px; 
            margin-right: 5px;
            font-weight: 500;
            text-align: right; 
            padding-right: 5px;
        }
        .nested-list li .nested-param-item code {
            background-color: var(--background-medium);
            color: var(--code-text); 
            padding: 3px 6px;
            border-radius: 4px;
            font-family: 'Fira Code', 'Cascadia Code', monospace;
            font-size: 0.9em;
            display: inline-block;
            flex-grow: 1; 
        }

        /* Special styling for trigger-rule match content - 保持不变 */
        .trigger-rule-match-detail {
            display: flex;
            flex-wrap: wrap;
            gap: 10px; 
            margin-top: 5px;
            width: 100%; 
        }
        .trigger-rule-match-detail span {
            background-color: var(--background-medium); 
            padding: 5px 10px;
            border-radius: 5px;
            font-family: 'Fira Code', 'Cascadia Code', monospace;
            font-size: 0.95em;
            color: var(--code-text); 
            white-space: nowrap; 
            display: flex; 
            align-items: baseline;
            flex-grow: 1; 
        }
        .trigger-rule-match-detail span strong {
            color: var(--text-secondary); 
            margin-right: 5px;
            font-weight: 500;
            flex-shrink: 0; 
        }
        .no-hooks-message {
            text-align: center;
            color: var(--secondary-color);
            margin-top: 50px;
            padding: 20px;
            background-color: var(--background-medium); 
            border-radius: 8px;
            font-size: 1.1em;
            font-style: italic;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{ .Title }} Overview</h1>

        <div class="actions-buttons">
            <button onclick="location.href='{{ .EditUrl }}'">修改配置文件</button>
            <button onclick="location.href='{{ .UploadUrl }}'">上传可执行文件</button>
        </div>

        <h2>Current Hooks</h2>
        {{ if .Hooks }}
            <ul class="hook-list">
                {{ range .Hooks }}
                <li class="hook-item">
                    <div><strong>ID:</strong> <code>{{ .ID }}</code></div>
                    <div><strong>Execute Command:</strong> <code>{{ .ExecuteCommand }}</code></div>
                    {{ if .CommandWorkingDirectory }}
                    <div><strong>Command Working Directory:</strong> <code>{{ .CommandWorkingDirectory }}</code></div>
                    {{ end }}
                    {{ if .ResponseMessage }}
                    <div><strong>Response Message:</strong> <pre>{{ .ResponseMessage }}</pre></div>
                    {{ end }}
                    {{ if .ResponseHeaders }}
                    <div>
                        <strong>Response Headers:</strong>
                        <ul class="nested-list">
                            {{ range .ResponseHeaders }}
                            <li><strong>{{ .Name }}:</strong> <code>{{ .Value }}</code></li>
                            {{ end }}
                        </ul>
                    </div>
                    {{ end }}
                    {{ if .CaptureCommandOutput }}
                    <div><strong>Include Command Output in Response:</strong> <code>{{ .CaptureCommandOutput }}</code></div>
                    {{ end }}
                    {{ if .StreamCommandOutput }}
                    <div><strong>Stream Command Output:</strong> <code>{{ .StreamCommandOutput }}</code></div>
                    {{ end }}
                    {{ if .CaptureCommandOutputOnError }}
                    <div><strong>Include Command Output on Error:</strong> <code>{{ .CaptureCommandOutputOnError }}</code></div>
                    {{ end }}

                    {{ if .PassEnvironmentToCommand }}
                    <div>
                        <strong>Pass Environment to Command:</strong>
                        <ul class="nested-list">
                            {{ range .PassEnvironmentToCommand }}
                            <li>
                                {{ if .Source }}<div class="nested-param-item"><strong>source:</strong> <code>{{ .Source }}</code></div>{{ end }}
                                {{ if .Name }}<div class="nested-param-item"><strong>name:</strong> <code>{{ .Name }}</code></div>{{ end }}
                                {{ if .EnvName }}<div class="nested-param-item"><strong>envname:</strong> <code>{{ .EnvName }}</code></div>{{ end }}
                                {{ if .Base64Decode }}<div class="nested-param-item"><strong>base64decode:</strong> <code>{{ .Base64Decode }}</code></div>{{ end }}
                            </li>
                            {{ end }}
                        </ul>
                    </div>
                    {{ end }}

                    {{ if .PassArgumentsToCommand }}
                    <div>
                        <strong>Pass Arguments to Command:</strong>
                        <ul class="nested-list">
                            {{ range .PassArgumentsToCommand }}
                            <li>
                                {{ if .Source }}<div class="nested-param-item"><strong>source:</strong> <code>{{ .Source }}</code></div>{{ end }}
                                {{ if .Name }}<div class="nested-param-item"><strong>name:</strong> <code>{{ .Name }}</code></div>{{ end }}
                                {{ if .EnvName }}<div class="nested-param-item"><strong>envname:</strong> <code>{{ .EnvName }}</code></div>{{ end }}
                                {{ if .Base64Decode }}<div class="nested-param-item"><strong>base64decode:</strong> <code>{{ .Base64Decode }}</code></div>{{ end }}
                            </li>
                            {{ end }}
                        </ul>
                    </div>
                    {{ end }}

                    {{ if .PassFileToCommand }}
                    <div>
                        <strong>Pass File to Command:</strong>
                        <ul class="nested-list">
                            {{ range .PassFileToCommand }}
                            <li>
                                {{ if .Source }}<div class="nested-param-item"><strong>source:</strong> <code>{{ .Source }}</code></div>{{ end }}
                                {{ if .Name }}<div class="nested-param-item"><strong>name:</strong> <code>{{ .Name }}</code></div>{{ end }}
                                {{ if .EnvName }}<div class="nested-param-item"><strong>envname:</strong> <code>{{ .EnvName }}</code></div>{{ end }}
                                {{ if .Base64Decode }}<div class="nested-param-item"><strong>base64decode:</strong> <code>{{ .Base64Decode }}</code></div>{{ end }}
                            </li>
                            {{ end }}
                        </ul>
                    </div>
                    {{ end }}

                    {{ if .JSONStringParameters }}
                    <div>
                        <strong>Parse Parameters as JSON:</strong>
                        <ul class="nested-list">
                            {{ range .JSONStringParameters }}
                            <li>
                                {{ if .Source }}<div class="nested-param-item"><strong>source:</strong> <code>{{ .Source }}</code></div>{{ end }}
                                {{ if .Name }}<div class="nested-param-item"><strong>name:</strong> <code>{{ .Name }}</code></div>{{ end }}
                                {{ if .EnvName }}<div class="nested-param-item"><strong>envname:</strong> <code>{{ .EnvName }}</code></div>{{ end }}
                                {{ if .Base64Decode }}<div class="nested-param-item"><strong>base64decode:</strong> <code>{{ .Base64Decode }}</code></div>{{ end }}
                            </li>
                            {{ end }}
                        </ul>
                    </div>
                    {{ end }}

                    {{ if .TriggerRule }}
                    <div>
                        <strong>Trigger Rule:</strong>
                        <ul class="nested-list">
                            {{ if .TriggerRule.And }}<li><strong>and:</strong> <code>(complex rule, see YAML for details)</code></li>{{ end }}
                            {{ if .TriggerRule.Or }}<li><strong>or:</strong> <code>(complex rule, see YAML for details)</code></li>{{ end }}
                            {{ if .TriggerRule.Not }}<li><strong>not:</strong> <code>(complex rule, see YAML for details)</code></li>{{ end }}
                            {{ if .TriggerRule.Match }}
                            <li>
                                <strong>match:</strong>
                                <div class="trigger-rule-match-detail">
                                    {{ if .TriggerRule.Match.Type }}<span><strong>type:</strong> <code>{{ .TriggerRule.Match.Type }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.Regex }}<span><strong>regex:</strong> <code>{{ .TriggerRule.Match.Regex }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.Secret }}<span><strong>secret:</strong> <code>{{ .TriggerRule.Match.Secret }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.Value }}<span><strong>value:</strong> <code>{{ .TriggerRule.Match.Value }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.Parameter.Source }}<span><strong>parameter source:</strong> <code>{{ .TriggerRule.Match.Parameter.Source }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.Parameter.Name }}<span><strong>parameter name:</strong> <code>{{ .TriggerRule.Match.Parameter.Name }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.Parameter.EnvName }}<span><strong>parameter envname:</strong> <code>{{ .TriggerRule.Match.Parameter.EnvName }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.Parameter.Base64Decode }}<span><strong>parameter base64decode:</strong> <code>{{ .TriggerRule.Match.Parameter.Base64Decode }}</code></span>{{ end }}
                                    {{ if .TriggerRule.Match.IPRange }}<span><strong>IP range:</strong> <code>{{ .TriggerRule.Match.IPRange }}</code></span>{{ end }}
                                </div>
                            </li>
                            {{ end }}
                        </ul>
                    </div>
                    {{ end }}

                    {{ if .TriggerRuleMismatchHttpResponseCode }}
                    <div><strong>Trigger Rule Mismatch HTTP Response Code:</strong> <code>{{ .TriggerRuleMismatchHttpResponseCode }}</code></div>
                    {{ end }}
                    {{ if .TriggerSignatureSoftFailures }}
                    <div><strong>Trigger Signature Soft Failures:</strong> <code>{{ .TriggerSignatureSoftFailures }}</code></div>
                    {{ end }}
                    {{ if .IncomingPayloadContentType }}
                    <div><strong>Incoming Payload Content Type:</strong> <code>{{ .IncomingPayloadContentType }}</code></div>
                    {{ end }}
                    {{ if .SuccessHttpResponseCode }}
                    <div><strong>Success HTTP Response Code:</strong> <code>{{ .SuccessHttpResponseCode }}</code></div>
                    {{ end }}
                    {{ if .HTTPMethods }}
                    <div><strong>HTTP Methods:</strong> <code>{{ range .HTTPMethods }}{{ . }} {{ end }}</code></div>
                    {{ end }}
                </li>
                {{ end }}
            </ul>
        {{ else }}
            <p class="no-hooks-message">当前没有配置任何 Webhook 接口。</p>
        {{ end }}

        <p class="hint">
            此页面展示当前 Webhook 的配置。
            "修改配置文件" 和 "上传可执行文件" 按钮将引导您至相应的操作页面。
            这些操作会通过 POST 请求触发独立的 Go 二进制文件来处理。
        </p>
    </div>
</body>
</html>
`

	tmpl, err := template.New("webhook").Parse(htmlTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML template: %v\n", err)
		fmt.Print("<h1>Error: Internal UI rendering issue.</h1>")
		os.Exit(1)
	}

	// 4. 将解析后的数据渲染到模板，并输出到标准输出
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		fmt.Print("<h1>Error: Internal UI rendering issue.</h1>")
		os.Exit(1)
	}

	// 将生成的 HTML 写入标准输出，webhook 会捕获它
	fmt.Print(buf.String())
}
