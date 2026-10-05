package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
)

// driverRoot JDBC 驱动文件的存放根目录（相对进程工作目录）。
const driverRoot = "./data/drivers"

// maxDriverUploadBytes 驱动文件上传大小上限（200MB）。
const maxDriverUploadBytes = 200 << 20

// DriverInfo 连接器驱动信息。
// native 类型由 Go 原生驱动内置支持；jdbc 类型需上传驱动包后经 JDBC Agent 代理连接。
type DriverInfo struct {
	Name           string     `json:"name"`
	DisplayName    string     `json:"display_name"`
	Category       string     `json:"category"`
	DriverType     string     `json:"driver_type"` // native=Go原生驱动, jdbc=JDBC代理
	DriverLib      string     `json:"driver_lib"`
	Status         string     `json:"status"` // ready=可用, need_driver=需上传驱动
	Description    string     `json:"description"`
	ConfigRequired bool       `json:"config_required"`
	FileName       string     `json:"file_name,omitempty"`
	FileSize       int64      `json:"file_size,omitempty"`
	UploadedAt     *time.Time `json:"uploaded_at,omitempty"`
}

// driverMeta 持久化到 ./data/drivers/:name/meta.json 的上传元信息。
type driverMeta struct {
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

var (
	driverMu sync.RWMutex
	drivers  = map[string]*DriverInfo{}
	// driverOrder 保证列表接口返回顺序稳定。
	driverOrder = []string{"postgres", "mysql", "sqlite", "dm", "kingbase", "gaussdb"}
)

func init() {
	seedDrivers := []DriverInfo{
		{
			Name: "postgres", DisplayName: "PostgreSQL", Category: "rdbms",
			DriverType: "native", DriverLib: "pgx/v5 v5.7.4", Status: "ready",
			Description: "Go原生pgx驱动，无需额外配置", ConfigRequired: false,
		},
		{
			Name: "mysql", DisplayName: "MySQL", Category: "rdbms",
			DriverType: "native", DriverLib: "go-sql-driver/mysql v1.9.2", Status: "ready",
			Description: "Go原生MySQL驱动，无需额外配置", ConfigRequired: false,
		},
		{
			Name: "sqlite", DisplayName: "SQLite", Category: "rdbms",
			DriverType: "native", DriverLib: "mattn/go-sqlite3 v1.14.22", Status: "ready",
			Description: "Go原生SQLite驱动，无需额外配置", ConfigRequired: false,
		},
		{
			Name: "dm", DisplayName: "达梦 DM8", Category: "rdbms",
			DriverType: "jdbc", DriverLib: "DmJdbcDriver18.jar", Status: "need_driver",
			Description: "通过JDBC Agent代理连接，需上传达梦JDBC驱动", ConfigRequired: true,
		},
		{
			Name: "kingbase", DisplayName: "人大金仓 KingbaseES", Category: "rdbms",
			DriverType: "jdbc", DriverLib: "kingbase8-jdbc.jar", Status: "need_driver",
			Description: "通过JDBC Agent代理连接，需上传金仓JDBC驱动", ConfigRequired: true,
		},
		{
			Name: "gaussdb", DisplayName: "华为 GaussDB", Category: "rdbms",
			DriverType: "jdbc", DriverLib: "gaussdb-jdbc.jar", Status: "need_driver",
			Description: "通过JDBC Agent代理连接，需上传GaussDB JDBC驱动", ConfigRequired: true,
		},
	}
	for i := range seedDrivers {
		d := seedDrivers[i]
		drivers[d.Name] = &d
	}
	loadPersistedDriverMeta()
}

// loadPersistedDriverMeta 启动时扫描驱动目录，恢复已上传驱动的元信息与 ready 状态。
func loadPersistedDriverMeta() {
	entries, err := os.ReadDir(driverRoot)
	if err != nil {
		return // 目录不存在属正常情况
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		d, ok := drivers[name]
		if !ok || d.DriverType != "jdbc" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(driverRoot, name, "meta.json"))
		if err != nil {
			continue
		}
		var meta driverMeta
		if err := json.Unmarshal(raw, &meta); err != nil || meta.FileName == "" {
			continue
		}
		// 驱动文件仍在磁盘上才恢复为 ready
		if _, err := os.Stat(filepath.Join(driverRoot, name, meta.FileName)); err != nil {
			continue
		}
		uploaded := meta.UploadedAt
		d.FileName = meta.FileName
		d.FileSize = meta.FileSize
		d.UploadedAt = &uploaded
		d.Status = "ready"
	}
}

