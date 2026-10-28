-- ============================================================
-- ODIN CRM 演示库初始化脚本（PostgreSQL 17）
-- 数据库：crm  用户：odin  容器：odin-pgsql:15432
-- 存放客户主数据，供「库存与销售分析本体」的 Customer 类映射。
-- 由 00_create_databases.sh 建库后经 psql 执行，无需 USE 语句。
-- 演示数据时间：2026年9月—10月（近两个月）
-- ============================================================

-- ------------------------------------------------------------
-- 客户表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(200),
    phone VARCHAR(20),
    level VARCHAR(20) DEFAULT 'normal',        -- normal, silver, gold, VIP
    total_orders INT DEFAULT 0,
    total_amount NUMERIC(12,2) DEFAULT 0,
    address VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE customers IS '客户主数据';

-- ------------------------------------------------------------
-- 示例数据：客户（10 条，覆盖不同 level）
-- ------------------------------------------------------------
INSERT INTO customers (name, email, phone, level, total_orders, total_amount, address, created_at) VALUES
('张伟', 'zhangwei@corp.com', '13800138001', 'VIP', 56, 234500.00, '北京市朝阳区建国路88号', '2026-09-01 09:00:00'),
('李娜', 'lina@corp.com', '13800138002', 'gold', 32, 128000.00, '上海市浦东新区陆家嘴环路1000号', '2026-09-01 09:30:00'),
('王强', 'wangqiang@corp.com', '13800138003', 'VIP', 48, 196000.00, '广州市天河区天河路385号', '2026-09-02 10:00:00'),
('刘洋', 'liuyang@corp.com', '13800138004', 'silver', 15, 45000.00, '深圳市南山区科技园南路18号', '2026-09-03 11:15:00'),
('陈静', 'chenjing@corp.com', '13800138005', 'normal', 5, 12000.00, '杭州市西湖区文三路90号', '2026-09-05 14:20:00'),
('赵鹏', 'zhaopeng@corp.com', '13800138006', 'gold', 28, 112000.00, '成都市高新区天府大道北段1700号', '2026-09-08 09:45:00'),
('孙丽', 'sunli@corp.com', '13800138007', 'VIP', 62, 310000.00, '南京市建邺区江东中路289号', '2026-09-10 10:30:00'),
('周杰', 'zhoujie@corp.com', '13800138008', 'normal', 3, 8500.00, '武汉市洪山区珞喻路1037号', '2026-09-12 15:00:00'),
('吴芳', 'wufang@corp.com', '13800138009', 'silver', 12, 36000.00, '西安市雁塔区高新路25号', '2026-09-15 11:20:00'),
('郑浩', 'zhenghao@corp.com', '13800138010', 'gold', 35, 156000.00, '重庆市渝北区龙溪街道红锦大道1号', '2026-09-18 16:40:00');
