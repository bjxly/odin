-- ============================================================
-- ODIN ERP 演示库初始化脚本（PostgreSQL 17）
-- 数据库：erp  用户：odin  容器：odin-pgsql:15432
-- 存放产品目录 / 销售订单 / 舰艇档案 / 舰艇货运数据。
-- 由 00_create_databases.sh 建库后经 psql 执行，无需 USE 语句。
-- 演示数据时间：2026年9月—10月（近两个月）
-- 说明：orders.customer_id 逻辑关联 crm.customers.id，
--       orders.product_id 逻辑关联本库 products.id（跨库不设物理外键）。
-- ============================================================

-- ------------------------------------------------------------
-- 产品表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    sku VARCHAR(50) UNIQUE,
    category VARCHAR(100),
    price NUMERIC(10,2),
    stock INT DEFAULT 0,
    supplier VARCHAR(200),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE products IS '产品目录';

-- ------------------------------------------------------------
-- 销售订单表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    order_no VARCHAR(50) NOT NULL UNIQUE,
    customer_id INT NOT NULL,                   -- 逻辑关联 crm.customers.id
    product_id INT,                             -- 逻辑关联 erp.products.id
    quantity INT DEFAULT 1,
    unit_price NUMERIC(10,2),
    amount NUMERIC(12,2),
    status VARCHAR(20) DEFAULT 'pending',      -- pending, paid, shipped, completed, cancelled
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ship_date TIMESTAMP,
    remark VARCHAR(500)
);
COMMENT ON TABLE orders IS '销售订单';

-- ------------------------------------------------------------
-- 示例数据：产品（8 条，SKU 与 wms 库存表关联）
-- ------------------------------------------------------------
INSERT INTO products (name, sku, category, price, stock, supplier, description, created_at) VALUES
('工业传感器A100', 'SENSOR-A100', '传感器', 2580.00, 150, '深圳华芯电子', '高精度温度传感器，测量范围-40~125℃', '2026-09-01 08:00:00'),
('智能网关GW-200', 'GATEWAY-GW200', '网关设备', 8900.00, 45, '北京智联科技', '支持Modbus/MQTT/OPC-UA多协议转换', '2026-09-01 08:00:00'),
('数据采集卡DAQ-16', 'DAQ-16CH', '采集设备', 4500.00, 80, '上海测控仪器', '16通道模拟量输入，采样率100kS/s', '2026-09-01 08:00:00'),
('PLC控制器S7-200', 'PLC-S7200', '控制器', 12800.00, 30, '西门子代理商', '中小型自动化控制，支持以太网通信', '2026-09-01 08:00:00'),
('变频驱动器VFD-75', 'VFD-75KW', '驱动设备', 6700.00, 25, 'ABB授权经销', '75kW三相变频调速驱动器', '2026-09-01 08:00:00'),
('工业交换机IS-24', 'SWITCH-IS24', '网络设备', 3200.00, 60, '杭州华三通信', '24口工业级管理型以太网交换机', '2026-09-01 08:00:00'),
('压力变送器PT-100', 'PRESS-PT100', '传感器', 1850.00, 200, '西安仪表厂', '0-100MPa压力测量，4-20mA输出', '2026-09-01 08:00:00'),
('人机界面HMI-10', 'HMI-10INCH', '显示设备', 5600.00, 35, '深圳步科电气', '10寸触控人机界面，支持远程访问', '2026-09-01 08:00:00');

