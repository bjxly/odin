-- ============================================================
-- ODIN 仓储/物流演示库初始化脚本（PostgreSQL 17）
-- 容器：odin-pgsql  数据库：demo  用户：odin
-- 存放仓库库存 / 物流运单 / 供应商数据，
-- 通过 product_sku 与 MySQL 的产品表关联，验证跨源虚拟集成。
-- 说明：docker-entrypoint-initdb.d 下的 .sql 会自动在 $POSTGRES_DB(demo) 中执行。
-- ============================================================

-- ------------------------------------------------------------
-- 仓库库存表（实时库存，与 MySQL 产品表通过 SKU 关联）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS warehouse_inventory (
    id SERIAL PRIMARY KEY,
    product_sku VARCHAR(50) NOT NULL,
    product_name VARCHAR(200),
    warehouse VARCHAR(100) NOT NULL,
    quantity INT DEFAULT 0,
    reserved INT DEFAULT 0,
    available INT GENERATED ALWAYS AS (quantity - reserved) STORED,
    min_stock INT DEFAULT 10,
    max_stock INT DEFAULT 1000,
    last_restock_date TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE warehouse_inventory IS '仓库实时库存';

-- ------------------------------------------------------------
-- 物流运单表（通过 order_no 与 MySQL 订单表关联）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shipments (
    id SERIAL PRIMARY KEY,
    shipment_no VARCHAR(50) NOT NULL UNIQUE,
    order_no VARCHAR(50),
    carrier VARCHAR(100),
    tracking_no VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',      -- pending, picked, packed, shipped, delivered
    ship_date TIMESTAMP,
    delivery_date TIMESTAMP,
    origin_warehouse VARCHAR(100),
    destination VARCHAR(500),
    weight_kg DECIMAL(8,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE shipments IS '物流运单';

-- ------------------------------------------------------------
-- 供应商表（通过 name 与 MySQL 产品表 supplier 字段关联）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    contact_person VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(200),
    address VARCHAR(500),
    rating DECIMAL(3,2) DEFAULT 0,             -- 0-5 分
    cooperation_years INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE suppliers IS '供应商档案';

-- ------------------------------------------------------------
-- 示例数据：仓库库存（8 条，对应 MySQL 8 个产品 SKU）
-- ------------------------------------------------------------
INSERT INTO warehouse_inventory (product_sku, product_name, warehouse, quantity, reserved, min_stock, max_stock, last_restock_date) VALUES
('SENSOR-A100', '工业传感器A100', '北京主仓', 150, 20, 30, 500, '2025-01-20'),
('GATEWAY-GW200', '智能网关GW-200', '北京主仓', 45, 5, 10, 100, '2025-01-15'),
('DAQ-16CH', '数据采集卡DAQ-16', '上海分仓', 80, 12, 20, 200, '2025-01-25'),
('PLC-S7200', 'PLC控制器S7-200', '上海分仓', 30, 8, 5, 50, '2025-01-10'),
('VFD-75KW', '变频驱动器VFD-75', '广州分仓', 25, 3, 5, 60, '2025-02-01'),
('SWITCH-IS24', '工业交换机IS-24', '北京主仓', 60, 10, 15, 150, '2025-01-28'),
('PRESS-PT100', '压力变送器PT-100', '广州分仓', 200, 30, 50, 500, '2025-02-05'),
('HMI-10INCH', '人机界面HMI-10', '上海分仓', 35, 4, 10, 80, '2025-01-18');

-- ------------------------------------------------------------
-- 示例数据：物流运单（6 条，order_no 关联 MySQL 订单）
-- ------------------------------------------------------------
INSERT INTO shipments (shipment_no, order_no, carrier, tracking_no, status, ship_date, delivery_date, origin_warehouse, destination, weight_kg) VALUES
('SHP-20250103-001', 'ORD-20250101-001', '顺丰速运', 'SF1234567890', 'delivered', '2025-01-02', '2025-01-03', '北京主仓', '北京市朝阳区建国路88号', 5.2),
('SHP-20250106-002', 'ORD-20250105-003', '德邦物流', 'DB9876543210', 'delivered', '2025-01-06', '2025-01-08', '北京主仓', '上海市浦东新区陆家嘴环路1000号', 25.0),
('SHP-20250111-003', 'ORD-20250110-005', '京东物流', 'JD2025011101', 'delivered', '2025-01-11', '2025-01-12', '北京主仓', '南京市建邺区江东中路289号', 48.5),
('SHP-20250116-004', 'ORD-20250115-007', '顺丰速运', 'SF2345678901', 'shipped', '2025-01-16', NULL, '北京主仓', '成都市高新区天府大道北段1700号', 18.0),
('SHP-20250121-005', 'ORD-20250120-009', '德邦物流', 'DB1234567891', 'delivered', '2025-01-21', '2025-01-24', '广州分仓', '重庆市渝北区龙溪街道红锦大道1号', 35.0),
('SHP-20250211-006', 'ORD-20250210-015', '京东物流', 'JD2025021101', 'shipped', '2025-02-11', NULL, '广州分仓', '广州市天河区天河路385号', 42.0);

-- ------------------------------------------------------------
-- 示例数据：供应商（5 条）
-- ------------------------------------------------------------
INSERT INTO suppliers (name, contact_person, phone, email, address, rating, cooperation_years, status) VALUES
('深圳华芯电子有限公司', '陈明', '0755-86001234', 'sales@huaxin.com', '深圳市南山区科技园', 4.50, 5, 'active'),
('北京智联科技股份有限公司', '王磊', '010-82001234', 'wang@zhilian-tech.com', '北京市海淀区中关村', 4.20, 3, 'active'),
('上海测控仪器有限公司', '李红', '021-64001234', 'lihong@shck.com', '上海市徐汇区漕河泾', 4.80, 8, 'active'),
('西门子(中国)代理商-京信达', '赵强', '010-65001234', 'zhao@jxd-automation.com', '北京市朝阳区望京', 4.00, 6, 'active'),
('ABB授权经销商-华控电气', '孙伟', '0512-67001234', 'sun@huakong.com', '苏州市工业园区', 4.30, 4, 'active');
