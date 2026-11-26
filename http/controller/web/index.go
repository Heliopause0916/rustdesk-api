package web

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
)

type Index struct {
}

func (i *Index) Index(c *gin.Context) {
	c.Redirect(302, "/_admin/")
}

// WebClient2 渲染 resources/web2/index.html，将 {{CUSTOM_CONFIG}} 占位符
// 替换为按 webclient2 前端约定的 INI 格式构造的自定义配置后返回。
// 前端 js/dist/index.js 中的 dD() 会读取 <script id="custom-config">
// 标签内容（先尝试 base64 解码，失败则按原文解析）。
func (i *Index) WebClient2(c *gin.Context) {
	b, err := os.ReadFile(global.Config.Gin.ResourcesPath + "/web2/index.html")
	if err != nil {
		panic(fmt.Errorf("read webclient2 index.html: %w", err))
	}
	s := string(b)
	if n := strings.Count(s, "{{CUSTOM_CONFIG}}"); n != 1 {
		panic(fmt.Errorf("webclient2 index.html: expected exactly 1 {{CUSTOM_CONFIG}} placeholder, got %d", n))
	}
	c.Data(200, "text/html; charset=utf-8", []byte(strings.Replace(s, "{{CUSTOM_CONFIG}}", customConfig(), 1)))
}

// customConfig 按 dD() 调研到的 INI 格式构造自定义配置：
// 节名 default-settings / override-settings，键名中的下划线会被前端替换为连字符。
// 值取自 conf/config.yaml 的 rustdesk 配置项；值为空的配置项不输出，
// 由前端回退到自身默认逻辑（如 api-server 缺省时按 id-server 端口-2 推导）。
func customConfig() string {
	rd := &global.Config.Rustdesk
	lines := []string{"[default-settings]"}
	for _, kv := range [...]struct{ key, val string }{
		{"custom-rendezvous-server", rd.IdServer},
		{"relay-server", rd.RelayServer},
		{"api-server", rd.ApiServer},
		{"key", rd.Key},
	} {
		if v := strings.TrimSpace(kv.val); v != "" {
			lines = append(lines, kv.key+" = "+v)
		}
	}
	return strings.Join(lines, "\n")
}
