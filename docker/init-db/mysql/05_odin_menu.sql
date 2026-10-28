USE smartj;
SET NAMES utf8mb4;

-- ============================================================
-- ODIN 菜单（2026-10 信息架构重构版）
--
-- 结构：
--   2000 ODIN（一级目录）
--   ├── 2010 助手        /odin/assistant
--   ├── 2020 本体治理    /odin/ontology
--   ├── 2030 数据接入    （目录）
--   │   ├── 2031 连接器管理  /odin/data-access/connectors
--   │   ├── 2032 数据源      /odin/data-access/sources
--   │   ├── 2033 映射工作台  /odin/data-access/mapping
--   │   └── 2034 推理        /odin/data-access/reason
--   └── 2040 数据洞察    /odin/insight
--
-- 首页复用现有 /dashboard 静态路由，不新建菜单。
-- 按钮权限 ID 规则：父菜单ID * 100 + 序号（如 2032 → 203201~）。
--
-- 注意：sys_menu 实际列为 menu_name/menu_type(1=目录 2=菜单 3=按钮)，
-- 本脚本可重复执行：先清空 ODIN 旧菜单段再幂等重建。
-- ============================================================

-- 清理旧 ODIN 菜单（含 2026-10 之前的 数据源管理/映射管理/智能查询/AI助手/业务洞察/变更历史 结构）
-- 注意：按钮权限 ID 为 6 位（父ID*100+序号），故清理范围需覆盖 id >= 2000 全段（系统菜单均 < 2000）
DELETE FROM sys_role_menu WHERE menu_id >= 2000;
DELETE FROM sys_menu WHERE id >= 2000;

-- ----------------------------
-- 一级目录：ODIN
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(2000, 0, 'ODIN', 1, '/odin', NULL, 'Compass', NULL, 50, 1, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, 'ODIN 本体数据洞察');

