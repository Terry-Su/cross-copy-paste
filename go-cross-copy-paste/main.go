package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
)

func main() {
	filePathPointer := flag.String("filepath", "", "共享文件路径")
	intervalPointer := flag.Int64("interval", 200, "检测时间间隔（毫秒）")
	flag.Parse()

	filePath := *filePathPointer
	interval := *intervalPointer

	fmt.Println("filepath", filePath)

	if filePath == "" || !fileExists(filePath) {
		panic("非法文件路径")
	}

	initialized := false
	lastClipboardText := ""
	lastShareFileText := ""

	for {
		clipboardTextChanged := false
		fileTextChanged := false

		fileContent, err := readFile(filePath)
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("无法读取文件内容: %v\n", err)
			// 有可能是远程文件，持续无法连接，增加轮询实践
			time.Sleep(5 * time.Millisecond)
			continue
		}

		clipboardContent, err := clipboard.ReadAll()
		if err != nil {
			fmt.Printf("无法读取剪贴板内容: %v\n", err)
			time.Sleep(1 * time.Second)
			// continue
			clipboardTextChanged = false
		}

		// 初始化
		if initialized {
			initialized = true
			lastShareFileText = fileContent
			lastClipboardText = clipboardContent
		}

		// 监听剪切版内容
		if clipboardContent != lastClipboardText {
			clipboardTextChanged = true
		}

		// 监听共享文本文件内容
		if lastShareFileText != fileContent {
			fileTextChanged = true
		}

		// 如果同时检测到剪切板和共享文件文本变化，优先用剪切板文本
		if clipboardTextChanged {
			lastClipboardText = clipboardContent
			if clipboardContent != fileContent {
				err := overwriteFile(filePath, clipboardContent)
				if err != nil {
					fmt.Printf("更新文件失败: %v\n", err)
				} else {
					lastShareFileText = clipboardContent
					fmt.Printf("文件已更新为剪贴板内容: %s\n", clipboardContent)
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
			continue
		}
		if fileTextChanged {
			lastShareFileText = fileContent
			if fileContent != clipboardContent {
				err := clipboard.WriteAll(fileContent)
				if err != nil {
					fmt.Printf("更新剪贴板失败: %v\n", err)
				} else {
					lastClipboardText = fileContent
					fmt.Printf("剪贴板已更新为文件内容: %s\n", fileContent)
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
			continue
		}

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}

// readFile 读取指定文件的内容
func readFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// overwriteFile 覆盖写入内容到指定文件
func overwriteFile(filePath, text string) error {
	// 打开文件（以写模式覆盖内容，不存在则创建）
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// 写入内容
	if _, err := f.WriteString(text); err != nil {
		return err
	}
	return nil
}

// fileExists 检测文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false // 文件不存在
	}
	return err == nil // 文件存在且无其他错误
}
