-- ============================================================
-- ODIN 示例数据脚本（进销存本体 + 军事/海军本体）
-- 数据库：smartj（MySQL 8.4）
--
-- 内容概览：
--   1) 数据源（odin_datasource）：crm_postgres / erp_postgres / wms_postgres
--   2) 本体定义（odin_ont_definition）：inventory 进销存 / naval 海军舰艇
--   3) 本体类（odin_ont_class）+ 类属性（odin_ont_class_property）
--   4) 本体关系（odin_ont_relation）
--   5) 本体规则（odin_ont_rule）+ 规则动作（odin_ont_action）
--   6) 映射配置（odin_mapping_config）
--
-- 迁移来源：smartg-odin/internal/seed/seed_ontology.go（Go 版 SmartG-Odin 种子数据）
-- 演示业务数据已全部迁至 PostgreSQL（odin-pgsql），由 init-db/postgres/00_create_databases.sh
-- 自动建库并导入：crm（customers）/ erp（products/orders/vessels/shipments）/ wms（库存/运单/供应商）。
--
-- ID 规划（示例数据统一使用 6xxxxx 段，避免与真实业务数据冲突）：
--   数据源     6001 - 6003
--   本体定义   6100 inventory / 6200 naval
--   本体类     6101 - 6108（inventory） / 6201 - 6202（naval）
--   类属性     61001 - 61058（inventory） / 62001 - 62018（naval）
--   关系       61101 - 61107（inventory） / 62101（naval）
--   规则       61201 - 61202
--   动作       61301 - 61302
--   映射配置   61401 - 61408（inventory） / 62401 - 62402（naval）
--
-- 本脚本可重复执行：先按 ID 段清理，再幂等重建。
-- ============================================================

USE smartj;
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ------------------------------------------------------------
-- 清理旧示例数据（按 ID 段）
-- ------------------------------------------------------------
DELETE FROM odin_mapping_config     WHERE id BETWEEN 61401 AND 62499;
DELETE FROM odin_ont_action         WHERE id BETWEEN 61301 AND 61399;
DELETE FROM odin_ont_rule           WHERE id BETWEEN 61201 AND 61299;
DELETE FROM odin_ont_relation       WHERE id BETWEEN 61101 AND 62199;
DELETE FROM odin_ont_class_property WHERE id BETWEEN 61001 AND 62999;
DELETE FROM odin_ont_class          WHERE id BETWEEN 6101  AND 6299;
DELETE FROM odin_ont_definition     WHERE id IN (6100, 6200);
DELETE FROM odin_datasource         WHERE id BETWEEN 6001  AND 6099;

-- ============================================================
-- 1. 数据源（3 个演示数据源，均为 PostgreSQL 独立库）
-- ============================================================
INSERT INTO odin_datasource
(id, code, name, type, host, port, db_name, schema_name, username, password_enc, jdbc_url, options,
 pool_max_size, pool_min_idle, pool_max_life_s, conn_timeout_s, query_timeout_s, tags, status,
 create_by, create_time, update_by, update_time, remark)
VALUES
(6001, 'crm_postgres', 'CRM PostgreSQL', 'postgresql',
 'localhost', 15432, 'crm', 'public', 'odin', 'Cqmygnsdss@123', NULL, NULL,
 10, 2, 3600, 30, 60, JSON_ARRAY('demo','postgresql','crm'), 1,
 'admin', NOW(), 'admin', NOW(), '客户主数据演示库（PostgreSQL 17，odin-pgsql:15432/crm）'),

(6002, 'erp_postgres', 'ERP PostgreSQL', 'postgresql',
 'localhost', 15432, 'erp', 'public', 'odin', 'Cqmygnsdss@123', NULL, NULL,
 10, 2, 3600, 30, 60, JSON_ARRAY('demo','postgresql','erp'), 1,
 'admin', NOW(), 'admin', NOW(), '产品/订单/舰艇演示库（PostgreSQL 17，odin-pgsql:15432/erp）'),

