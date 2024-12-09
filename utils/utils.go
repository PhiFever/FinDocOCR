package utils

import (
	"FinDocOCR/config"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/carlmjohnson/requests"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var logger = config.GetLogger()

func FindSuffixesInDir(dir string, fileSuffixes []string) ([]string, error) {
	var filePaths []string

	// filepath.Walk 遍历指定的目录
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查是否是文件
		if !info.IsDir() {
			// 获取文件扩展名并转换为小写
			ext := strings.ToLower(filepath.Ext(path))

			// 检查扩展名是否在目标序列中
			for _, suffix := range fileSuffixes {
				if suffix == ext {
					filePaths = append(filePaths, path)
					break
				}
			}
		}

		return nil
	})

	return filePaths, err
}

const (
	MaxFileSizeBytes = 4 * 1024 * 1024 // 4MB
	MinDimension     = 15              // 最短边至少15px
	MaxDimension     = 4096            // 最长边最大4096px
)

func ImageResize(imagePath string) ([]byte, error) {
	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %v", err)
	}

	// 处理PDF文件
	if strings.ToLower(filepath.Ext(imagePath)) == ".pdf" {
		return readPDF(imagePath)
	}

	// 读取原始图像
	img, err := imgio.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("打开图像失败: %v", err)
	}

	// 获取原始尺寸
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// 计算缩放后的尺寸
	newWidth, newHeight := calculateOptimalDimensions(originalWidth, originalHeight)

	// 调整图像尺寸
	resizedImg := transform.Resize(img, newWidth, newHeight, transform.Linear)

	// 渐进式压缩直到满足大小要求
	quality := 95
	var encodedSize int
	var buffer bytes.Buffer

	for quality >= 60 { // 最低质量限制
		buffer.Reset()

		// 根据原始图像格式选择编码方式
		switch filepath.Ext(strings.ToLower(imagePath)) {
		case ".png":
			if err := png.Encode(&buffer, resizedImg); err != nil {
				return nil, fmt.Errorf("PNG编码失败: %v", err)
			}
		default: // 默认使用JPEG
			if err := jpeg.Encode(&buffer, resizedImg, &jpeg.Options{Quality: quality}); err != nil {
				return nil, fmt.Errorf("JPEG编码失败: %v", err)
			}
		}

		// 计算base64和urlencode后的大小
		base64Str := base64.StdEncoding.EncodeToString(buffer.Bytes())
		urlEncodedStr := url.QueryEscape(base64Str)
		encodedSize = len(urlEncodedStr)

		if encodedSize <= MaxFileSizeBytes {
			break
		}

		// 如果超过大小限制，降低质量继续尝试
		quality -= 5
	}

	if encodedSize > MaxFileSizeBytes {
		return nil, fmt.Errorf("无法将图像压缩到4MB以下，当前大小: %d bytes", encodedSize)
	}

	return buffer.Bytes(), nil
}

// 读取PDF文件
func readPDF(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开PDF文件失败: %v", err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// 计算最优尺寸
func calculateOptimalDimensions(width, height int) (int, int) {
	// 确保最短边不小于最小尺寸
	if width < MinDimension || height < MinDimension {
		if width < height {
			ratio := float64(height) / float64(width)
			return MinDimension, int(float64(MinDimension) * ratio)
		}
		ratio := float64(width) / float64(height)
		return int(float64(MinDimension) * ratio), MinDimension
	}

	// 确保最长边不超过最大尺寸
	if width > MaxDimension || height > MaxDimension {
		if width > height {
			ratio := float64(height) / float64(width)
			return MaxDimension, int(float64(MaxDimension) * ratio)
		}
		ratio := float64(width) / float64(height)
		return int(float64(MaxDimension) * ratio), MaxDimension
	}

	return width, height
}

func GetMultipleInvoice(imageBytes []byte, accessToken string) ([]byte, error) {
	mimeType := http.DetectContentType(imageBytes)

	supportedTypes := map[string]string{
		"image/jpeg":      "image",
		"image/png":       "image",
		"application/pdf": "pdf_file",
	}

	paramKey, supported := supportedTypes[mimeType]
	if !supported {
		return nil, fmt.Errorf("不支持的类型：%s", mimeType)
	}

	params := url.Values{
		paramKey:           {base64.StdEncoding.EncodeToString(imageBytes)},
		"verify_parameter": {"false"},
		"probability":      {"false"},
		"location":         {"false"},
	}

	var result bytes.Buffer
	err := requests.URL("https://aip.baidubce.com/rest/2.0/ocr/v1/multiple_invoice").
		Param("access_token", accessToken).
		ContentType("application/x-www-form-urlencoded").
		Accept("application/json").
		BodyForm(params).
		ToBytesBuffer(&result).
		Fetch(context.Background())

	return result.Bytes(), err
}
