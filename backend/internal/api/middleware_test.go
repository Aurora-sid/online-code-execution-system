package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORSMiddleware_AllowAll(t *testing.T) {
	origins := []string{"*"}

	router := gin.New()
	router.Use(CORSMiddleware(origins))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Test: 请求带 Origin 头
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际: %d", w.Code)
	}

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("期望 Allow-Origin 为请求源, 实际: %s", allowOrigin)
	}
}

func TestCORSMiddleware_SpecificOrigins(t *testing.T) {
	origins := []string{"http://example.com", "http://localhost:5173"}

	router := gin.New()
	router.Use(CORSMiddleware(origins))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Test 1: 白名单内的源
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:5173" {
		t.Errorf("白名单源期望匹配, 实际: %s", allowOrigin)
	}

	// Test 2: 非白名单的源
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Origin", "http://malicious.com")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	allowOrigin2 := w2.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin2 != "http://example.com" {
		t.Errorf("非白名单源期望返回第一个配置源, 实际: %s", allowOrigin2)
	}
}

func TestCORSMiddleware_OptionsRequest(t *testing.T) {
	origins := []string{"*"}

	router := gin.New()
	router.Use(CORSMiddleware(origins))
	router.OPTIONS("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS 请求期望状态码 204, 实际: %d", w.Code)
	}
}

func TestIsOriginAllowed(t *testing.T) {
	// 测试设置 CORS 配置
	SetCORSConfig([]string{"http://example.com", "http://localhost:5173"})

	tests := []struct {
		origin   string
		expected bool
	}{
		{"http://example.com", true},
		{"http://localhost:5173", true},
		{"http://malicious.com", false},
		{"", true}, // 空 origin 允许
	}

	for _, tt := range tests {
		result := IsOriginAllowed(tt.origin)
		if result != tt.expected {
			t.Errorf("IsOriginAllowed(%q) = %v, 期望 %v", tt.origin, result, tt.expected)
		}
	}
}

func TestIsOriginAllowed_Wildcard(t *testing.T) {
	SetCORSConfig([]string{"*"})

	if !IsOriginAllowed("http://any-origin.com") {
		t.Error("配置 * 时应允许任何来源")
	}
}
