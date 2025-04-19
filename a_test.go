package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 尝试使用不同编码进行解码
func tryDecodeWithEncoding(data []byte, decoder transform.Transformer) (string, error) {
	reader := transform.NewReader(bytes.NewReader(data), decoder)
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// 检测可能的编码
func detectEncoding(data []byte) {
	// 尝试 GBK 解码
	if result, err := tryDecodeWithEncoding(data, simplifiedchinese.GBK.NewDecoder()); err == nil {
		fmt.Printf("可能是 GBK 编码: %s\n", result)
	}

	// 尝试 GB18030 解码
	if result, err := tryDecodeWithEncoding(data, simplifiedchinese.GB18030.NewDecoder()); err == nil {
		fmt.Printf("可能是 GB18030 编码: %s\n", result)
	}

	// 尝试 UTF-8 解码（直接输出）
	fmt.Printf("原始 UTF-8 输出: %s\n", string(data))
}

func TestA(t *testing.T) {
	chaos := "凳蹦戎昆储。"
	body := []byte(chaos)

	fmt.Println("尝试检测编码...")
	detectEncoding(body)

	// 使用 GBK 解码
	reader := transform.NewReader(bytes.NewReader(body), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("解码失败: %v\n", err)
		return
	}

	fmt.Println("GBK 解码结果:", string(decoded))
}