(6003, 'wms_postgres', 'WMS PostgreSQL', 'postgresql',
 'localhost', 15432, 'wms', 'public', 'odin', 'Cqmygnsdss@123', NULL, NULL,
 10, 2, 3600, 30, 60, JSON_ARRAY('demo','postgresql','wms'), 1,
 'admin', NOW(), 'admin', NOW(), '仓储/物流/供应商演示库（PostgreSQL 17，odin-pgsql:15432/wms）');

-- ============================================================
-- 2. 本体定义
-- ============================================================
INSERT INTO odin_ont_definition
(id, code, name, description, version, preset_questions, status, create_by, create_time, update_by, update_time, remark)
VALUES
(6100, 'inventory', '库存与销售分析本体',
 '跨 PostgreSQL crm/erp/wms 三库的客户/订单/产品/库存/运单/供应商本体，含 VIP 客户与大额订单两个规则虚拟类，用于演示本体驱动的跨源虚拟集成查询与推理。',
 '2.0.0',
 JSON_ARRAY('查询华东区 VIP 客户今年的订单总额','分析近期订单延迟的主要原因','本月销售额环比增长情况如何','哪些产品库存低于安全库存需要补货'),
 1, 'admin', NOW(), 'admin', NOW(), '进销存示例本体'),

(6200, 'naval', '海军舰艇本体',
 '舰艇与货运本体（PostgreSQL erp.vessels / erp.shipments），用于演示跨源关联查询。',
 '2.0.0',
 JSON_ARRAY('列出所有舰艇及其货运任务','查询某舰艇近两个月的货运记录','各舰艇货运量排名 Top 5'),
 1, 'admin', NOW(), 'admin', NOW(), '军事/海军示例本体');

-- ============================================================
-- 3. 本体类：inventory（6 个正常类 + 2 个虚拟派生类）
-- ============================================================
INSERT INTO odin_ont_class
(id, ontology_id, code, name, description, parent_class_id, icon, sort_order, identity_keys, status,
 create_by, create_time, update_by, update_time, remark)
VALUES
(6101, 6100, 'Customer',           '客户',       '企业客户主数据（PostgreSQL crm.customers）',          NULL,  'User',        10, JSON_ARRAY('id'),   1, 'admin', NOW(), 'admin', NOW(), NULL),
(6102, 6100, 'Order',              '订单',       '销售订单（PostgreSQL erp.orders）',                    NULL,  'Tickets',     20, JSON_ARRAY('id'),   1, 'admin', NOW(), 'admin', NOW(), NULL),
(6103, 6100, 'Product',            '产品',       '产品目录（PostgreSQL erp.products）',                  NULL,  'Goods',       30, JSON_ARRAY('id'),   1, 'admin', NOW(), 'admin', NOW(), NULL),
(6104, 6100, 'WarehouseInventory', '仓库库存',   '实时仓库库存（PostgreSQL wms.warehouse_inventory）',   NULL,  'Box',         40, JSON_ARRAY('id'),   1, 'admin', NOW(), 'admin', NOW(), NULL),
(6105, 6100, 'Shipment',           '物流运单',   '物流运单（PostgreSQL wms.shipments）',                 NULL,  'Van',         50, JSON_ARRAY('id'),   1, 'admin', NOW(), 'admin', NOW(), NULL),
(6106, 6100, 'Supplier',           '供应商',     '供应商档案（PostgreSQL wms.suppliers）',               NULL,  'OfficeBuilding', 60, JSON_ARRAY('id'), 1, 'admin', NOW(), 'admin', NOW(), NULL),
(6107, 6100, 'VipCustomer',        'VIP客户',    '规则虚拟类：level=VIP 的客户',                        6101,  'Star',        70, JSON_ARRAY('id'),   1, 'admin', NOW(), 'admin', NOW(), 'virtual'),
(6108, 6100, 'BigOrder',           '大额订单',   '规则虚拟类：amount>50000 的订单',                     6102,  'Money',       80, JSON_ARRAY('id'),   1, 'admin', NOW(), 'admin', NOW(), 'virtual');

-- ============================================================
-- 4. 本体类：naval（2 个正常类）
-- ============================================================
INSERT INTO odin_ont_class
(id, ontology_id, code, name, description, parent_class_id, icon, sort_order, identity_keys, status,
 create_by, create_time, update_by, update_time, remark)
