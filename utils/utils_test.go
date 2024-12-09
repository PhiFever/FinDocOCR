package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImageResize 是主测试函数，包含多个子测试场景
func TestImageResize(t *testing.T) {
	// 创建测试用的临时目录
	tempDir, err := os.MkdirTemp("", "image-resize-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试场景：正常尺寸的JPEG图像
	t.Run("Normal JPEG Image", func(t *testing.T) {
		// 创建测试图像
		imagePath := filepath.Join(tempDir, "normal.jpg")
		createTestImage(t, imagePath, 800, 600, "jpeg")

		// 执行图像处理
		result, err := ImageResize(imagePath)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证处理后的图像大小是否符合要求
		assert.True(t, isProcessedImageValid(result))
	})

	// 测试场景：特大尺寸的PNG图像
	t.Run("Oversized PNG Image", func(t *testing.T) {
		imagePath := filepath.Join(tempDir, "large.png")
		createTestImage(t, imagePath, 5000, 5000, "png")

		result, err := ImageResize(imagePath)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证最长边是否被正确限制在4096像素以内
		img, err := decodeImage(result)
		assert.NoError(t, err)
		bounds := img.Bounds()
		assert.LessOrEqual(t, bounds.Dx(), MaxDimension)
		assert.LessOrEqual(t, bounds.Dy(), MaxDimension)
	})

	// 测试场景：非常小的图像
	t.Run("Tiny Image", func(t *testing.T) {
		imagePath := filepath.Join(tempDir, "tiny.jpg")
		createTestImage(t, imagePath, 10, 10, "jpeg")

		result, err := ImageResize(imagePath)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证最短边是否被正确放大到至少15像素
		img, err := decodeImage(result)
		assert.NoError(t, err)
		bounds := img.Bounds()
		assert.GreaterOrEqual(t, bounds.Dx(), MinDimension)
		assert.GreaterOrEqual(t, bounds.Dy(), MinDimension)
	})

	// 测试场景：不存在的文件
	t.Run("Non-existent File", func(t *testing.T) {
		_, err := ImageResize("non_existent.jpg")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件不存在")
	})

	// 测试场景：PDF文件
	t.Run("PDF File", func(t *testing.T) {
		pdfPath := filepath.Join(tempDir, "test.pdf")
		createTestPDF(t, pdfPath)

		result, err := ImageResize(pdfPath)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	// 测试场景：验证base64和urlencode后的大小限制
	t.Run("Encoded Size Limit", func(t *testing.T) {
		imagePath := filepath.Join(tempDir, "large_quality.jpg")
		createTestImage(t, imagePath, 2000, 2000, "jpeg")

		result, err := ImageResize(imagePath)
		assert.NoError(t, err)

		// 验证base64和urlencode后的大小是否在4MB以内
		base64Str := base64.StdEncoding.EncodeToString(result)
		urlEncodedStr := url.QueryEscape(base64Str)
		assert.LessOrEqual(t, len(urlEncodedStr), MaxFileSizeBytes)
	})
}

// createTestImage 创建指定尺寸和格式的测试图像
func createTestImage(t *testing.T, path string, width, height int, format string) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充一些测试图案
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 100,
				A: 255,
			})
		}
	}

	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	switch format {
	case "jpeg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(file, img)
	}
	require.NoError(t, err)
}

// createTestPDF 创建测试用的PDF文件
func createTestPDF(t *testing.T, path string) {
	// 创建一个简单的PDF文件
	content := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF")
	err := os.WriteFile(path, content, 0644)
	require.NoError(t, err)
}

// decodeImage 解码图像数据
func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// isProcessedImageValid 验证处理后的图像是否符合所有要求
func isProcessedImageValid(imageData []byte) bool {
	// 检查是否能成功解码
	img, err := decodeImage(imageData)
	if err != nil {
		return false
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// 验证尺寸限制
	if width < MinDimension || height < MinDimension {
		return false
	}
	if width > MaxDimension || height > MaxDimension {
		return false
	}

	// 验证编码后的大小限制
	base64Str := base64.StdEncoding.EncodeToString(imageData)
	urlEncodedStr := url.QueryEscape(base64Str)
	if len(urlEncodedStr) > MaxFileSizeBytes {
		return false
	}

	return true
}
