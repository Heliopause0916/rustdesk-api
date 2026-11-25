package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
)

// TestWebInitWebClient2Routes 验证 /webclient2 精确路由与 catch-all
// 路由 /webclient2/*filepath 的渲染边界：精确路径、尾斜杠与
// index.html（含大小写变体）均返回渲染后的页面，其余路径提供静态
// 资源；relay-server 未配置时响应体不含该行（空值省略行为）。
func TestWebInitWebClient2Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// resources/web2 为 gitignore 的 vendor 产物，新克隆仓库可能缺失。
	if _, err := os.Stat("../../resources/web2/index.html"); err != nil {
		t.Skipf("前置条件不满足：resources/web2 未 vendor（%v），需先放置 webclient2 静态文件", err)
	}

	origWebClient := global.Config.App.WebClient
	origResourcesPath := global.Config.Gin.ResourcesPath
	origRustdesk := global.Config.Rustdesk
	t.Cleanup(func() {
		global.Config.App.WebClient = origWebClient
		global.Config.Gin.ResourcesPath = origResourcesPath
		global.Config.Rustdesk = origRustdesk
	})

	global.Config.App.WebClient = 1
	global.Config.Gin.ResourcesPath = "../../resources"
	global.Config.Rustdesk.IdServer = "192.168.1.66:21116"
	global.Config.Rustdesk.ApiServer = "http://127.0.0.1:21114"
	global.Config.Rustdesk.RelayServer = ""   // 锁定空值省略行为
	global.Config.Rustdesk.Key = "test-key\n" // key-file 内容常带尾随换行，输出前需 trim

	g := gin.New()
	WebInit(g)

	for _, path := range []string{"/webclient2", "/webclient2/", "/webclient2/index.html", "/webclient2/INDEX.html"} {
		w := httptest.NewRecorder()
		g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, w.Code, http.StatusOK)
		}
		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Fatalf("GET %s content-type = %q, want text/html", path, ct)
		}
		body := w.Body.String()
		if strings.Contains(body, "{{CUSTOM_CONFIG}}") {
			t.Fatalf("GET %s body still contains {{CUSTOM_CONFIG}}", path)
		}
		for _, want := range []string{
			"[default-settings]",
			"custom-rendezvous-server = 192.168.1.66:21116",
			"api-server = http://127.0.0.1:21114",
			"key = test-key",
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("GET %s body missing %q", path, want)
			}
		}
		if strings.Contains(body, "key = test-key\n") {
			t.Fatalf("GET %s body contains untrimmed key value", path)
		}
		// relay-server 未配置时响应体不得包含该行。
		if strings.Contains(body, "relay-server") {
			t.Fatalf("GET %s body contains relay-server though relay-server is unset", path)
		}
	}

	// catch-all 其余路径继续提供静态资源。
	w2 := httptest.NewRecorder()
	g.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/webclient2/main.dart.js", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("GET /webclient2/main.dart.js status = %d, want %d", w2.Code, http.StatusOK)
	}
}
