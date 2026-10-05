package seed

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"smartg-odin/internal/common/logger"
)

// createSQLiteFiles 创建可供实际查询执行的 SQLite 演示库文件。
// 路径相对于进程工作目录（项目根 smartg-odin）。
func createSQLiteFiles() {
	if err := os.MkdirAll("./data", 0o755); err != nil {
		logger.L.Errorf("create data dir failed: %v", err)
		return
	}
	createWarehouseDB("./data/warehouse.db")
}

// openSQLite 打开（或创建）一个 SQLite 文件连接。
func openSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// exec 顺序执行多条语句。
func exec(db *sql.DB, stmts ...string) error {
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

// createWarehouseDB 构建 warehouse.db：customers / orders / products。
func createWarehouseDB(path string) {
	db, err := openSQLite(path)
	if err != nil {
		logger.L.Errorf("open warehouse.db failed: %v", err)
		return
	}
	defer db.Close()

	if err := exec(db,
		"DROP TABLE IF EXISTS customers",
		`CREATE TABLE customers (
			id INTEGER PRIMARY KEY,
			name TEXT,
			email TEXT,
			level TEXT,
			total_orders INTEGER,
			created_at TEXT
		)`,
		"DROP TABLE IF EXISTS orders",
		`CREATE TABLE orders (
			id INTEGER PRIMARY KEY,
			customer_id INTEGER,
			product TEXT,
			quantity INTEGER,
			amount REAL,
			status TEXT,
			created_at TEXT
		)`,
		"DROP TABLE IF EXISTS products",
		`CREATE TABLE products (
			id INTEGER PRIMARY KEY,
			name TEXT,
			category TEXT,
			price REAL,
			stock INTEGER
		)`,
	); err != nil {
		logger.L.Errorf("create warehouse tables failed: %v", err)
		return
	}

	// customers（5 行）
	customers := [][]interface{}{
		{1, "张伟", "zhangwei@corp.com", "VIP", 42, "2023-01-15 09:00:00"},
		{2, "李娜", "lina@corp.com", "Normal", 12, "2023-03-22 14:30:00"},
		{3, "王强", "wangqiang@corp.com", "VIP", 88, "2022-11-05 10:15:00"},
		{4, "刘洋", "liuyang@corp.com", "Normal", 5, "2023-07-18 16:45:00"},
		{5, "陈静", "chenjing@corp.com", "VIP", 30, "2023-09-01 11:20:00"},
	}
	insertRows(db, "INSERT INTO customers (id,name,email,level,total_orders,created_at) VALUES (?,?,?,?,?,?)", customers)

	// orders（8 行）
	orders := [][]interface{}{
		{1, 1, "服务器", 2, 25000.0, "paid", "2023-10-01 10:00:00"},
		{2, 1, "交换机", 5, 7500.0, "paid", "2023-10-05 11:00:00"},
		{3, 2, "笔记本", 3, 15000.0, "shipped", "2023-10-08 09:30:00"},
		{4, 3, "显示器", 10, 12000.0, "paid", "2023-10-10 14:00:00"},
		{5, 3, "键盘", 20, 2000.0, "pending", "2023-10-12 15:20:00"},
		{6, 4, "笔记本", 1, 5000.0, "paid", "2023-10-15 10:45:00"},
		{7, 5, "服务器", 4, 48000.0, "shipped", "2023-10-18 16:00:00"},
		{8, 5, "交换机", 2, 3000.0, "paid", "2023-10-20 09:15:00"},
	}
	insertRows(db, "INSERT INTO orders (id,customer_id,product,quantity,amount,status,created_at) VALUES (?,?,?,?,?,?,?)", orders)

	// products（5 行）
	products := [][]interface{}{
		{1, "服务器", "硬件", 12000.0, 50},
		{2, "交换机", "网络设备", 1500.0, 120},
		{3, "笔记本", "办公设备", 5000.0, 80},
		{4, "显示器", "外设", 1200.0, 200},
		{5, "键盘", "外设", 100.0, 500},
	}
	insertRows(db, "INSERT INTO products (id,name,category,price,stock) VALUES (?,?,?,?,?)", products)

	logger.L.Infof("warehouse.db seeded at %s", path)
}

// insertRows 使用统一占位符语句批量插入数据行。
func insertRows(db *sql.DB, stmt string, rows [][]interface{}) {
	tx, err := db.Begin()
	if err != nil {
		logger.L.Errorf("begin tx failed: %v", err)
		return
	}
	prep, err := tx.Prepare(stmt)
	if err != nil {
		logger.L.Errorf("prepare stmt failed: %v", err)
		_ = tx.Rollback()
		return
	}
	defer prep.Close()

	for _, r := range rows {
		if _, err := prep.Exec(r...); err != nil {
			logger.L.Errorf("insert row failed: %v", err)
			_ = tx.Rollback()
			return
		}
	}
	if err := tx.Commit(); err != nil {
		logger.L.Errorf("commit tx failed: %v", err)
	}
}