-- ----------------------------
-- 二级菜单
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(2010, 2000, '助手',     2, '/odin/assistant',  'odin/assistant/index', 'ChatDotRound', 'odin:assistant:chat', 1, 1, 1, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(2020, 2000, '本体治理', 2, '/odin/ontology',   'odin/ontology/index',  'Connection',   'odin:ontology:list',  2, 1, 1, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(2030, 2000, '数据接入', 1, '/odin/data-access', NULL,                  'Coin',         NULL,                  3, 1, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(2040, 2000, '数据洞察', 2, '/odin/insight',    'odin/insight/index',   'DataAnalysis', 'odin:insight:list',   4, 1, 1, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 数据接入子菜单
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(2031, 2030, '连接器管理', 2, '/odin/data-access/connectors', 'odin/data-access/connectors/index', 'SetUp', 'odin:connector:list',  1, 1, 1, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(2032, 2030, '数据源',     2, '/odin/data-access/sources',    'odin/data-access/sources/index',    'Files', 'odin:datasource:list', 2, 1, 1, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(2033, 2030, '映射工作台', 2, '/odin/data-access/mapping',    'odin/data-access/mapping/index',    'Link',  'odin:mapping:list',    3, 1, 1, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(2034, 2030, '推理',       2, '/odin/data-access/reason',     'odin/data-access/reason/index',     'Cpu',   'odin:reason:list',     4, 1, 1, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：助手（2010）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(201001, 2010, '对话',     3, NULL, NULL, NULL, 'odin:assistant:chat',    1, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(201002, 2010, '会话管理', 3, NULL, NULL, NULL, 'odin:assistant:session', 2, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(201003, 2010, '技能管理', 3, NULL, NULL, NULL, 'odin:assistant:skills',  3, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(201004, 2010, '草稿审阅', 3, NULL, NULL, NULL, 'odin:assistant:draft',   4, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：本体治理（2020）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(202001, 2020, '本体查询', 3, NULL, NULL, NULL, 'odin:ontology:query',    1, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(202002, 2020, '本体新增', 3, NULL, NULL, NULL, 'odin:ontology:create',   2, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(202003, 2020, '本体修改', 3, NULL, NULL, NULL, 'odin:ontology:update',   3, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(202004, 2020, '本体删除', 3, NULL, NULL, NULL, 'odin:ontology:delete',   4, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(202005, 2020, '类管理',   3, NULL, NULL, NULL, 'odin:ontology:class',    5, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(202006, 2020, '关系管理', 3, NULL, NULL, NULL, 'odin:ontology:relation', 6, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：连接器管理（2031）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(203101, 2031, '连接器查询', 3, NULL, NULL, NULL, 'odin:connector:query',  1, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203102, 2031, '连接器新增', 3, NULL, NULL, NULL, 'odin:connector:create', 2, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203103, 2031, '连接器修改', 3, NULL, NULL, NULL, 'odin:connector:update', 3, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203104, 2031, '连接器删除', 3, NULL, NULL, NULL, 'odin:connector:delete', 4, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：数据源（2032）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(203201, 2032, '数据源查询', 3, NULL, NULL, NULL, 'odin:datasource:query',  1, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203202, 2032, '数据源新增', 3, NULL, NULL, NULL, 'odin:datasource:create', 2, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203203, 2032, '数据源修改', 3, NULL, NULL, NULL, 'odin:datasource:update', 3, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203204, 2032, '数据源删除', 3, NULL, NULL, NULL, 'odin:datasource:delete', 4, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203205, 2032, '测试连接',   3, NULL, NULL, NULL, 'odin:datasource:test',   5, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203206, 2032, '同步元数据', 3, NULL, NULL, NULL, 'odin:datasource:sync',   6, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：映射工作台（2033）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(203301, 2033, '映射查询', 3, NULL, NULL, NULL, 'odin:mapping:query',    1, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203302, 2033, '映射新增', 3, NULL, NULL, NULL, 'odin:mapping:create',   2, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203303, 2033, '映射修改', 3, NULL, NULL, NULL, 'odin:mapping:update',   3, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203304, 2033, '映射删除', 3, NULL, NULL, NULL, 'odin:mapping:delete',   4, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203305, 2033, '映射验证', 3, NULL, NULL, NULL, 'odin:mapping:validate', 5, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：推理（2034）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(203401, 2034, '推理执行', 3, NULL, NULL, NULL, 'odin:reason:execute', 1, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203402, 2034, '推理查询', 3, NULL, NULL, NULL, 'odin:reason:query',   2, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(203403, 2034, '规则保存', 3, NULL, NULL, NULL, 'odin:reason:save',    3, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：数据洞察（2040）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(204001, 2040, '洞察查询', 3, NULL, NULL, NULL, 'odin:insight:query',  1, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204002, 2040, '洞察新增', 3, NULL, NULL, NULL, 'odin:insight:create', 2, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204003, 2040, '洞察修改', 3, NULL, NULL, NULL, 'odin:insight:update', 3, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204004, 2040, '洞察删除', 3, NULL, NULL, NULL, 'odin:insight:delete', 4, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204005, 2040, '执行监控', 3, NULL, NULL, NULL, 'odin:insight:run',    5, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204006, 2040, '处理告警', 3, NULL, NULL, NULL, 'odin:insight:handle', 6, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：智能查询（挂载在 数据洞察 2040 下，/odin/query 菜单已并入助手）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(204011, 2040, '执行查询', 3, NULL, NULL, NULL, 'odin:query:execute',  11, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204012, 2040, '查询历史', 3, NULL, NULL, NULL, 'odin:query:history',  12, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204013, 2040, '模板管理', 3, NULL, NULL, NULL, 'odin:query:template', 13, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204014, 2040, '模板新增', 3, NULL, NULL, NULL, 'odin:query:create',   14, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(204015, 2040, '模板删除', 3, NULL, NULL, NULL, 'odin:query:delete',   15, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);

-- ----------------------------
-- 按钮权限：变更历史（挂载在 本体治理 2020 下，/odin/changes 独立菜单已移除）
-- ----------------------------
INSERT IGNORE INTO sys_menu (id, parent_id, menu_name, menu_type, path, component, icon, permission, sort_order, visible, keep_alive, is_external, status, del_flg, create_by, create_time, update_by, update_time, version, remark) VALUES
(202011, 2020, '变更查询', 3, NULL, NULL, NULL, 'odin:changes:list',   11, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL),
(202012, 2020, '变更回滚', 3, NULL, NULL, NULL, 'odin:changes:revert', 12, 0, 0, 0, 1, 0, 'admin', NOW(), 'admin', NOW(), 0, NULL);
