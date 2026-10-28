-- ============================================================
-- SmartJ-Odin 应用系统表（MySQL 8.4，库：smartj）
-- 实体继承 BaseSimpleEntity：id/create_by/create_time/update_by/update_time
-- ============================================================
USE smartj;
SET NAMES utf8mb4;

-- 数据源配置
CREATE TABLE IF NOT EXISTS odin_datasource (
    id              BIGINT       NOT NULL COMMENT '主键',
    code            VARCHAR(64)  NOT NULL COMMENT '数据源编码(唯一)',
    name            VARCHAR(128) NOT NULL COMMENT '名称',
    type            VARCHAR(32)  NOT NULL COMMENT '类型: mysql/postgresql/dameng/kingbase/gaussdb/oracle/sqlserver/duckdb_file/rest',
    host            VARCHAR(255) NULL COMMENT '主机',
    port            INT          NULL COMMENT '端口',
    db_name         VARCHAR(128) NULL COMMENT '数据库名',
    schema_name     VARCHAR(128) NULL COMMENT '模式名',
    username        VARCHAR(128) NULL COMMENT '用户名',
    password_enc    VARCHAR(512) NULL COMMENT 'Jasypt加密后的密码',
    jdbc_url        VARCHAR(1024)NULL COMMENT '完整JDBC URL(优先)',
    options         JSON         NULL COMMENT '扩展参数',
    pool_max_size   INT          NOT NULL DEFAULT 10 COMMENT '连接池最大连接数',
    pool_min_idle   INT          NOT NULL DEFAULT 2  COMMENT '连接池最小空闲',
    pool_max_life_s INT          NOT NULL DEFAULT 3600 COMMENT '连接最大存活秒',
    conn_timeout_s  INT          NOT NULL DEFAULT 30 COMMENT '连接超时秒',
    query_timeout_s INT          NOT NULL DEFAULT 60 COMMENT '查询超时秒',
    tags            JSON         NULL COMMENT '标签数组',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    remark          VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ds_code (code),
    KEY idx_ds_type (type),
    KEY idx_ds_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据源配置';

-- 数据源测试记录
CREATE TABLE IF NOT EXISTS odin_datasource_test (
    id             BIGINT      NOT NULL COMMENT '主键',
    datasource_id  BIGINT      NOT NULL COMMENT '数据源ID',
    test_time      DATETIME    NOT NULL COMMENT '测试时间',
    success        TINYINT     NOT NULL DEFAULT 0 COMMENT '是否成功 0=否 1=是',
    latency_ms     INT         NULL COMMENT '耗时毫秒',
    db_version     VARCHAR(128)NULL COMMENT '数据库版本',
    error_message  VARCHAR(1000)NULL COMMENT '错误信息',
    create_by      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
    create_time    DATETIME    NOT NULL COMMENT '创建时间',
    update_by      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
    update_time    DATETIME    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_dst_ds (datasource_id),
    KEY idx_dst_time (test_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据源测试记录';

-- 表元数据
CREATE TABLE IF NOT EXISTS odin_schema_table (
    id             BIGINT      NOT NULL COMMENT '主键',
    datasource_id  BIGINT      NOT NULL COMMENT '数据源ID',
    schema_name    VARCHAR(128)NULL COMMENT '模式名',
    table_name     VARCHAR(128)NOT NULL COMMENT '表名',
    table_type     VARCHAR(32) NOT NULL DEFAULT 'TABLE' COMMENT '表类型',
    comment        VARCHAR(500)NULL COMMENT '表注释',
    row_estimate   BIGINT      NULL COMMENT '行数估计',
    fingerprint    VARCHAR(64) NULL COMMENT '结构指纹(变更检测)',
    last_sync_time DATETIME    NULL COMMENT '最后同步时间',
    status         TINYINT     NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
    create_time    DATETIME    NOT NULL COMMENT '创建时间',
    update_by      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
    update_time    DATETIME    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_st (datasource_id, schema_name, table_name),
    KEY idx_st_ds (datasource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='表元数据';

-- 列元数据
CREATE TABLE IF NOT EXISTS odin_schema_column (
    id               BIGINT      NOT NULL COMMENT '主键',
    table_id         BIGINT      NOT NULL COMMENT '表元数据ID',
    column_name      VARCHAR(128)NOT NULL COMMENT '列名',
    data_type        VARCHAR(64) NOT NULL COMMENT '源库原始类型',
    canonical_type   VARCHAR(32) NULL COMMENT '归一类型: string/int/decimal/bool/date/datetime/json',
    nullable         TINYINT     NOT NULL DEFAULT 1 COMMENT '是否可空 0=否 1=是',
    default_value    VARCHAR(255)NULL COMMENT '默认值',
    comment          VARCHAR(500)NULL COMMENT '列注释',
    max_length       INT         NULL COMMENT '最大长度',
    numeric_precision INT        NULL COMMENT '数值精度',
    numeric_scale    INT         NULL COMMENT '数值小数位',
    ordinal_position INT         NOT NULL COMMENT '列序号',
    is_primary_key   TINYINT     NOT NULL DEFAULT 0 COMMENT '是否主键 0=否 1=是',
    is_unique        TINYINT     NOT NULL DEFAULT 0 COMMENT '是否唯一 0=否 1=是',
    synonyms         JSON        NULL COMMENT '列同义词(AI对齐用)',
    status           TINYINT     NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by        VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
    create_time      DATETIME    NOT NULL COMMENT '创建时间',
    update_by        VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
    update_time      DATETIME    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sc (table_id, column_name),
    KEY idx_sc_table (table_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='列元数据';

-- 外键元数据
CREATE TABLE IF NOT EXISTS odin_schema_foreign_key (
    id              BIGINT      NOT NULL COMMENT '主键',
    table_id        BIGINT      NOT NULL COMMENT '表元数据ID',
    constraint_name VARCHAR(128)NOT NULL COMMENT '约束名',
    column_name     VARCHAR(128)NOT NULL COMMENT '列名',
    ref_schema      VARCHAR(128)NULL COMMENT '引用模式',
    ref_table       VARCHAR(128)NOT NULL COMMENT '引用表',
    ref_column      VARCHAR(128)NOT NULL COMMENT '引用列',
    on_delete       VARCHAR(32) NULL COMMENT '删除级联策略',
    on_update       VARCHAR(32) NULL COMMENT '更新级联策略',
    create_by       VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME    NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_fk_table (table_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外键元数据';

-- 索引元数据
CREATE TABLE IF NOT EXISTS odin_schema_index (
    id           BIGINT      NOT NULL COMMENT '主键',
    table_id     BIGINT      NOT NULL COMMENT '表元数据ID',
    index_name   VARCHAR(128)NOT NULL COMMENT '索引名',
    column_names JSON        NOT NULL COMMENT '列名数组',
    is_unique    TINYINT     NOT NULL DEFAULT 0 COMMENT '是否唯一 0=否 1=是',
    index_type   VARCHAR(32) NOT NULL DEFAULT 'BTREE' COMMENT '索引类型',
    create_by    VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人',
    create_time  DATETIME    NOT NULL COMMENT '创建时间',
    update_by    VARCHAR(64) NOT NULL DEFAULT '' COMMENT '更新人',
    update_time  DATETIME    NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_si (table_id, index_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='索引元数据';

-- ============================================================
-- B组：本体管理
-- ============================================================

-- 本体定义
CREATE TABLE IF NOT EXISTS odin_ont_definition (
    id              BIGINT       NOT NULL COMMENT '主键',
    code            VARCHAR(64)  NOT NULL COMMENT '本体编码(唯一)',
    name            VARCHAR(128) NOT NULL COMMENT '本体名称',
    description     TEXT         NULL COMMENT '描述',
    version         VARCHAR(32)  NOT NULL DEFAULT '1.0' COMMENT '版本号',
    preset_questions JSON        NULL COMMENT '助手预置问题列表(JSON字符串数组)',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    remark          VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ont_def_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='本体定义';

-- 本体类
CREATE TABLE IF NOT EXISTS odin_ont_class (
    id              BIGINT       NOT NULL COMMENT '主键',
    ontology_id     BIGINT       NOT NULL COMMENT '本体定义ID',
    code            VARCHAR(64)  NOT NULL COMMENT '类编码',
    name            VARCHAR(128) NOT NULL COMMENT '类名称',
    description     TEXT         NULL COMMENT '描述',
    parent_class_id BIGINT       NULL COMMENT '父类ID',
    icon            VARCHAR(64)  NULL COMMENT '图标',
    sort_order      INT          NOT NULL DEFAULT 0 COMMENT '排序',
    identity_keys   JSON         NULL COMMENT '唯一标识键数组',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    remark          VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ont_class (ontology_id, code),
    KEY idx_oc_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='本体类';

-- 本体类属性
CREATE TABLE IF NOT EXISTS odin_ont_class_property (
    id              BIGINT       NOT NULL COMMENT '主键',
    class_id        BIGINT       NOT NULL COMMENT '本体类ID',
    code            VARCHAR(64)  NOT NULL COMMENT '属性编码',
    name            VARCHAR(128) NOT NULL COMMENT '属性名称',
    data_type       VARCHAR(32)  NOT NULL COMMENT '数据类型: string/int/decimal/bool/date/datetime/json/enum',
    description     TEXT         NULL COMMENT '描述',
    required        TINYINT      NOT NULL DEFAULT 0 COMMENT '是否必填 0=否 1=是',
    default_value   VARCHAR(256) NULL COMMENT '默认值',
    enum_values     JSON         NULL COMMENT '枚举值列表',
    unit            VARCHAR(32)  NULL COMMENT '单位',
    sort_order      INT          NOT NULL DEFAULT 0 COMMENT '排序',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ocp (class_id, code),
    KEY idx_ocp_class (class_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='本体类属性';

-- 本体关系
CREATE TABLE IF NOT EXISTS odin_ont_relation (
    id              BIGINT       NOT NULL COMMENT '主键',
    ontology_id     BIGINT       NOT NULL COMMENT '本体定义ID',
    code            VARCHAR(64)  NOT NULL COMMENT '关系编码',
    name            VARCHAR(128) NOT NULL COMMENT '关系名称',
    description     TEXT         NULL COMMENT '描述',
    source_class_id BIGINT       NOT NULL COMMENT '源类ID',
    target_class_id BIGINT       NOT NULL COMMENT '目标类ID',
    cardinality     VARCHAR(8)   NOT NULL COMMENT '基数: 1:1/1:N/N:M',
    join_keys       JSON         NULL COMMENT '关联键映射',
    inverse_name    VARCHAR(128) NULL COMMENT '反向关系名',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    remark          VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ont_rel (ontology_id, code),
    KEY idx_or_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='本体关系';

-- 本体规则
CREATE TABLE IF NOT EXISTS odin_ont_rule (
    id              BIGINT       NOT NULL COMMENT '主键',
    ontology_id     BIGINT       NOT NULL COMMENT '本体定义ID',
    code            VARCHAR(64)  NOT NULL COMMENT '规则编码',
    name            VARCHAR(128) NOT NULL COMMENT '规则名称',
    description     TEXT         NULL COMMENT '描述',
    rule_type       VARCHAR(32)  NOT NULL COMMENT '规则类型: derivation/constraint/validation',
    expression      TEXT         NULL COMMENT '规则表达式',
    priority        INT          NOT NULL DEFAULT 0 COMMENT '优先级',
    enabled         TINYINT      NOT NULL DEFAULT 1 COMMENT '是否启用 0=否 1=是',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ont_rule (ontology_id, code),
    KEY idx_orule_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='本体规则';

-- 本体动作
CREATE TABLE IF NOT EXISTS odin_ont_action (
    id              BIGINT       NOT NULL COMMENT '主键',
    ontology_id     BIGINT       NOT NULL COMMENT '本体定义ID',
    code            VARCHAR(64)  NOT NULL COMMENT '动作编码',
    name            VARCHAR(128) NOT NULL COMMENT '动作名称',
    description     TEXT         NULL COMMENT '描述',
    action_type     VARCHAR(32)  NOT NULL COMMENT '动作类型: query/update/notification',
    target_class_id BIGINT       NULL COMMENT '目标类ID',
    parameters      JSON         NULL COMMENT '参数定义',
    preconditions   JSON         NULL COMMENT '前置条件',
    enabled         TINYINT      NOT NULL DEFAULT 1 COMMENT '是否启用 0=否 1=是',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ont_action (ontology_id, code),
    KEY idx_oa_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='本体动作';

-- ============================================================
-- C组：映射配置
-- ============================================================

-- 映射配置
CREATE TABLE IF NOT EXISTS odin_mapping_config (
    id              BIGINT       NOT NULL COMMENT '主键',
    ontology_id     BIGINT       NOT NULL COMMENT '本体定义ID',
    class_id        BIGINT       NOT NULL COMMENT '本体类ID',
    datasource_id   BIGINT       NOT NULL COMMENT '数据源ID',
    table_name      VARCHAR(128) NOT NULL COMMENT '物理表名',
    schema_name     VARCHAR(128) NULL COMMENT '模式名',
    column_mappings JSON         NULL COMMENT '列映射关系',
    filter_condition TEXT        NULL COMMENT '过滤条件',
    priority        INT          NOT NULL DEFAULT 0 COMMENT '优先级',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    remark          VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    UNIQUE KEY uk_mc (ontology_id, class_id, datasource_id, table_name),
    KEY idx_mc_ontology (ontology_id),
    KEY idx_mc_class (class_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='映射配置';

-- ============================================================
-- D组：查询溯源
-- ============================================================

-- 查询模板
CREATE TABLE IF NOT EXISTS odin_query_template (
    id              BIGINT       NOT NULL COMMENT '主键',
    name            VARCHAR(128) NOT NULL COMMENT '模板名称',
    description     TEXT         NULL COMMENT '描述',
    ontology_id     BIGINT       NULL COMMENT '本体定义ID',
    logical_sql     TEXT         NULL COMMENT '本体逻辑SQL',
    parameters      JSON         NULL COMMENT '参数定义',
    tags            JSON         NULL COMMENT '标签数组',
    usage_count     INT          NOT NULL DEFAULT 0 COMMENT '使用次数',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    remark          VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    KEY idx_qt_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='查询模板';

-- 查询历史
CREATE TABLE IF NOT EXISTS odin_query_history (
    id              BIGINT       NOT NULL COMMENT '主键',
    session_id      BIGINT       NULL COMMENT '会话ID',
    user_id         BIGINT       NULL COMMENT '用户ID',
    natural_language TEXT        NULL COMMENT '自然语言输入',
    logical_sql     TEXT         NULL COMMENT '本体逻辑SQL',
    physical_sql    TEXT         NULL COMMENT '物理执行SQL',
    result_summary  JSON         NULL COMMENT '结果摘要',
    latency_ms      INT          NULL COMMENT '耗时毫秒',
    status          VARCHAR(16)  NOT NULL DEFAULT 'success' COMMENT '状态: success/failed/timeout',
    error_message   TEXT         NULL COMMENT '错误信息',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_qh_session (session_id),
    KEY idx_qh_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='查询历史';

-- 查询追踪
CREATE TABLE IF NOT EXISTS odin_query_trace (
    id              BIGINT       NOT NULL COMMENT '主键',
    history_id      BIGINT       NOT NULL COMMENT '查询历史ID',
    step_index      INT          NOT NULL COMMENT '步骤序号',
    step_type       VARCHAR(32)  NOT NULL COMMENT '步骤类型: source_query/merge/transform',
    datasource_id   BIGINT       NULL COMMENT '数据源ID',
    sql_text        TEXT         NULL COMMENT '执行SQL',
    row_count       INT          NULL COMMENT '返回行数',
    latency_ms      INT          NULL COMMENT '耗时毫秒',
    details         JSON         NULL COMMENT '详细信息',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_qt_history (history_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='查询追踪';

-- 意图缓存
CREATE TABLE IF NOT EXISTS odin_intent_cache (
    id                  BIGINT       NOT NULL COMMENT '主键',
    nl_hash             VARCHAR(64)  NOT NULL COMMENT '自然语言哈希',
    ontology_version    VARCHAR(32)  NULL COMMENT '本体版本',
    logical_sql         TEXT         NULL COMMENT '缓存的本体逻辑SQL',
    physical_plan_hash  VARCHAR(64)  NULL COMMENT '物理执行计划哈希',
    hit_count           INT          NOT NULL DEFAULT 0 COMMENT '命中次数',
    last_hit_time       DATETIME     NULL COMMENT '最后命中时间',
    status              TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by           VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time         DATETIME     NOT NULL COMMENT '创建时间',
    update_by           VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time         DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ic (nl_hash, ontology_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='意图缓存';

-- ============================================================
-- F组：助手会话
-- ============================================================

-- 助手会话
CREATE TABLE IF NOT EXISTS odin_assistant_session (
    id              BIGINT       NOT NULL COMMENT '主键',
    user_id         BIGINT       NOT NULL COMMENT '用户ID',
    title           VARCHAR(256) NULL COMMENT '会话标题',
    ontology_id     BIGINT       NULL COMMENT '本体定义ID',
    status          VARCHAR(16)  NOT NULL DEFAULT 'active' COMMENT '状态: active/archived',
    message_count   INT          NOT NULL DEFAULT 0 COMMENT '消息数',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_as_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='助手会话';

-- 助手消息
CREATE TABLE IF NOT EXISTS odin_assistant_message (
    id              BIGINT       NOT NULL COMMENT '主键',
    session_id      BIGINT       NOT NULL COMMENT '会话ID',
    role            VARCHAR(16)  NOT NULL COMMENT '角色: user/assistant/system',
    content         TEXT         NULL COMMENT '消息内容',
    content_type    VARCHAR(32)  NOT NULL DEFAULT 'text' COMMENT '内容类型: text/sql/table/chart/conclusion',
    metadata        JSON         NULL COMMENT '元数据',
    tokens_used     INT          NULL COMMENT '消耗token数',
    latency_ms      INT          NULL COMMENT '耗时毫秒',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_am_session (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='助手消息';

-- ============================================================
-- E组：AI草稿审阅
-- ============================================================

-- AI草稿
CREATE TABLE IF NOT EXISTS odin_ai_draft (
    id              BIGINT       NOT NULL COMMENT '主键',
    draft_type      VARCHAR(32)  NOT NULL COMMENT '草稿类型: datasource/ontology/mapping',
    ontology_id     BIGINT       NULL COMMENT '本体定义ID',
    status          VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT '状态: pending/approved/rejected/published',
    created_by_ai   TINYINT      NOT NULL DEFAULT 1 COMMENT '是否AI创建 0=否 1=是',
    summary         TEXT         NULL COMMENT '摘要',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_ad_type (draft_type),
    KEY idx_ad_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI草稿';

-- AI草稿明细
CREATE TABLE IF NOT EXISTS odin_ai_draft_item (
    id              BIGINT       NOT NULL COMMENT '主键',
    draft_id        BIGINT       NOT NULL COMMENT '草稿ID',
    item_type       VARCHAR(32)  NULL COMMENT '明细类型',
    action          VARCHAR(16)  NULL COMMENT '操作: create/update/delete',
    target_id       BIGINT       NULL COMMENT '目标实体ID',
    before_snapshot JSON         NULL COMMENT '变更前快照',
    after_snapshot  JSON         NULL COMMENT '变更后快照',
    review_status   VARCHAR(16)  NOT NULL DEFAULT 'pending' COMMENT '审阅状态: pending/approved/rejected',
    reviewer_comment TEXT        NULL COMMENT '审阅意见',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_adi_draft (draft_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI草稿明细';

-- 概念知识库
CREATE TABLE IF NOT EXISTS odin_concept_kb (
    id              BIGINT       NOT NULL COMMENT '主键',
    ontology_id     BIGINT       NOT NULL COMMENT '本体定义ID',
    concept_code    VARCHAR(64)  NOT NULL COMMENT '概念编码',
    concept_name    VARCHAR(128) NOT NULL COMMENT '概念名称',
    summary         TEXT         NULL COMMENT '概念摘要',
    embedding_status VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '向量化状态: pending/done/failed',
    source_type     VARCHAR(16)  NOT NULL DEFAULT 'manual' COMMENT '来源类型: manual/ai',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_ckb (ontology_id, concept_code),
    KEY idx_ckb_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='概念知识库';

-- ============================================================
-- H组：业务洞察
-- ============================================================

-- 洞察监控规则
CREATE TABLE IF NOT EXISTS odin_insight_monitor (
    id                    BIGINT       NOT NULL COMMENT '主键',
    name                  VARCHAR(128) NOT NULL COMMENT '监控名称',
    description           TEXT         NULL COMMENT '描述',
    ontology_id           BIGINT       NULL COMMENT '本体定义ID',
    rule_expression       TEXT         NULL COMMENT '规则表达式',
    schedule_cron         VARCHAR(64)  NULL COMMENT '调度Cron表达式',
    thresholds            JSON         NULL COMMENT '阈值配置',
    notification_channels JSON         NULL COMMENT '通知渠道',
    enabled               TINYINT      NOT NULL DEFAULT 1 COMMENT '是否启用 0=否 1=是',
    status                TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by             VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time           DATETIME     NOT NULL COMMENT '创建时间',
    update_by             VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time           DATETIME     NOT NULL COMMENT '更新时间',
    remark                VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    KEY idx_im_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='洞察监控规则';

-- 洞察告警
CREATE TABLE IF NOT EXISTS odin_insight_alert (
    id              BIGINT       NOT NULL COMMENT '主键',
    monitor_id      BIGINT       NOT NULL COMMENT '监控规则ID',
    trigger_time    DATETIME     NOT NULL COMMENT '触发时间',
    severity        VARCHAR(16)  NULL COMMENT '严重程度: info/warning/critical',
    title           VARCHAR(256) NULL COMMENT '告警标题',
    detail          TEXT         NULL COMMENT '告警详情',
    metric_values   JSON         NULL COMMENT '指标值',
    handled         TINYINT      NOT NULL DEFAULT 0 COMMENT '是否已处理 0=否 1=是',
    handled_by      VARCHAR(64)  NULL COMMENT '处理人',
    handled_time    DATETIME     NULL COMMENT '处理时间',
    handle_note     TEXT         NULL COMMENT '处理备注',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_ia_monitor (monitor_id),
    KEY idx_ia_trigger (trigger_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='洞察告警';

-- 洞察看板（AI 助手图表另存）
CREATE TABLE IF NOT EXISTS odin_insight_dashboard (
    id            BIGINT       NOT NULL COMMENT '主键',
    name          VARCHAR(128) NOT NULL COMMENT '看板名称',
    ontology_id   BIGINT       NULL COMMENT '本体定义ID',
    title         VARCHAR(256) NULL COMMENT '图表标题',
    chart_type    VARCHAR(16)  NULL COMMENT '图型: bar/line/pie',
    columns_json  JSON         NULL COMMENT '数据集列头',
    rows_json     JSON         NULL COMMENT '数据集行(字符串二维数组)',
    chart_option  JSON         NULL COMMENT '图表配置快照(选择/聚合等)',
    source        VARCHAR(32)  NOT NULL DEFAULT 'assistant_chart' COMMENT '来源',
    status        TINYINT      NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
    create_by     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time   DATETIME     NOT NULL COMMENT '创建时间',
    update_by     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time   DATETIME     NOT NULL COMMENT '更新时间',
    remark        VARCHAR(500) NULL COMMENT '备注',
    PRIMARY KEY (id),
    KEY idx_idb_ontology (ontology_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='洞察看板';

-- ============================================================
-- I组：变更历史
-- ============================================================

-- 变更历史
CREATE TABLE IF NOT EXISTS odin_change_history (
    id              BIGINT       NOT NULL COMMENT '主键',
    entity_type     VARCHAR(64)  NOT NULL COMMENT '实体类型',
    entity_id       BIGINT       NOT NULL COMMENT '实体ID',
    action          VARCHAR(16)  NOT NULL COMMENT '操作: create/update/delete/publish/revert',
    before_snapshot JSON         NULL COMMENT '变更前快照',
    after_snapshot  JSON         NULL COMMENT '变更后快照',
    operator        VARCHAR(64)  NULL COMMENT '操作人',
    reason          TEXT         NULL COMMENT '变更原因',
    create_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建人',
    create_time     DATETIME     NOT NULL COMMENT '创建时间',
    update_by       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '更新人',
    update_time     DATETIME     NOT NULL COMMENT '更新时间',
    PRIMARY KEY (id),
    KEY idx_ch_entity (entity_type, entity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='变更历史';
