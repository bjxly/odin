-- ============================================================
-- ODIN ERP 演示库初始化脚本（MySQL 8.4）
-- 容器：odin-mysql  数据库：smartj
-- 存放客户 / 订单 / 产品等 ERP 核心业务数据，
-- 供「库存与销售分析本体」映射验证完整链路。
-- ============================================================

USE smartj;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ------------------------------------------------------------
-- 客户表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(200),
    phone VARCHAR(20),
    level VARCHAR(20) DEFAULT 'normal',        -- normal, silver, gold, VIP
    total_orders INT DEFAULT 0,
    total_amount DECIMAL(12,2) DEFAULT 0,
    address VARCHAR(500),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户主数据';

-- ------------------------------------------------------------
-- 产品表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    sku VARCHAR(50) UNIQUE,
    category VARCHAR(100),
    price DECIMAL(10,2),
    stock INT DEFAULT 0,
    supplier VARCHAR(200),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品目录';

-- ------------------------------------------------------------
-- 订单表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(50) NOT NULL UNIQUE,
    customer_id INT NOT NULL,
    product_id INT,
    quantity INT DEFAULT 1,
    unit_price DECIMAL(10,2),
    amount DECIMAL(12,2),
    status VARCHAR(20) DEFAULT 'pending',      -- pending, paid, shipped, completed, cancelled
    order_date DATETIME DEFAULT CURRENT_TIMESTAMP,
    ship_date DATETIME,
    remark VARCHAR(500),
    FOREIGN KEY (customer_id) REFERENCES customers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='销售订单';

SET FOREIGN_KEY_CHECKS = 1;

-- ------------------------------------------------------------
-- 示例数据：客户（10 条，覆盖不同 level）
-- ------------------------------------------------------------
INSERT INTO customers (name, email, phone, level, total_orders, total_amount, address) VALUES
('张伟', 'zhangwei@corp.com', '13800138001', 'VIP', 56, 234500.00, '北京市朝阳区建国路88号'),
('李娜', 'lina@corp.com', '13800138002', 'gold', 32, 128000.00, '上海市浦东新区陆家嘴环路1000号'),
('王强', 'wangqiang@corp.com', '13800138003', 'VIP', 48, 196000.00, '广州市天河区天河路385号'),
('刘洋', 'liuyang@corp.com', '13800138004', 'silver', 15, 45000.00, '深圳市南山区科技园南路18号'),
('陈静', 'chenjing@corp.com', '13800138005', 'normal', 5, 12000.00, '杭州市西湖区文三路90号'),
('赵鹏', 'zhaopeng@corp.com', '13800138006', 'gold', 28, 112000.00, '成都市高新区天府大道北段1700号'),
('孙丽', 'sunli@corp.com', '13800138007', 'VIP', 62, 310000.00, '南京市建邺区江东中路289号'),
('周杰', 'zhoujie@corp.com', '13800138008', 'normal', 3, 8500.00, '武汉市洪山区珞喻路1037号'),
('吴芳', 'wufang@corp.com', '13800138009', 'silver', 12, 36000.00, '西安市雁塔区高新路25号'),
('郑浩', 'zhenghao@corp.com', '13800138010', 'gold', 35, 156000.00, '重庆市渝北区龙溪街道红锦大道1号');

-- ------------------------------------------------------------
-- 示例数据：产品（8 条）
-- ------------------------------------------------------------
INSERT INTO products (name, sku, category, price, stock, supplier, description) VALUES
('工业传感器A100', 'SENSOR-A100', '传感器', 2580.00, 150, '深圳华芯电子', '高精度温度传感器，测量范围-40~125℃'),
('智能网关GW-200', 'GATEWAY-GW200', '网关设备', 8900.00, 45, '北京智联科技', '支持Modbus/MQTT/OPC-UA多协议转换'),
('数据采集卡DAQ-16', 'DAQ-16CH', '采集设备', 4500.00, 80, '上海测控仪器', '16通道模拟量输入，采样率100kS/s'),
('PLC控制器S7-200', 'PLC-S7200', '控制器', 12800.00, 30, '西门子代理商', '中小型自动化控制，支持以太网通信'),
('变频驱动器VFD-75', 'VFD-75KW', '驱动设备', 6700.00, 25, 'ABB授权经销', '75kW三相变频调速驱动器'),
('工业交换机IS-24', 'SWITCH-IS24', '网络设备', 3200.00, 60, '杭州华三通信', '24口工业级管理型以太网交换机'),
('压力变送器PT-100', 'PRESS-PT100', '传感器', 1850.00, 200, '西安仪表厂', '0-100MPa压力测量，4-20mA输出'),
('人机界面HMI-10', 'HMI-10INCH', '显示设备', 5600.00, 35, '深圳步科电气', '10寸触控人机界面，支持远程访问');