VALUES
(6201, 6200, 'Vessel',   '舰艇', '舰艇档案（PostgreSQL erp.vessels）',   NULL, 'Ship',        10, JSON_ARRAY('id'), 1, 'admin', NOW(), 'admin', NOW(), NULL),
(6202, 6200, 'Shipment', '货运', '货运记录（PostgreSQL erp.shipments）', NULL, 'TakeawayBox', 20, JSON_ARRAY('id'), 1, 'admin', NOW(), 'admin', NOW(), NULL);

-- ============================================================
-- 5. 本体类属性：inventory
--    同义词（synonyms）合并进 description，供 AI 语义对齐使用。
-- ============================================================
INSERT INTO odin_ont_class_property
(id, class_id, code, name, data_type, description, required, default_value, enum_values, unit, sort_order, status,
 create_by, create_time, update_by, update_time)
VALUES
-- Customer（6101）
(61001, 6101, 'name',        '客户名称',   'string',   NULL,                          0, NULL, NULL, NULL, 1, 1, 'admin', NOW(), 'admin', NOW()),
(61002, 6101, 'email',       '邮箱',       'string',   NULL,                          0, NULL, NULL, NULL, 2, 1, 'admin', NOW(), 'admin', NOW()),
(61003, 6101, 'phone',       '电话',       'string',   NULL,                          0, NULL, NULL, NULL, 3, 1, 'admin', NOW(), 'admin', NOW()),
(61004, 6101, 'level',       '客户等级',   'string',   NULL,                          0, NULL, JSON_ARRAY('normal','silver','gold','VIP'), NULL, 4, 1, 'admin', NOW(), 'admin', NOW()),
(61005, 6101, 'totalOrders', '累计订单数', 'int',      NULL,                          0, '0',  NULL, '单', 5, 1, 'admin', NOW(), 'admin', NOW()),
(61006, 6101, 'totalAmount', '累计金额',   'decimal',  NULL,                          0, '0',  NULL, '元', 6, 1, 'admin', NOW(), 'admin', NOW()),
(61007, 6101, 'address',     '地址',       'string',   NULL,                          0, NULL, NULL, NULL, 7, 1, 'admin', NOW(), 'admin', NOW()),
(61008, 6101, 'createdAt',   '创建时间',   'datetime', NULL,                          0, NULL, NULL, NULL, 8, 1, 'admin', NOW(), 'admin', NOW()),
-- Order（6102）
(61011, 6102, 'orderNo',   '订单号',   'string',   '同义词：订单编号,单号',                 0, NULL, NULL, NULL, 1, 1, 'admin', NOW(), 'admin', NOW()),
(61012, 6102, 'quantity',  '数量',     'int',      '同义词：销量,销售数量,销售量,购买数量',  0, '1',  NULL, '件', 2, 1, 'admin', NOW(), 'admin', NOW()),
(61013, 6102, 'unitPrice', '单价',     'decimal',  '同义词：价格',                          0, NULL, NULL, '元', 3, 1, 'admin', NOW(), 'admin', NOW()),
(61014, 6102, 'amount',    '金额',     'decimal',  '同义词：销售额,销售金额,总金额,总价,订单金额', 0, NULL, NULL, '元', 4, 1, 'admin', NOW(), 'admin', NOW()),
(61015, 6102, 'status',    '状态',     'string',   NULL,                                   0, NULL, JSON_ARRAY('pending','paid','shipped','completed','cancelled'), NULL, 5, 1, 'admin', NOW(), 'admin', NOW()),
(61016, 6102, 'orderDate', '下单时间', 'datetime', '同义词：下单日期,订单日期',              0, NULL, NULL, NULL, 6, 1, 'admin', NOW(), 'admin', NOW()),
(61017, 6102, 'shipDate',  '发货时间', 'datetime', NULL,                                   0, NULL, NULL, NULL, 7, 1, 'admin', NOW(), 'admin', NOW()),
(61018, 6102, 'remark',    '备注',     'string',   NULL,                                   0, NULL, NULL, NULL, 8, 1, 'admin', NOW(), 'admin', NOW()),
-- Product（6103）
(61021, 6103, 'name',        '产品名称', 'string',  '同义词：产品,品名',   0, NULL, NULL, NULL, 1, 1, 'admin', NOW(), 'admin', NOW()),
(61022, 6103, 'sku',         'SKU',      'string',  '同义词：产品编号',    0, NULL, NULL, NULL, 2, 1, 'admin', NOW(), 'admin', NOW()),
(61023, 6103, 'category',    '品类',     'string',  '同义词：类别,分类',   0, NULL, NULL, NULL, 3, 1, 'admin', NOW(), 'admin', NOW()),
(61024, 6103, 'price',       '单价',     'decimal', '同义词：价格',        0, NULL, NULL, '元', 4, 1, 'admin', NOW(), 'admin', NOW()),
(61025, 6103, 'stock',       '库存',     'int',     '同义词：库存量,存货', 0, '0',  NULL, '件', 5, 1, 'admin', NOW(), 'admin', NOW()),
(61026, 6103, 'supplier',    '供应商',   'string',  NULL,                 0, NULL, NULL, NULL, 6, 1, 'admin', NOW(), 'admin', NOW()),
(61027, 6103, 'description', '描述',     'string',  NULL,                 0, NULL, NULL, NULL, 7, 1, 'admin', NOW(), 'admin', NOW()),
-- WarehouseInventory（6104）
(61031, 6104, 'productSku',      '产品SKU',      'string',   NULL, 0, NULL, NULL, NULL, 1, 1, 'admin', NOW(), 'admin', NOW()),
(61032, 6104, 'productName',     '产品名称',     'string',   NULL, 0, NULL, NULL, NULL, 2, 1, 'admin', NOW(), 'admin', NOW()),
(61033, 6104, 'warehouse',       '仓库',         'string',   NULL, 0, NULL, NULL, NULL, 3, 1, 'admin', NOW(), 'admin', NOW()),
(61034, 6104, 'quantity',        '库存数量',     'int',      NULL, 0, '0',  NULL, '件', 4, 1, 'admin', NOW(), 'admin', NOW()),
(61035, 6104, 'reserved',        '预留数量',     'int',      NULL, 0, '0',  NULL, '件', 5, 1, 'admin', NOW(), 'admin', NOW()),
(61036, 6104, 'available',       '可用数量',     'int',      NULL, 0, '0',  NULL, '件', 6, 1, 'admin', NOW(), 'admin', NOW()),
(61037, 6104, 'minStock',        '最低库存',     'int',      NULL, 0, '0',  NULL, '件', 7, 1, 'admin', NOW(), 'admin', NOW()),
(61038, 6104, 'maxStock',        '最高库存',     'int',      NULL, 0, '0',  NULL, '件', 8, 1, 'admin', NOW(), 'admin', NOW()),
(61039, 6104, 'lastRestockDate', '最近补货日期', 'datetime', NULL, 0, NULL, NULL, NULL, 9, 1, 'admin', NOW(), 'admin', NOW()),
-- Shipment（6105）
(61041, 6105, 'shipmentNo',      '运单号',     'string',   NULL, 0, NULL, NULL, NULL,  1, 1, 'admin', NOW(), 'admin', NOW()),
(61042, 6105, 'orderNo',         '关联订单号', 'string',   NULL, 0, NULL, NULL, NULL,  2, 1, 'admin', NOW(), 'admin', NOW()),
(61043, 6105, 'carrier',         '承运商',     'string',   NULL, 0, NULL, NULL, NULL,  3, 1, 'admin', NOW(), 'admin', NOW()),
(61044, 6105, 'trackingNo',      '物流单号',   'string',   NULL, 0, NULL, NULL, NULL,  4, 1, 'admin', NOW(), 'admin', NOW()),
(61045, 6105, 'status',          '状态',       'string',   NULL, 0, NULL, JSON_ARRAY('pending','in_transit','delivered','cancelled'), NULL, 5, 1, 'admin', NOW(), 'admin', NOW()),
(61046, 6105, 'shipDate',        '发货日期',   'datetime', NULL, 0, NULL, NULL, NULL,  6, 1, 'admin', NOW(), 'admin', NOW()),
(61047, 6105, 'deliveryDate',    '送达日期',   'datetime', NULL, 0, NULL, NULL, NULL,  7, 1, 'admin', NOW(), 'admin', NOW()),
(61048, 6105, 'originWarehouse', '发货仓库',   'string',   NULL, 0, NULL, NULL, NULL,  8, 1, 'admin', NOW(), 'admin', NOW()),
(61049, 6105, 'destination',     '目的地',     'string',   NULL, 0, NULL, NULL, NULL,  9, 1, 'admin', NOW(), 'admin', NOW()),
(61050, 6105, 'weightKg',        '重量',       'decimal',  NULL, 0, NULL, NULL, 'kg',  10, 1, 'admin', NOW(), 'admin', NOW()),
-- Supplier（6106）
(61051, 6106, 'name',             '供应商名称', 'string',  NULL, 0, NULL, NULL, NULL, 1, 1, 'admin', NOW(), 'admin', NOW()),
(61052, 6106, 'contactPerson',    '联系人',     'string',  NULL, 0, NULL, NULL, NULL, 2, 1, 'admin', NOW(), 'admin', NOW()),
(61053, 6106, 'phone',            '电话',       'string',  NULL, 0, NULL, NULL, NULL, 3, 1, 'admin', NOW(), 'admin', NOW()),
(61054, 6106, 'email',            '邮箱',       'string',  NULL, 0, NULL, NULL, NULL, 4, 1, 'admin', NOW(), 'admin', NOW()),
(61055, 6106, 'address',          '地址',       'string',  NULL, 0, NULL, NULL, NULL, 5, 1, 'admin', NOW(), 'admin', NOW()),
(61056, 6106, 'rating',           '评分',       'decimal', NULL, 0, NULL, NULL, '分', 6, 1, 'admin', NOW(), 'admin', NOW()),
(61057, 6106, 'cooperationYears', '合作年限',   'int',     NULL, 0, '0',  NULL, '年', 7, 1, 'admin', NOW(), 'admin', NOW()),
(61058, 6106, 'status',           '状态',       'string',  NULL, 0, NULL, JSON_ARRAY('active','suspended','blacklist'), NULL, 8, 1, 'admin', NOW(), 'admin', NOW());

