package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template" // 导入 io/fs 包用于文件系统操作
	"os"            // 导入 path/filepath 用于处理路径
	"sort"          // 导入 sort 包用于排序文件列表
)

// getEnvStr 从环境变量获取字符串，如果不存在则返回默认值
func getEnvStr(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// FileInfo 结构体用于存储文件或目录的信息
type FileInfo struct {
	Name  string
	IsDir bool
}

func main() {
	var title string
	var homeUrl string
	var uploadActionURL string

	// 定义命令行参数
	flag.StringVar(&title, "title", "上传可执行文件", "UI page title")
	flag.StringVar(&homeUrl, "home", "/ui", "URL to return to the main UI")
	flag.StringVar(&uploadActionURL, "upload-submit", "/upload-submit", "URL for the file upload processing hook")

	flag.Parse()

	// 获取 UPLOAD_DEST_DIR
	uploadDestDir := getEnvStr("UPLOAD_DEST_DIR", "/etc/webhook/scripts/upload_destination/")

	// 读取目标目录内容
	var dirContents []FileInfo
	files, err := os.ReadDir(uploadDestDir)
	if err != nil {
		// 如果目录不存在或无法读取，记录错误但不阻止页面加载
		fmt.Fprintf(os.Stderr, "Error reading upload destination directory '%s': %v\n", uploadDestDir, err)
		// 可以选择在这里渲染一个包含错误信息的页面，或者让列表为空
	} else {
		for _, file := range files {
			dirContents = append(dirContents, FileInfo{
				Name:  file.Name(),
				IsDir: file.IsDir(),
			})
		}
		// 按名称排序，使显示更整齐
		sort.Slice(dirContents, func(i, j int) bool {
			// 目录优先，然后按名称排序
			if dirContents[i].IsDir != dirContents[j].IsDir {
				return dirContents[i].IsDir // 目录排在文件前面
			}
			return dirContents[i].Name < dirContents[j].Name
		})
	}

	// 定义要传递给模板的数据结构
	type TemplateData struct {
		Title           string
		HomeURL         string
		UploadActionURL string
		URLPrefix       string
		DestDirContents []FileInfo // 新增：目录内容列表
		DestDirPath     string     // 新增：目标目录路径
	}

	prefix := getEnvStr("URL_PREFIX", "hooks")
	if prefix != "" {
		prefix = "/" + prefix
	}

	templateData := TemplateData{
		Title:           title,
		HomeURL:         prefix + homeUrl,
		UploadActionURL: prefix + uploadActionURL,
		URLPrefix:       prefix,
		DestDirContents: dirContents,   // 传递目录内容
		DestDirPath:     uploadDestDir, // 传递目录路径，以便在页面显示
	}

	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{ .Title }}</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 20px; background-color: #f0f2f5; color: #333; line-height: 1.6; display: flex; flex-direction: column; justify-content: center; align-items: center; min-height: 100vh; text-align: center;}
        .container { max-width: 600px; margin: 20px auto; background-color: #ffffff; padding: 30px 40px; border-radius: 10px; box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
        h1 { color: #2c3e50; margin-bottom: 20px; font-size: 2em; }
        p { font-size: 1.1em; color: #555; margin-bottom: 25px; }
        .button-group { margin-top: 25px; }
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
            margin: 0 10px; /* Added margin for spacing */
        }
        button:hover {
            background-color: #0056b3;
            transform: translateY(-2px);
        }
        button:active {
            transform: translateY(0);
            box-shadow: none;
        }
        .file-input-container {
            margin-bottom: 20px;
        }
        input[type="file"] {
            padding: 8px;
            border: 1px solid #ccc;
            border-radius: 4px;
        }
        .dir-contents {
            margin-top: 30px;
            border-top: 1px dashed #ddd;
            padding-top: 20px;
            text-align: left;
        }
        .dir-contents h2 {
            color: #2c3e50;
            font-size: 1.5em;
            margin-bottom: 15px;
        }
        .dir-contents ul {
            list-style: none;
            padding: 0;
            margin: 0;
            max-height: 200px; /* 限制高度，可滚动 */
            overflow-y: auto;
            border: 1px solid #eee;
            border-radius: 5px;
            background-color: #fcfcfc;
            padding: 10px;
        }
        .dir-contents li {
            padding: 8px 0;
            border-bottom: 1px dotted #eee;
            color: #555;
            font-family: monospace; /* 等宽字体更适合显示文件路径 */
            font-size: 0.95em;
        }
        .dir-contents li:last-child {
            border-bottom: none;
        }
        .dir-contents li.directory {
            font-weight: bold;
            color: #007bff; /* 目录颜色 */
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{ .Title }}</h1>
        <form id="uploadForm">
            <div class="file-input-container">
                <label for="fileInput">选择文件：</label>
                <input type="file" id="fileInput" name="file" required>
            </div>
            <div class="button-group">
                <button type="submit">上传</button>
                <button type="button" onclick="location.href='{{ .HomeURL }}'">返回</button>
            </div>
        </form>
        <div id="response" style="margin-top: 20px; color: green;"></div>

        <div class="dir-contents">
            <h2>目录 "{{ .DestDirPath }}":</h2>
            {{ if .DestDirContents }}
            <ul>
                {{ range .DestDirContents }}
                    <li {{ if .IsDir }}class="directory"{{ end }}>
                        {{ if .IsDir }}📁 {{ else }}📄 {{ end }} {{ .Name }}
                    </li>
                {{ end }}
            </ul>
            {{ else }}
                <p>目录为空或无法读取目录内容。</p>
            {{ end }}
        </div>
    </div>

    <script>
        document.getElementById('uploadForm').addEventListener('submit', async function(event) {
            event.preventDefault();

            const fileInput = document.getElementById('fileInput');
            const file = fileInput.files[0];

            if (!file) {
                alert('请选择一个文件进行上传。');
                return;
            }

            const reader = new FileReader();
            reader.readAsDataURL(file);

            reader.onload = async function() {
                const base64String = reader.result.split(',')[1];
                const fileName = file.name;

                const payload = {
                    file_name: fileName,
                    file_content: base64String
                };

                try {
                    const uploadURL = '{{ .UploadActionURL }}';

                    const response = await fetch(uploadURL, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify(payload)
                    });

                    const responseText = await response.text();
                    document.open();
                    document.write(responseText);
                    document.close();

                } catch (error) {
                    console.error('上传过程中发生错误:', error);
                    document.getElementById('response').innerHTML = '<span style="color: red;">上传失败: ' + error.message + '</span>';
                }
            };

            reader.onerror = function(error) {
                console.error('文件读取错误:', error);
                alert('读取文件时发生错误。');
            };
        });
    </script>
</body>
</html>
`
	tmpl, err := template.New("uploadForm").Parse(htmlTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing HTML template for upload form: %v\n", err)
		fmt.Print("<h1>Error: Internal UI rendering issue for upload form.</h1>")
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template for upload form: %v\n", err)
		fmt.Print("<h1>Error: Internal UI rendering issue for upload form.</h1>")
		os.Exit(1)
	}

	fmt.Print(buf.String())
}
