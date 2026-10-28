-- ============================================================
-- ODIN WMS 演示库初始化脚本（PostgreSQL 17）
-- 数据库：wms  用户：odin  容器：odin-pgsql:15432
-- 存放仓库库存 / 物流运单 / 供应商数据。
-- 由 00_create_databases.sh 建库后经 psql 执行，无需 USE 语句。
-- 演示数据时间：2026年9月—10月（近两个月）
-- 跨库关联：product_sku → erp.products.sku，
--           order_no    → erp.orders.order_no，
--           name        → erp.products.supplier（供应商简称对应）
-- ============================================================

-- ------------------------------------------------------------
-- 仓库库存表（实时库存，与 erp 产品表通过 SKU 关联）
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
-- 物流运单表（通过 order_no 与 erp 订单表关联）
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
    weight_kg NUMERIC(8,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE shipments IS '物流运单';

-- ------------------------------------------------------------
-- 供应商表（通过 name 与 erp 产品表 supplier 字段关联）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    contact_person VARCHAR(100),
    phone VARCHAR(20),
    email VARCHAR(200),
    address VARCHAR(500),
    rating NUMERIC(3,2) DEFAULT 0,             -- 0-5 分
    cooperation_years INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE suppliers IS '供应商档案';

-- ------------------------------------------------------------
-- 示例数据：仓库库存（8 条，对应 erp 8 个产品 SKU，最近补货在9-10月）
-- ------------------------------------------------------------
INSERT INTO warehouse_inventory (product_sku, product_name, warehouse, quantity, reserved, min_stock, max_stock, last_restock_date) VALUES
('SENSOR-A100', '工业传感器A100', '北京主仓', 150, 20, 30, 500, '2026-10-20 08:00:00'),
('GATEWAY-GW200', '智能网关GW-200', '北京主仓', 45, 5, 10, 100, '2026-10-12 08:00:00'),
('DAQ-16CH', '数据采集卡DAQ-16', '上海分仓', 80, 12, 20, 200, '2026-10-25 08:00:00'),
('PLC-S7200', 'PLC控制器S7-200', '上海分仓', 30, 8, 5, 50, '2026-09-28 08:00:00'),
('VFD-75KW', '变频驱动器VFD-75', '广州分仓', 25, 3, 5, 60, '2026-10-08 08:00:00'),
('SWITCH-IS24', '工业交换机IS-24', '北京主仓', 60, 10, 15, 150, '2026-10-18 08:00:00'),
('PRESS-PT100', '压力变送器PT-100', '广州分仓', 200, 30, 50, 500, '2026-10-22 08:00:00'),
('HMI-10INCH', '人机界面HMI-10', '上海分仓', 35, 4, 10, 80, '2026-09-30 08:00:00');

-- ------------------------------------------------------------
-- 示例数据：物流运单（8 条，order_no 关联 erp 订单，覆盖 9-10 月）
-- ------------------------------------------------------------
INSERT INTO shipments (shipment_no, order_no, carrier, tracking_no, status, ship_date, delivery_date, origin_warehouse, destination, weight_kg) VALUES
('SHP-20260916-001', 'ORD-20260915-001', '顺丰速运', 'SF1234567890', 'delivered', '2026-09-16', '2026-09-17', '北京主仓', '北京市朝阳区建国路88号', 5.2),
('SHP-20260919-002', 'ORD-20260917-003', '德邦物流', 'DB9876543210', 'delivered', '2026-09-19', '2026-09-21', '北京主仓', '上海市浦东新区陆家嘴环路1000号', 25.0),
('SHP-20260923-003', 'ORD-20260922-005', '京东物流', 'JD2026092301', 'delivered', '2026-09-23', '2026-09-24', '北京主仓', '南京市建邺区江东中路289号', 48.5),
('SHP-20260927-004', 'ORD-20260926-007', '顺丰速运', 'SF2345678901', 'shipped', '2026-09-27', NULL, '北京主仓', '成都市高新区天府大道北段1700号', 18.0),
('SHP-20261001-005', 'ORD-20260930-009', '德邦物流', 'DB1234567891', 'delivered', '2026-10-01', '2026-10-03', '广州分仓', '重庆市渝北区龙溪街道红锦大道1号', 35.0),
('SHP-20261017-006', 'ORD-20261015-015', '京东物流', 'JD2026101701', 'shipped', '2026-10-17', NULL, '广州分仓', '广州市天河区天河路385号', 42.0),
('SHP-20261019-007', 'ORD-20261017-016', '顺丰速运', 'SF3456789012', 'delivered', '2026-10-19', '2026-10-21', '上海分仓', '上海市浦东新区陆家嘴环路1000号', 66.0),
('SHP-20261024-008', 'ORD-20261022-018', '德邦物流', 'DB2345678901', 'shipped', '2026-10-24', NULL, '北京主仓', '重庆市渝北区龙溪街道红锦大道1号', 40.0);

-- ------------------------------------------------------------
-- 示例数据：供应商（5 条）
-- ------------------------------------------------------------
INSERT INTO suppliers (name, contact_person, phone, email, address, rating, cooperation_years, status) VALUES
('深圳华芯电子有限公司', '陈明', '0755-86001234', 'sales@huaxin.com', '深圳市南山区科技园', 4.50, 5, 'active'),
('北京智联科技股份有限公司', '王磊', '010-82001234', 'wang@zhilian-tech.com', '北京市海淀区中关村', 4.20, 3, 'active'),
('上海测控仪器有限公司', '李红', '021-64001234', 'lihong@shck.com', '上海市徐汇区漕河泾', 4.80, 8, 'active'),
('西门子(中国)代理商-京信达', '赵强', '010-65001234', 'zhao@jxd-automation.com', '北京市朝阳区望京', 4.00, 6, 'active'),
('ABB授权经销商-华控电气', '孙伟', '0512-67001234', 'sun@huakong.com', '苏州市工业园区', 4.30, 4, 'active');
