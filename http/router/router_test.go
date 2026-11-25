package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
)

// TestWebInitWebClient2Routes 验证 /webclient2 显式精确路由与
// StaticFS 通配路由 /webclient2/*filepath 并存时注册不 panic，
// 且显式路由按 INI 格式渲染自定义配置。
func TestWebInitWebClient2Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	global.Config.App.WebClient = 1
	global.Config.Gin.ResourcesPath = "../../resources"
	global.Config.Rustdesk.IdServer = "192.168.1.66:21116"
	global.Config.Rustdesk.ApiServer = "http://127.0.0.1:21114"
	global.Config.Rustdesk.Key = "test-key"

	g := gin.New()
	WebInit(g)

	w := httptest.NewRecorder()
	g.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/webclient2", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /webclient2 status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("GET /webclient2 content-type = %q, want text/html", ct)
	}
	body := w.Body.String()
	if strings.Contains(body, "{{CUSTOM_CONFIG}}") {
		t.Fatal("GET /webclient2 body still contains {{CUSTOM_CONFIG}}")
	}
	for _, want := range []string{
		"[default-settings]",
		"custom-rendezvous-server = 192.168.1.66:21116",
		"api-server = http://127.0.0.1:21114",
		"key = test-key",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("GET /webclient2 body missing %q", want)
		}
	}

	// StaticFS 通配路由继续提供静态资源。
	w2 := httptest.NewRecorder()
	g.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/webclient2/main.dart.js", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("GET /webclient2/main.dart.js status = %d, want %d", w2.Code, http.StatusOK)
	}
}
