package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/ai"
	"smartg-odin/internal/common/config"
	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/database"
	"smartg-odin/internal/handler"
	"smartg-odin/internal/middleware"
	"smartg-odin/internal/seed"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[FATAL] load config failed: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	defer logger.Sync()

	// 注入 AI 配置（默认关闭，零值即确定性规则路径）
	ai.Configure(cfg.AI)

	// 初始化数据库
	if err := database.InitWithDriver(cfg.Database.Driver, cfg.Database.DSN); err != nil {
		logger.L.Fatalf("[main] init database failed: %v", err)
	}

	// Seed 数据（仅当表为空时写入）
	seed.Run(database.GetDB())

	// 启动时刷新所有 active 数据源的 schema 缓存（失败仅告警，不中断启动）
	handler.RefreshAllSchemas()

	// 设置 Gin 模式
	mode := cfg.Server.Mode
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.TraceID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// 路由注册
	v1 := r.Group("/api/v1")
	{
		// Health
		v1.GET("/health", handler.HealthCheck)
		v1.GET("/ready", handler.ReadyCheck)

		// Workspace (deprecated，兼容旧前端)
		v1.GET("/workspace", handler.GetWorkspace)
		v1.PUT("/workspace", handler.SaveWorkspace)

		// Datasources
		ds := v1.Group("/datasources")
		{
			ds.GET("", handler.ListDatasources)
			ds.POST("", handler.CreateDatasource)
			ds.GET("/:id", handler.GetDatasource)
			ds.PUT("/:id", handler.UpdateDatasource)
			ds.DELETE("/:id", handler.DeleteDatasource)
			ds.POST("/:id/test", handler.TestDatasource)
			ds.GET("/:id/schema", handler.GetDatasourceSchema)
			ds.POST("/:id/schema/refresh", handler.RefreshDatasourceSchema)
		}

		// Ontologies
		ont := v1.Group("/ontologies")
		{
			ont.GET("", handler.ListOntologies)
			ont.POST("", handler.CreateOntology)
			ont.GET("/:id", handler.GetOntology)
			ont.PUT("/:id", handler.UpdateOntology)
			ont.DELETE("/:id", handler.DeleteOntology)
			ont.GET("/:id/vocabulary", handler.GetVocabulary)

			// Classes
			ont.GET("/:id/classes", handler.ListClasses)
			ont.POST("/:id/classes", handler.CreateClass)
			ont.PUT("/:id/classes/:classId", handler.UpdateClass)
			ont.DELETE("/:id/classes/:classId", handler.DeleteClass)

			// Properties
			ont.GET("/:id/classes/:classId/properties", handler.ListProperties)
			ont.POST("/:id/classes/:classId/properties", handler.CreateProperty)
			ont.PUT("/:id/classes/:classId/properties/:propId", handler.UpdateProperty)
			ont.DELETE("/:id/classes/:classId/properties/:propId", handler.DeleteProperty)

			// Relations
			ont.GET("/:id/relations", handler.ListRelations)
			ont.POST("/:id/relations", handler.CreateRelation)
			ont.PUT("/:id/relations/:relId", handler.UpdateRelation)
			ont.DELETE("/:id/relations/:relId", handler.DeleteRelation)

			// Rules
			ont.GET("/:id/rules", handler.ListRules)
			ont.POST("/:id/rules", handler.CreateRule)
			ont.PUT("/:id/rules/:ruleId", handler.UpdateRule)
			ont.DELETE("/:id/rules/:ruleId", handler.DeleteRule)
		}

		// Mappings
		mp := v1.Group("/mappings")
		{
			mp.GET("", handler.ListMappings)
			mp.POST("", handler.CreateMapping)
			mp.POST("/suggest", handler.SuggestMappings)
			mp.GET("/:id", handler.GetMapping)
			mp.PUT("/:id", handler.UpdateMapping)
			mp.DELETE("/:id", handler.DeleteMapping)
			mp.POST("/:id/validate", handler.ValidateMapping)
		}

		// Query
		q := v1.Group("/query")
		{
			q.POST("", handler.ExecuteQuery)
			q.POST("/dry-run", handler.DryRunQuery)
			q.POST("/explain", handler.ExplainQuery)
			q.POST("/nl-parse", handler.NLParse)
			q.GET("/trace/:traceId", handler.GetQueryTrace)
			q.GET("/history", handler.GetQueryHistory)
			q.GET("/templates", handler.ListQueryTemplates)
			q.POST("/templates", handler.SaveQueryTemplate)
		}

		// Reason（推理解释，不执行 SQL）
		v1.POST("/reason/explain", handler.ReasonExplain)

		// Assistant（P3：SSE 对话 + 技能清单 + 会话历史）
		as := v1.Group("/assistant")
		{
			as.POST("/chat", handler.AssistantChat)
			as.GET("/skills", handler.AssistantSkills)
			as.GET("/sessions", handler.AssistantSessions)
			as.GET("/sessions/:id/messages", handler.AssistantSessionMessages)
		}

		// Drivers（连接器驱动管理）
		v1.GET("/drivers", handler.ListDrivers)
		v1.POST("/drivers/:name/upload", handler.UploadDriver)
	}

	// NoRoute 统一返回 JSON
	r.NoRoute(func(c *gin.Context) {
		middleware.Error(c, &errs.AppError{
			Code:    40400,
			Message: fmt.Sprintf("route not found: %s %s", c.Request.Method, c.Request.URL.Path),
		})
	})

	// 优雅关闭
	addr := cfg.Server.Addr()
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatalf("[main] listen failed: %v", err)
		}
	}()

	logger.L.Infof("ODIN server started on %s (mode=%s)", addr, mode)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.L.Errorf("[main] server forced to shutdown: %v", err)
	}

	// 关闭数据库连接
	if err := database.Close(); err != nil {
		logger.L.Errorf("[main] close database: %v", err)
	}

	logger.L.Info("Server exited gracefully")
}
