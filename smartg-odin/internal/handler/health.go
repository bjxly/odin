package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/database"
	"smartg-odin/internal/ds/connector"
	"smartg-odin/internal/middleware"
)

// HealthCheck GET /api/v1/health
// 返回服务存活状态与版本信息，不做任何外部依赖检查。
func HealthCheck(c *gin.Context) {
	ok(c, gin.H{
		"status":    "ok",
		"version":   Version,
		"go":        runtime.Version(),
		"timestamp": time.Now().UnixMilli(),
	})
}

// ReadyCheck GET /api/v1/ready
// 就绪探针：检查数据库连通性，失败时返回 HTTP 503。
func ReadyCheck(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		middleware.ErrorWithHTTP(c, http.StatusServiceUnavailable,
			errs.ServiceDown(errs.New(errs.CodeServiceDown, "database not initialized")))
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		middleware.ErrorWithHTTP(c, http.StatusServiceUnavailable, errs.ServiceDown(err))
		return
	}
	if err := sqlDB.Ping(); err != nil {
		middleware.ErrorWithHTTP(c, http.StatusServiceUnavailable, errs.ServiceDown(err))
		return
	}

	ok(c, gin.H{
		"status":                     "ready",
		"version":                    Version,
		"database":                   "connected",
		"supported_datasource_types": connector.SupportedTypes(),
	})
}