-- ------------------------------------------------------------
-- 示例数据：订单（15 条，关联客户与产品，覆盖不同 status）
-- ------------------------------------------------------------
INSERT INTO orders (order_no, customer_id, product_id, quantity, unit_price, amount, status, order_date, remark) VALUES
('ORD-20250101-001', 1, 1, 10, 2580.00, 25800.00, 'completed', '2025-01-01 09:30:00', '加急订单'),
('ORD-20250102-002', 1, 4, 2, 12800.00, 25600.00, 'completed', '2025-01-02 14:00:00', NULL),
('ORD-20250105-003', 2, 2, 5, 8900.00, 44500.00, 'shipped', '2025-01-05 10:15:00', '物流单号SF1234567'),
('ORD-20250108-004', 3, 5, 3, 6700.00, 20100.00, 'completed', '2025-01-08 16:45:00', NULL),
('ORD-20250110-005', 7, 1, 50, 2580.00, 129000.00, 'completed', '2025-01-10 08:00:00', '大客户批量采购'),
('ORD-20250112-006', 4, 3, 4, 4500.00, 18000.00, 'paid', '2025-01-12 11:30:00', NULL),
('ORD-20250115-007', 6, 6, 8, 3200.00, 25600.00, 'shipped', '2025-01-15 09:00:00', NULL),
('ORD-20250118-008', 7, 4, 5, 12800.00, 64000.00, 'paid', '2025-01-18 15:20:00', '含安装调试服务'),
('ORD-20250120-009', 10, 7, 20, 1850.00, 37000.00, 'completed', '2025-01-20 10:00:00', NULL),
('ORD-20250122-010', 3, 8, 2, 5600.00, 11200.00, 'pending', '2025-01-22 13:45:00', NULL),
('ORD-20250125-011', 1, 2, 10, 8900.00, 89000.00, 'paid', '2025-01-25 09:15:00', '二期项目采购'),
('ORD-20250128-012', 5, 1, 3, 2580.00, 7740.00, 'cancelled', '2025-01-28 16:00:00', '客户取消'),
('ORD-20250201-013', 8, 3, 1, 4500.00, 4500.00, 'pending', '2025-02-01 08:30:00', NULL),
('ORD-20250205-014', 9, 7, 10, 1850.00, 18500.00, 'paid', '2025-02-05 11:00:00', NULL),
('ORD-20250210-015', 7, 5, 4, 6700.00, 26800.00, 'shipped', '2025-02-10 14:30:00', NULL);

-- ------------------------------------------------------------
-- 海军舰艇本体数据（naval 本体：Vessel / Shipment）
-- 原位于 analytics_sqlite（SQLite 文件），已迁移至 MySQL。
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vessels (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50),          -- destroyer, frigate, carrier, submarine
    capacity DECIMAL(10,2),    -- 吨位
    flag VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    commissioned_date DATE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='舰艇档案';

CREATE TABLE IF NOT EXISTS shipments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    shipment_no VARCHAR(50) NOT NULL UNIQUE,
    vessel_id INT,
    carrier VARCHAR(100),
    cargo VARCHAR(200),
    status VARCHAR(20) DEFAULT 'pending',  -- pending, loading, at_sea, delivered
    ship_date DATETIME,
    delivery_date DATETIME,
    origin_port VARCHAR(100),
    destination_port VARCHAR(100),
    FOREIGN KEY (vessel_id) REFERENCES vessels(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='舰艇货运记录';

INSERT INTO vessels (name, type, capacity, flag, status, commissioned_date) VALUES
('青岛舰', 'destroyer', 7500.00, 'CN', 'active', '2015-03-15'),
('合肥舰', 'destroyer', 7500.00, 'CN', 'active', '2015-12-01'),
('黄山舰', 'frigate', 4000.00, 'CN', 'active', '2008-05-20'),
('辽宁舰', 'carrier', 67000.00, 'CN', 'active', '2012-09-25'),
('长城艇', 'submarine', 3000.00, 'CN', 'maintenance', '2010-08-10');

INSERT INTO shipments (shipment_no, vessel_id, carrier, cargo, status, ship_date, delivery_date, origin_port, destination_port) VALUES
('NAV-2025-001', 1, '海军物流', '补给物资', 'delivered', '2025-01-05', '2025-01-08', '青岛港', '上海港'),
('NAV-2025-002', 2, '海军物流', '装备配件', 'at_sea', '2025-01-10', NULL, '上海港', '广州港'),
('NAV-2025-003', 3, '民用航运', '食品补给', 'delivered', '2025-01-12', '2025-01-15', '大连港', '青岛港'),
('NAV-2025-004', 4, '海军物流', '航空燃料', 'loading', '2025-01-20', NULL, '宁波港', '三亚港'),
('NAV-2025-005', 1, '民用航运', '医疗物资', 'pending', NULL, NULL, '天津港', '青岛港');