-- ============================================================
-- 6. 本体类属性：naval
-- ============================================================
INSERT INTO odin_ont_class_property
(id, class_id, code, name, data_type, description, required, default_value, enum_values, unit, sort_order, status,
 create_by, create_time, update_by, update_time)
VALUES
-- Vessel（6201）
(62001, 6201, 'name',     '舰名',   'string',  NULL, 0, NULL, NULL, NULL, 1, 1, 'admin', NOW(), 'admin', NOW()),
(62002, 6201, 'type',     '舰型',   'string',  NULL, 0, NULL, JSON_ARRAY('destroyer','frigate','carrier','submarine'), NULL, 2, 1, 'admin', NOW(), 'admin', NOW()),
(62003, 6201, 'capacity', '载重吨', 'decimal', NULL, 0, NULL, NULL, '吨', 3, 1, 'admin', NOW(), 'admin', NOW()),
(62004, 6201, 'flag',     '船旗国', 'string',  NULL, 0, NULL, NULL, NULL, 4, 1, 'admin', NOW(), 'admin', NOW()),
-- Shipment（6202）
(62011, 6202, 'shipmentNo',      '运单号',   'string',   NULL, 0, NULL, NULL, NULL, 1, 1, 'admin', NOW(), 'admin', NOW()),
(62012, 6202, 'carrier',         '承运方',   'string',   NULL, 0, NULL, NULL, NULL, 2, 1, 'admin', NOW(), 'admin', NOW()),
(62013, 6202, 'cargo',           '货物',     'string',   NULL, 0, NULL, NULL, NULL, 3, 1, 'admin', NOW(), 'admin', NOW()),
(62014, 6202, 'status',          '状态',     'string',   NULL, 0, NULL, JSON_ARRAY('pending','loading','at_sea','delivered'), NULL, 4, 1, 'admin', NOW(), 'admin', NOW()),
(62015, 6202, 'shipDate',        '发货日期', 'datetime', NULL, 0, NULL, NULL, NULL, 5, 1, 'admin', NOW(), 'admin', NOW()),
(62016, 6202, 'deliveryDate',    '送达日期', 'datetime', NULL, 0, NULL, NULL, NULL, 6, 1, 'admin', NOW(), 'admin', NOW()),
(62017, 6202, 'originPort',      '起运港',   'string',   NULL, 0, NULL, NULL, NULL, 7, 1, 'admin', NOW(), 'admin', NOW()),
(62018, 6202, 'destinationPort', '目的港',   'string',   NULL, 0, NULL, NULL, NULL, 8, 1, 'admin', NOW(), 'admin', NOW());

