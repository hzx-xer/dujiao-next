package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// uploadCacheControl 让浏览器和 CDN 缓存上传文件 7 天，到期后凭 Last-Modified 校验。
// 不标记 immutable：同一路径的文件有可能被替换。
const uploadCacheControl = "public, max-age=604800"

// UploadStatic 给上传文件加安全头和缓存头。
//
// 所有上传文件一律禁止脚本：即使上传校验被绕过（如 telegram 场景不限扩展名、上游图片），
// 直接打开 /uploads/x.html 也不会在站点源下执行脚本；<img src> 引用不受影响。
// SVG 额外强制下载。
func UploadStatic() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "sandbox; script-src 'none'")
		c.Header("X-Content-Type-Options", "nosniff")
		if strings.EqualFold(path.Ext(c.Request.URL.Path), ".svg") {
			c.Header("Content-Disposition", "attachment")
		}
		c.Writer = &uploadCacheWriter{ResponseWriter: c.Writer}
	}
}

// uploadCacheWriter 只给成功的响应加缓存头：404 被缓存的话，文件补传后仍会长期打不开。
type uploadCacheWriter struct {
	gin.ResponseWriter
}

func (w *uploadCacheWriter) WriteHeader(code int) {
	if code == http.StatusOK || code == http.StatusPartialContent || code == http.StatusNotModified {
		w.Header().Set("Cache-Control", uploadCacheControl)
	}
	w.ResponseWriter.WriteHeader(code)
}