-- ------------------------------------------------------------
-- 示例数据：订单（18 条，2026年9月—10月，覆盖不同 status，
-- 含 amount>50000 的大额订单用于 BigOrder 规则演示）
-- ------------------------------------------------------------
INSERT INTO orders (order_no, customer_id, product_id, quantity, unit_price, amount, status, order_date, ship_date, remark) VALUES
('ORD-20260915-001', 1, 1, 10, 2580.00, 25800.00, 'completed', '2026-09-15 09:30:00', '2026-09-16 10:00:00', '加急订单'),
('ORD-20260916-002', 1, 4, 2, 12800.00, 25600.00, 'completed', '2026-09-16 14:00:00', '2026-09-18 09:00:00', NULL),
('ORD-20260917-003', 2, 2, 5, 8900.00, 44500.00, 'shipped', '2026-09-17 10:15:00', '2026-09-19 15:30:00', '物流单号SF1234567'),
('ORD-20260919-004', 3, 5, 3, 6700.00, 20100.00, 'completed', '2026-09-19 16:45:00', '2026-09-21 09:20:00', NULL),
('ORD-20260922-005', 7, 1, 50, 2580.00, 129000.00, 'completed', '2026-09-22 08:00:00', '2026-09-23 10:00:00', '大客户批量采购'),
('ORD-20260924-006', 4, 3, 4, 4500.00, 18000.00, 'paid', '2026-09-24 11:30:00', NULL, NULL),
('ORD-20260926-007', 6, 6, 8, 3200.00, 25600.00, 'shipped', '2026-09-26 09:00:00', '2026-09-27 14:00:00', NULL),
('ORD-20260928-008', 7, 4, 5, 12800.00, 64000.00, 'paid', '2026-09-28 15:20:00', NULL, '含安装调试服务'),
('ORD-20260930-009', 10, 7, 20, 1850.00, 37000.00, 'completed', '2026-09-30 10:00:00', '2026-10-01 09:30:00', NULL),
('ORD-20261003-010', 3, 8, 2, 5600.00, 11200.00, 'pending', '2026-10-03 13:45:00', NULL, NULL),
('ORD-20261006-011', 1, 2, 10, 8900.00, 89000.00, 'paid', '2026-10-06 09:15:00', NULL, '二期项目采购'),
('ORD-20261008-012', 5, 1, 3, 2580.00, 7740.00, 'cancelled', '2026-10-08 16:00:00', NULL, '客户取消'),
('ORD-20261010-013', 8, 3, 1, 4500.00, 4500.00, 'pending', '2026-10-10 08:30:00', NULL, NULL),
('ORD-20261013-014', 9, 7, 10, 1850.00, 18500.00, 'paid', '2026-10-13 11:00:00', NULL, NULL),
('ORD-20261015-015', 7, 5, 4, 6700.00, 26800.00, 'shipped', '2026-10-15 14:30:00', '2026-10-17 09:00:00', NULL),
('ORD-20261017-016', 2, 4, 6, 12800.00, 76800.00, 'completed', '2026-10-17 10:40:00', '2026-10-19 13:10:00', '产线扩容采购'),
('ORD-20261020-017', 6, 8, 4, 5600.00, 22400.00, 'paid', '2026-10-20 15:50:00', NULL, NULL),
('ORD-20261022-018', 10, 2, 8, 8900.00, 71200.00, 'shipped', '2026-10-22 09:25:00', '2026-10-24 10:00:00', '分两批发货');

-- ------------------------------------------------------------
-- 舰艇本体数据（naval 本体：Vessel / Shipment）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS vessels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50),          -- destroyer, frigate, carrier, submarine
    capacity NUMERIC(10,2),    -- 吨位
    flag VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    commissioned_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE vessels IS '舰艇档案';

CREATE TABLE IF NOT EXISTS shipments (
    id SERIAL PRIMARY KEY,
    shipment_no VARCHAR(50) NOT NULL UNIQUE,
    vessel_id INT REFERENCES vessels(id),
    carrier VARCHAR(100),
    cargo VARCHAR(200),
    status VARCHAR(20) DEFAULT 'pending',  -- pending, loading, at_sea, delivered
    ship_date TIMESTAMP,
    delivery_date TIMESTAMP,
    origin_port VARCHAR(100),
    destination_port VARCHAR(100)
);
COMMENT ON TABLE shipments IS '舰艇货运记录';

INSERT INTO vessels (name, type, capacity, flag, status, commissioned_date) VALUES
('青岛舰', 'destroyer', 7500.00, 'CN', 'active', '2015-03-15'),
('合肥舰', 'destroyer', 7500.00, 'CN', 'active', '2015-12-01'),
('黄山舰', 'frigate', 4000.00, 'CN', 'active', '2008-05-20'),
('辽宁舰', 'carrier', 67000.00, 'CN', 'active', '2012-09-25'),
('长城艇', 'submarine', 3000.00, 'CN', 'maintenance', '2010-08-10');

INSERT INTO shipments (shipment_no, vessel_id, carrier, cargo, status, ship_date, delivery_date, origin_port, destination_port) VALUES
('NAV-2026-001', 1, '海军物流', '补给物资', 'delivered', '2026-09-05', '2026-09-08', '青岛港', '上海港'),
('NAV-2026-002', 2, '海军物流', '装备配件', 'at_sea', '2026-09-20', NULL, '上海港', '广州港'),
('NAV-2026-003', 3, '民用航运', '食品补给', 'delivered', '2026-10-02', '2026-10-05', '大连港', '青岛港'),
('NAV-2026-004', 4, '海军物流', '航空燃料', 'loading', '2026-10-18', NULL, '宁波港', '三亚港'),
('NAV-2026-005', 1, '民用航运', '医疗物资', 'pending', NULL, NULL, '天津港', '青岛港');