-- ============================================================
-- 7. 本体关系：inventory
--    Go 版 join_condition_json 直接映射为 join_keys。
-- ============================================================
INSERT INTO odin_ont_relation
(id, ontology_id, code, name, description, source_class_id, target_class_id, cardinality, join_keys, inverse_name, status,
 create_by, create_time, update_by, update_time, remark)
VALUES
(61101, 6100, 'places',     '下单',     '客户下订单',                    6101, 6102, '1:N',
 JSON_OBJECT('from_column','id','to_column','customer_id','join_type','LEFT'), 'ordersOf',  1, 'admin', NOW(), 'admin', NOW(), 'association'),
(61102, 6100, 'contains',   '包含产品', '订单包含的产品',                6102, 6103, 'N:1',
 JSON_OBJECT('from_column','product_id','to_column','id','join_type','LEFT'),  'orderedIn', 1, 'admin', NOW(), 'admin', NOW(), 'association'),
(61103, 6100, 'orderedIn',  '对应订单', '产品对应订单（反向关系，供产品维度聚合订单度量）', 6103, 6102, '1:N',
 JSON_OBJECT('from_column','id','to_column','product_id','join_type','LEFT'),  'contains',  1, 'admin', NOW(), 'admin', NOW(), 'association'),
(61104, 6100, 'ships',      '发运',     '订单发运为物流运单',            6102, 6105, '1:N',
 JSON_OBJECT('from_column','order_no','to_column','order_no','join_type','LEFT'), NULL,     1, 'admin', NOW(), 'admin', NOW(), 'association'),
