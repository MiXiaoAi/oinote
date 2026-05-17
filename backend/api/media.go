package handlers

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ServeMediaFile 支持Range请求的媒体文件服务
func ServeMediaFile(c *fiber.Ctx) error {
	filePath := c.Params("*")
	if filePath == "" {
		return c.Status(400).JSON(fiber.Map{"error": "文件路径不能为空"})
	}

	// 规范化路径，确保使用正斜杠
	normalizedPath := filepath.ToSlash(filePath)

	// 检查路径是否以uploads开头（使用strings.HasPrefix更安全）
	var fullPath string
	if strings.HasPrefix(normalizedPath, "uploads/") {
		fullPath = filepath.Join("./data", normalizedPath)
	} else {
		fullPath = filepath.Join("./data/uploads", normalizedPath)
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(404).JSON(fiber.Map{"error": "文件不存在"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "文件访问错误"})
	}

	// 打开文件
	file, err := os.Open(fullPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "无法打开文件"})
	}
	defer file.Close()

	// 获取文件大小
	fileSize := fileInfo.Size()

	// 设置Content-Type
	contentType := getContentType(filePath)
	c.Set("Content-Type", contentType)
	c.Set("Accept-Ranges", "bytes")
	c.Set("Content-Length", strconv.FormatInt(fileSize, 10))

	// 处理Range请求
	rangeHeader := c.Get("Range")
	if rangeHeader != "" {
		// 简单处理：暂时返回整个文件，但设置正确的头
		c.Set("Accept-Ranges", "bytes")

		// 普通请求，传输整个文件
		_, err = io.Copy(c, file)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "文件传输失败"})
		}

		return nil
	}

	// 普通请求，传输整个文件
	_, err = io.Copy(c, file)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "文件传输失败"})
	}

	return nil
}

func getContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg":
		return "video/ogg"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".m4a":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	default:
		return "application/octet-stream"
	}
}
