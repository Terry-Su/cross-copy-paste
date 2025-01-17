![](https://terry-su.github.io/CDN/images/Snipaste_2025-01-17_22-34-01.webp)
## 背景
如今sunshine/moonlight串流技术越来越流行，让我们能够几乎无延迟丝滑从多端远程操作另外一台电脑（win/mac/ubuntu）（体验不是rdp能比的）。虽然串流主要应用场景是游戏，但用来办公体验也远优于普通远程桌面。但串流如果用来办公有个痛点是不支持双向复制粘贴（仅win to win支持单向粘贴文本）。
另外一个场景是windows自带强大虚拟机hyperv不开启增强模式也无法使用双向复制粘贴。hyperv win/ubuntu在一些使用场景无法使用增强模式，比如win显卡虚拟化，ubuntu稳60帧交互。

本项目旨在提供技术方案，目前有golang实现，可自行编译使用（使用更安心），也可按实现思路用自己想用的编程语言实现。

抛砖引玉，实现方法有很多，在此讲下我的思路、方案。

## 核心
使用共享文件（共享文本文件。目前图省事，只实现跨端复制粘贴文本）。

## 本质需求
pc1和pc2有一个共享文本文件，通过程序实现pc1和pc2双向同步剪切板。

## 方案
以pc1为例：每隔一段时间（比如200毫秒），分别判断：剪切板跟之前相比是否变化、文本文件跟之前相比是否变化。

若剪切板发生变化，则更新文本文件。

若文本文件发生变化，则更新剪切板。

pc2同理。

pc1和pc2同时运行程序即可实现双向复制粘贴文本。


## golang实现
```go
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
			// 有可能是远程文件，持续无法连接，增加轮询时间
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
```