(61105, 6100, 'stores',     '库存于',   '产品在各仓库的库存',            6103, 6104, '1:N',
 JSON_OBJECT('from_column','sku','to_column','product_sku','join_type','LEFT'),  NULL,      1, 'admin', NOW(), 'admin', NOW(), 'association'),
(61106, 6100, 'suppliedBy', '供应自',   '产品的供应商',                  6103, 6106, 'N:1',
 JSON_OBJECT('from_column','supplier','to_column','name','join_type','LEFT'),    'supplies',1, 'admin', NOW(), 'admin', NOW(), 'association'),
(61107, 6100, 'classifies', '分类为',   '客户按规则派生为 VIP 客户',     6101, 6107, '1:N',
 NULL,                                                                          NULL,      1, 'admin', NOW(), 'admin', NOW(), 'inheritance');

-- ============================================================
-- 8. 本体关系：naval
-- ============================================================
INSERT INTO odin_ont_relation
(id, ontology_id, code, name, description, source_class_id, target_class_id, cardinality, join_keys, inverse_name, status,
 create_by, create_time, update_by, update_time, remark)
VALUES
(62101, 6200, 'carries', '承运', '舰艇承运货运记录', 6201, 6202, '1:N',
 JSON_OBJECT('from_column','id','to_column','vessel_id','join_type','LEFT'), NULL, 1, 'admin', NOW(), 'admin', NOW(), 'association');

