package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// getEnvStr 从环境变量获取字符串，如果不存在则返回默认值
func getEnvStr(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// renderResponse 渲染 HTML 响应页面
func renderResponse(w io.Writer, title, message, homeUrl string) {
	fmt.Fprintf(w, `<!DOCTYPE html>
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
        .error-detail {
            display: block;
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
            padding: 10px;
            margin-top: 10px;
            border-radius: 5px;
            text-align: left;
            white-space: pre-wrap;
            word-break: break-all;
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
</html>`, title, title, message, homeUrl)
}

func main() {
	var homeUrl string
	var originalFilename string // 通过命令行参数接收原始文件名

	// flag.StringVar 声明命令行参数
	flag.StringVar(&homeUrl, "home", "/ui", "URL to return to after processing")
	flag.StringVar(&originalFilename, "file-name", "", "Original name of the uploaded file")
	flag.Parse() // 解析命令行参数

	// 从环境变量获取 upload_dest_dir 和 url_prefix
	uploadDestDir := getEnvStr("UPLOAD_DEST_DIR", "/etc/webhook/scripts/upload_destination/")
	prefix := getEnvStr("URL_PREFIX", "hooks")
	if prefix != "" {
		prefix = "/" + prefix
	}
	homeUrl = prefix + homeUrl // 确保 homeUrl 拼接正确，形如 /hooks/ui

	// 从环境变量获取 webhook 创建的临时文件路径
	uploadedFilePath := os.Getenv("UPLOADED_FILE_PATH")

	// 检查必要参数和环境变量
	if uploadedFilePath == "" {
		renderResponse(os.Stdout, "上传失败", "脚本：未接收到上传文件路径 (UPLOADED_FILE_PATH 环境变量未设置)。", homeUrl)
		os.Exit(1)
	}
	if originalFilename == "" {
		originalFilename = filepath.Base(uploadedFilePath)
		fmt.Fprintf(os.Stderr, "Warning: Original filename not provided, using '%s' from temporary path.\n", originalFilename)
	}

	// 确保目标目录存在
	err := os.MkdirAll(uploadDestDir, 0755) // 0755 权限：所有者读写执行，组和其他用户读和执行
	if err != nil {
		renderResponse(os.Stdout, "上传失败", fmt.Sprintf("脚本：无法创建目标目录 %s: %v", uploadDestDir, err), homeUrl)
		os.Exit(1)
	}

	// 构造最终的文件路径
	safeFilename := filepath.Base(originalFilename)
	destFilePath := filepath.Join(uploadDestDir, safeFilename)

	// 移动/复制 webhook 创建的临时文件到目标位置
	finalMoveSuccess := false
	err = os.Rename(uploadedFilePath, destFilePath)
	if err != nil {
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Op == "rename" {
			// Cross-device link, need to copy
			srcFile, openErr := os.Open(uploadedFilePath)
			if openErr != nil {
				renderResponse(os.Stdout, "上传失败", fmt.Sprintf("脚本：无法打开上传的临时文件 %s: %v", uploadedFilePath, openErr), homeUrl)
				os.Exit(1)
			}
			defer srcFile.Close()

			dstFile, createErr := os.Create(destFilePath)
			if createErr != nil {
				renderResponse(os.Stdout, "上传失败", fmt.Sprintf("脚本：无法创建目标文件 %s: %v", destFilePath, createErr), homeUrl)
				os.Exit(1)
			}
			defer dstFile.Close()

			_, copyErr := io.Copy(dstFile, srcFile)
			if copyErr != nil {
				renderResponse(os.Stdout, "上传失败", fmt.Sprintf("脚本：无法复制文件内容: %v", copyErr), homeUrl)
				os.Exit(1)
			}
			err = dstFile.Sync()
			if err != nil {
				renderResponse(os.Stdout, "上传失败", fmt.Sprintf("脚本：无法同步文件到磁盘: %v", err), homeUrl)
				os.Exit(1)
			}
			os.Remove(uploadedFilePath) // Remove temp file after successful copy
			finalMoveSuccess = true
		} else {
			renderResponse(os.Stdout, "上传失败", fmt.Sprintf("脚本：无法移动文件: %v", err), homeUrl)
			os.Exit(1)
		}
	} else {
		finalMoveSuccess = true // os.Rename succeeded
	}

	// 如果文件成功移动或复制，则赋予可执行权限
	if finalMoveSuccess {
		// 设置文件权限为 0755 (-rwxr-xr-x)
		// 确保文件所有者可以读、写、执行
		// 组用户和其他用户可以读、执行
		err = os.Chmod(destFilePath, 0755)
		if err != nil {
			// 权限设置失败不应阻止文件上传成功的报告，但应记录警告
			fmt.Fprintf(os.Stderr, "Warning: Could not set executable permissions on '%s': %v\n", destFilePath, err)
			renderResponse(os.Stdout, "上传成功 (有警告)", fmt.Sprintf("文件 '%s' 已成功上传到 %s，但无法设置可执行权限：%v", safeFilename, uploadDestDir, err), homeUrl)
			os.Exit(0) // 即使有警告，也视为成功上传
		}
	}

	renderResponse(os.Stdout, "上传成功", fmt.Sprintf("文件 '%s' 已成功上传到 %s", safeFilename, uploadDestDir), homeUrl)
	os.Exit(0)
}