// ListDrivers GET /api/v1/drivers
// 返回系统支持的连接器及驱动信息。
func ListDrivers(c *gin.Context) {
	driverMu.RLock()
	list := make([]DriverInfo, 0, len(drivers))
	for _, name := range driverOrder {
		if d, ok := drivers[name]; ok {
			list = append(list, *d)
		}
	}
	// 兜底：不在固定顺序中的条目按名称排序追加
	extra := make([]DriverInfo, 0)
	known := make(map[string]bool, len(driverOrder))
	for _, name := range driverOrder {
		known[name] = true
	}
	for name, d := range drivers {
		if !known[name] {
			extra = append(extra, *d)
		}
	}
	driverMu.RUnlock()

	sort.Slice(extra, func(i, j int) bool { return extra[i].Name < extra[j].Name })
	list = append(list, extra...)
	ok(c, list)
}

// UploadDriver POST /api/v1/drivers/:name/upload
// 上传 JDBC 驱动文件（multipart/form-data, field="file"），仅接受 driver_type=jdbc 的连接器。
func UploadDriver(c *gin.Context) {
	name := strings.ToLower(strings.TrimSpace(c.Param("name")))
	if name == "" {
		fail(c, errs.ParamError("缺少路径参数: name"))
		return
	}

	driverMu.RLock()
	d, found := drivers[name]
	driverType := ""
	if found {
		driverType = d.DriverType
	}
	driverMu.RUnlock()

	if !found {
		fail(c, errs.NotFound("连接器不存在: "+name))
		return
	}
	if driverType != "jdbc" {
		fail(c, errs.ParamError(fmt.Sprintf("连接器 %s 使用Go原生驱动（%s），无需上传驱动文件", name, driverType)))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fail(c, errs.ParamError("未找到上传文件字段 file: "+err.Error()))
		return
	}
	if fileHeader.Size <= 0 {
		fail(c, errs.ParamError("上传文件为空"))
		return
	}
	if fileHeader.Size > maxDriverUploadBytes {
		fail(c, errs.ParamError("驱动文件超过大小上限 200MB"))
		return
	}

	// 仅取文件名部分，防止路径穿越
	fileName := filepath.Base(filepath.ToSlash(fileHeader.Filename))
	if fileName == "" || fileName == "." || fileName == "/" {
		fail(c, errs.ParamError("非法的文件名"))
		return
	}

	dir := filepath.Join(driverRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fail(c, errs.Internal("创建驱动目录失败: "+err.Error(), err))
		return
	}
	dest := filepath.Join(dir, fileName)
	if err := c.SaveUploadedFile(fileHeader, dest); err != nil {
		fail(c, errs.Internal("保存驱动文件失败: "+err.Error(), err))
		return
	}

	now := time.Now()
	meta := driverMeta{FileName: fileName, FileSize: fileHeader.Size, UploadedAt: now}
	if raw, err := json.MarshalIndent(meta, "", "  "); err == nil {
		if err := os.WriteFile(filepath.Join(dir, "meta.json"), raw, 0o644); err != nil {
			logger.Warnf("[driver] 写入驱动元信息失败 name=%s err=%v", name, err)
		}
	}

	driverMu.Lock()
	d.FileName = fileName
	d.FileSize = fileHeader.Size
	d.UploadedAt = &now
	d.Status = "ready"
	resp := gin.H{
		"name":        d.Name,
		"driver_type": d.DriverType,
		"file_name":   fileName,
		"file_size":   fileHeader.Size,
		"uploaded_at": now,
		"status":      d.Status,
	}
	driverMu.Unlock()

	logger.Infof("[driver] uploaded jdbc driver name=%s file=%s size=%d", name, fileName, fileHeader.Size)
	okMsg(c, resp, "驱动上传成功")
}