-- ============================================================
-- 9. 本体规则：VIP 客户 / 大额订单
--    Go 版 condition_json + action_json 合并写入 expression 字段。
-- ============================================================
INSERT INTO odin_ont_rule
(id, ontology_id, code, name, description, rule_type, expression, priority, enabled, status,
 create_by, create_time, update_by, update_time)
VALUES
(61201, 6100, 'vip_customer_rule', 'VIP客户派生规则',
 '当客户 level=VIP 时，派生为 VipCustomer 虚拟类实例',
 'derivation',
 JSON_OBJECT(
   'condition', JSON_OBJECT('type','value_match','property','level','operator','eq','value','VIP'),
   'action',    JSON_OBJECT('type','expand_class','target_class','VipCustomer')
 ),
 10, 1, 1, 'admin', NOW(), 'admin', NOW()),

(61202, 6100, 'big_order_rule', '大额订单派生规则',
 '当订单 amount>50000 时，派生为 BigOrder 虚拟类实例',
 'derivation',
 JSON_OBJECT(
   'condition', JSON_OBJECT('type','threshold','property','amount','operator','gt','value',50000),
   'action',    JSON_OBJECT('type','expand_class','target_class','BigOrder')
 ),
 20, 1, 1, 'admin', NOW(), 'admin', NOW());

-- ============================================================
-- 10. 规则动作
-- ============================================================
INSERT INTO odin_ont_action
(id, ontology_id, code, name, description, action_type, target_class_id, parameters, preconditions, enabled, status,
 create_by, create_time, update_by, update_time)
VALUES
(61301, 6100, 'expand_vip_customer', '展开为 VIP 客户类',
 '规则 vip_customer_rule 命中后，将 Customer 实例展开到 VipCustomer 虚拟类',
 'query', 6107,
 JSON_OBJECT('rule_code','vip_customer_rule','action','expand_class'),
 JSON_OBJECT('property','level','operator','eq','value','VIP'),
 1, 1, 'admin', NOW(), 'admin', NOW()),

(61302, 6100, 'expand_big_order', '展开为大额订单类',
 '规则 big_order_rule 命中后，将 Order 实例展开到 BigOrder 虚拟类',
 'query', 6108,
 JSON_OBJECT('rule_code','big_order_rule','action','expand_class'),
 JSON_OBJECT('property','amount','operator','gt','value',50000),
 1, 1, 'admin', NOW(), 'admin', NOW());

-- ============================================================
-- 11. 映射配置：inventory
--     虚拟类（VipCustomer / BigOrder）复用父类物理映射，过滤条件由规则注入。
-- ============================================================
INSERT INTO odin_mapping_config
(id, ontology_id, class_id, datasource_id, table_name, schema_name, column_mappings, filter_condition, priority, status,
 create_by, create_time, update_by, update_time, remark)
