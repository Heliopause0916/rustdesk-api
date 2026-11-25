package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/web"
	"net/http"
	"strings"
)

func WebInit(g *gin.Engine) {
	i := &web.Index{}
	g.GET("/", i.Index)

	if global.Config.App.WebClient == 1 {
		g.GET("/webclient-config/index.js", i.ConfigJs)
	}

	if global.Config.App.WebClient == 1 {
		g.GET("/webclient2", i.WebClient2)
		g.StaticFS("/webclient", http.Dir(global.Config.Gin.ResourcesPath+"/web"))
		g.Any("/webclient2/*filepath", webclient2Handler(i))
	}
	g.StaticFS("/_admin", http.Dir(global.Config.Gin.ResourcesPath+"/admin"))
}

// webclient2Handler 服务 /webclient2/*filepath：空路径、仅尾斜杠与
// index.html（含大小写变体）返回渲染过 {{CUSTOM_CONFIG}} 的页面，
// 其余路径回退为静态文件服务，避免通配路由返回未渲染的原始 index.html。
func webclient2Handler(i *web.Index) gin.HandlerFunc {
	fs := http.StripPrefix("/webclient2", http.FileServer(http.Dir(global.Config.Gin.ResourcesPath+"/web2")))
	return func(c *gin.Context) {
		p := c.Param("filepath")
		if p == "" || p == "/" || strings.EqualFold(p, "/index.html") {
			i.WebClient2(c)
			return
		}
		fs.ServeHTTP(c.Writer, c.Request)
	}
}