VALUES
-- Customer → crm_postgres.public.customers
(61401, 6100, 6101, 6001, 'customers', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','name','columnName','name','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','email','columnName','email','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','phone','columnName','phone','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','level','columnName','level','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','totalOrders','columnName','total_orders','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','totalAmount','columnName','total_amount','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','address','columnName','address','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','createdAt','columnName','created_at','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL CRM 客户主数据'),

-- Order → erp_postgres.public.orders
(61402, 6100, 6102, 6002, 'orders', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','orderNo','columnName','order_no','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','quantity','columnName','quantity','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','unitPrice','columnName','unit_price','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','amount','columnName','amount','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','status','columnName','status','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','orderDate','columnName','order_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','shipDate','columnName','ship_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','remark','columnName','remark','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL ERP 销售订单'),

-- Product → erp_postgres.public.products
(61403, 6100, 6103, 6002, 'products', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','name','columnName','name','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','sku','columnName','sku','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','category','columnName','category','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','price','columnName','price','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','stock','columnName','stock','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','supplier','columnName','supplier','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','description','columnName','description','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL ERP 产品目录'),

-- WarehouseInventory → wms_postgres.public.warehouse_inventory
(61404, 6100, 6104, 6003, 'warehouse_inventory', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','productSku','columnName','product_sku','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','productName','columnName','product_name','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','warehouse','columnName','warehouse','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','quantity','columnName','quantity','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','reserved','columnName','reserved','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','available','columnName','available','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','minStock','columnName','min_stock','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','maxStock','columnName','max_stock','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','lastRestockDate','columnName','last_restock_date','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL WMS 实时仓库库存'),

-- Shipment → wms_postgres.public.shipments（物流运单）
(61405, 6100, 6105, 6003, 'shipments', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','shipmentNo','columnName','shipment_no','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','orderNo','columnName','order_no','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','carrier','columnName','carrier','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','trackingNo','columnName','tracking_no','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','status','columnName','status','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','shipDate','columnName','ship_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','deliveryDate','columnName','delivery_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','originWarehouse','columnName','origin_warehouse','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','destination','columnName','destination','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','weightKg','columnName','weight_kg','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL WMS 物流运单'),

-- Supplier → wms_postgres.public.suppliers
(61406, 6100, 6106, 6003, 'suppliers', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','name','columnName','name','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','contactPerson','columnName','contact_person','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','phone','columnName','phone','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','email','columnName','email','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','address','columnName','address','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','rating','columnName','rating','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','cooperationYears','columnName','cooperation_years','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','status','columnName','status','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL WMS 供应商档案'),

-- VipCustomer（虚拟类）→ crm_postgres.public.customers + level=VIP 过滤
(61407, 6100, 6107, 6001, 'customers', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','name','columnName','name','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','email','columnName','email','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','phone','columnName','phone','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','level','columnName','level','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','totalOrders','columnName','total_orders','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','totalAmount','columnName','total_amount','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','address','columnName','address','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','createdAt','columnName','created_at','confidence',1.0,'validationStatus','validated')
 ),
 'level = ''VIP''', 20, 1, 'admin', NOW(), 'admin', NOW(), 'VIP客户虚拟类映射（规则派生）'),

-- BigOrder（虚拟类）→ erp_postgres.public.orders + amount>50000 过滤
(61408, 6100, 6108, 6002, 'orders', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','orderNo','columnName','order_no','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','quantity','columnName','quantity','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','unitPrice','columnName','unit_price','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','amount','columnName','amount','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','status','columnName','status','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','orderDate','columnName','order_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','shipDate','columnName','ship_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','remark','columnName','remark','confidence',1.0,'validationStatus','validated')
 ),
 'amount > 50000', 20, 1, 'admin', NOW(), 'admin', NOW(), '大额订单虚拟类映射（规则派生）');

-- ============================================================
-- 12. 映射配置：naval
-- ============================================================
INSERT INTO odin_mapping_config
(id, ontology_id, class_id, datasource_id, table_name, schema_name, column_mappings, filter_condition, priority, status,
 create_by, create_time, update_by, update_time, remark)
VALUES
-- Vessel → erp_postgres.public.vessels
(62401, 6200, 6201, 6002, 'vessels', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','name','columnName','name','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','type','columnName','type','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','capacity','columnName','capacity','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','flag','columnName','flag','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL ERP 舰艇档案'),

-- Shipment → erp_postgres.public.shipments
(62402, 6200, 6202, 6002, 'shipments', 'public',
 JSON_ARRAY(
   JSON_OBJECT('propertyCode','shipmentNo','columnName','shipment_no','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','carrier','columnName','carrier','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','cargo','columnName','cargo','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','status','columnName','status','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','shipDate','columnName','ship_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','deliveryDate','columnName','delivery_date','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','originPort','columnName','origin_port','confidence',1.0,'validationStatus','validated'),
   JSON_OBJECT('propertyCode','destinationPort','columnName','destination_port','confidence',1.0,'validationStatus','validated')
 ),
 NULL, 10, 1, 'admin', NOW(), 'admin', NOW(), 'PostgreSQL ERP 舰艇货运记录');

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 示例数据装载完成
--   * 3 个数据源
--   * 2 个本体（inventory 进销存 / naval 海军舰艇）
--   * 10 个本体类（含 2 个规则派生虚拟类）
--   * 62 个类属性
--   * 8 条本体关系
--   * 2 条派生规则 + 2 个规则动作
--   * 10 条映射配置
-- ============================================================
