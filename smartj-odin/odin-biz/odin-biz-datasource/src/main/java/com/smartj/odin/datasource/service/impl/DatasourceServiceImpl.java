/*
 * SmartJ 企业级多租户基础框架
 * Copyright (C) 2024-2026 SmartJ Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package com.smartj.odin.datasource.service.impl;

import com.mybatisflex.core.paginate.Page;
import com.mybatisflex.core.query.QueryWrapper;
import com.smartj.common.core.domain.PageResult;
import com.smartj.common.core.exception.BusinessException;
import com.smartj.common.core.util.BeanCopyUtil;
import com.smartj.odin.common.udap.Connector;
import com.smartj.odin.common.udap.model.ColumnMeta;
import com.smartj.odin.common.udap.model.ConnectorConfig;
import com.smartj.odin.common.udap.model.ForeignKeyMeta;
import com.smartj.odin.common.udap.model.IndexMeta;
import com.smartj.odin.common.udap.model.TableMeta;
import com.smartj.odin.common.udap.model.TestResult;
import com.smartj.odin.datasource.converter.DatasourceConverter;
import com.smartj.odin.datasource.dto.DatasourceDTO;
import com.smartj.odin.datasource.dto.DatasourceQuery;
import com.smartj.odin.datasource.entity.Datasource;
import com.smartj.odin.datasource.entity.DatasourceTest;
import com.smartj.odin.datasource.entity.SchemaColumn;
import com.smartj.odin.datasource.entity.SchemaForeignKey;
import com.smartj.odin.datasource.entity.SchemaIndex;
import com.smartj.odin.datasource.entity.SchemaTable;
import com.smartj.odin.datasource.error.DatasourceErrorCode;
import com.smartj.odin.datasource.mapper.DatasourceMapper;
import com.smartj.odin.datasource.mapper.DatasourceTestMapper;
import com.smartj.odin.datasource.mapper.SchemaColumnMapper;
import com.smartj.odin.datasource.mapper.SchemaForeignKeyMapper;
import com.smartj.odin.datasource.mapper.SchemaIndexMapper;
import com.smartj.odin.datasource.mapper.SchemaTableMapper;
import com.smartj.odin.datasource.service.DatasourceService;
import com.smartj.odin.datasource.vo.DatasourceTestResultVO;
import com.smartj.odin.datasource.vo.DatasourceVO;
import com.smartj.odin.datasource.vo.SchemaColumnVO;
import com.smartj.odin.datasource.vo.SchemaTableVO;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.jasypt.encryption.StringEncryptor;
import org.jasypt.exceptions.EncryptionOperationNotPossibleException;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.DigestUtils;
import org.springframework.util.StringUtils;

import java.nio.charset.StandardCharsets;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.stream.Collectors;

import static com.smartj.odin.datasource.entity.table.DatasourceTableDef.DATASOURCE;
import static com.smartj.odin.datasource.entity.table.DatasourceTestTableDef.DATASOURCE_TEST;
import static com.smartj.odin.datasource.entity.table.SchemaColumnTableDef.SCHEMA_COLUMN;
import static com.smartj.odin.datasource.entity.table.SchemaForeignKeyTableDef.SCHEMA_FOREIGN_KEY;
import static com.smartj.odin.datasource.entity.table.SchemaIndexTableDef.SCHEMA_INDEX;
import static com.smartj.odin.datasource.entity.table.SchemaTableTableDef.SCHEMA_TABLE;

/**
 * 数据源服务实现。
 *
 * <p>提供数据源 CRUD、连接测试、元数据同步与浏览能力。密码使用 Jasypt 加密存储，
 * 连接与内省委托 {@link Connector}。写操作使用事务保证一致性。
 *
 * @author SmartG-Odin
 * @since 1.0.0
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class DatasourceServiceImpl implements DatasourceService {

    private final DatasourceMapper datasourceMapper;
    private final DatasourceTestMapper datasourceTestMapper;
    private final SchemaTableMapper schemaTableMapper;
    private final SchemaColumnMapper schemaColumnMapper;
    private final SchemaForeignKeyMapper schemaForeignKeyMapper;
    private final SchemaIndexMapper schemaIndexMapper;
    private final StringEncryptor stringEncryptor;

    /** 连接器（无状态单例，由 UdapConfig 声明并构造注入） */
    private final Connector connector;

    /** 分页查询数据源，支持编码/名称模糊、类型/状态精确筛选 */
    @Override
    @Transactional(readOnly = true)
    public PageResult<DatasourceVO> page(DatasourceQuery query) {
        QueryWrapper qw = QueryWrapper.create()
                .where(DATASOURCE.CODE.like(query.getCode()).when(StringUtils.hasText(query.getCode())))
                .and(DATASOURCE.NAME.like(query.getName()).when(StringUtils.hasText(query.getName())))
                .and(DATASOURCE.TYPE.eq(query.getType()).when(StringUtils.hasText(query.getType())))
                .and(DATASOURCE.STATUS.eq(query.getStatus()).when(query.getStatus() != null))
                .orderBy(DATASOURCE.CREATE_TIME.desc());

        Page<Datasource> page = datasourceMapper.paginate(Page.of(query.getCurrent(), query.getSize()), qw);
        List<DatasourceVO> records = page.getRecords().stream()
                .map(DatasourceConverter::toVO)
                .collect(Collectors.toList());
        return PageResult.of(records, page.getTotalRow(), page.getPageNumber(), page.getPageSize());
    }

    /** 查询全部数据源（不分页，仅返回启用状态） */
    @Override
    @Transactional(readOnly = true)
    public List<DatasourceVO> listAll() {
        List<Datasource> list = datasourceMapper.selectListByQuery(
                QueryWrapper.create().orderBy(DATASOURCE.CREATE_TIME.desc()));
        return list.stream()
                .map(DatasourceConverter::toVO)
                .collect(Collectors.toList());
    }

    /** 根据ID查询数据源详情，不存在时抛业务异常 */
    @Override
    @Transactional(readOnly = true)
    public DatasourceVO getById(Long id) {
        return DatasourceConverter.toVO(loadEntity(id));
    }

    /** 新增数据源：校验编码唯一，密码加密后入库 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public Long create(DatasourceDTO dto) {
        Datasource existing = datasourceMapper.selectOneByQuery(
                QueryWrapper.create().where(DATASOURCE.CODE.eq(dto.getCode())));
        if (existing != null) {
            throw new BusinessException(DatasourceErrorCode.DATASOURCE_CODE_EXISTS);
        }
        Datasource entity = new Datasource();
        copyDtoToEntity(dto, entity, true);
        datasourceMapper.insert(entity);
        return entity.getId();
    }

    /** 更新数据源：先查后改，密码留空则保持原值 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void update(DatasourceDTO dto) {
        Datasource entity = loadEntity(dto.getId());
        // 编码唯一性校验（排除自身）
        Datasource sameCode = datasourceMapper.selectOneByQuery(
                QueryWrapper.create().where(DATASOURCE.CODE.eq(dto.getCode())));
        if (sameCode != null && !sameCode.getId().equals(dto.getId())) {
            throw new BusinessException(DatasourceErrorCode.DATASOURCE_CODE_EXISTS);
        }
        boolean changePassword = StringUtils.hasText(dto.getPassword());
        copyDtoToEntity(dto, entity, changePassword);
        datasourceMapper.update(entity);
    }

    /** 删除数据源，并级联删除其测试记录与元数据 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public void delete(Long id) {
        loadEntity(id);
        // 删除元数据（列/外键/索引依赖表ID，批量按 tableId 集合删除）
        List<SchemaTable> tables = schemaTableMapper.selectListByQuery(
                QueryWrapper.create().where(SCHEMA_TABLE.DATASOURCE_ID.eq(id)));
        deleteChildrenByTables(tables);
        schemaTableMapper.deleteByQuery(QueryWrapper.create().where(SCHEMA_TABLE.DATASOURCE_ID.eq(id)));
        datasourceTestMapper.deleteByQuery(QueryWrapper.create().where(DATASOURCE_TEST.DATASOURCE_ID.eq(id)));
        datasourceMapper.deleteById(id);
    }

    /** 测试连接：解密密码→建连→记录 odin_datasource_test→返回结果。远程内省不纳入事务，仅单条测试记录自动提交 */
    @Override
    public DatasourceTestResultVO testConnection(Long id) {
        Datasource ds = loadEntity(id);
        ConnectorConfig config = toConnectorConfig(ds);
        TestResult result = connector.testConnection(config);

        DatasourceTest record = new DatasourceTest();
        record.setDatasourceId(id);
        record.setTestTime(LocalDateTime.now());
        record.setSuccess(result.isSuccess() ? 1 : 0);
        record.setLatencyMs(result.getLatencyMs() == null ? null : result.getLatencyMs().intValue());
        record.setDbVersion(result.getDbVersion());
        record.setErrorMessage(result.getErrorMessage());
        datasourceTestMapper.insert(record);

        DatasourceTestResultVO vo = new DatasourceTestResultVO();
        vo.setSuccess(result.isSuccess());
        vo.setLatencyMs(result.getLatencyMs());
        vo.setDbVersion(result.getDbVersion());
        vo.setErrorMessage(result.getErrorMessage());
        return vo;
    }

    /** 同步元数据：内省表→逐表内省列/外键/索引→全量覆盖式落库，返回表数量 */
    @Override
    @Transactional(rollbackFor = Exception.class)
    public int syncMetadata(Long id) {
        Datasource ds = loadEntity(id);
        ConnectorConfig config = toConnectorConfig(ds);

        // 清理旧元数据（批量按 tableId 集合删除）
        List<SchemaTable> oldTables = schemaTableMapper.selectListByQuery(
                QueryWrapper.create().where(SCHEMA_TABLE.DATASOURCE_ID.eq(id)));
        deleteChildrenByTables(oldTables);
        schemaTableMapper.deleteByQuery(QueryWrapper.create().where(SCHEMA_TABLE.DATASOURCE_ID.eq(id)));

        List<TableMeta> tables = connector.listTables(config);
        LocalDateTime now = LocalDateTime.now();
        for (TableMeta tm : tables) {
            // 先内省列，用于计算结构指纹
            List<ColumnMeta> columnMetas = connector.listColumns(config, tm.getSchemaName(), tm.getTableName());

            SchemaTable st = new SchemaTable();
            st.setDatasourceId(id);
            st.setSchemaName(tm.getSchemaName());
            st.setTableName(tm.getTableName());
            st.setTableType(tm.getTableType());
            st.setComment(tm.getComment());
            // TODO 行数估计：后续按方言从 information_schema / 统计信息估算，避免每次同步执行全表 COUNT
            st.setRowEstimate(null);
            st.setFingerprint(computeFingerprint(columnMetas));
            st.setLastSyncTime(now);
            st.setStatus(1);
            schemaTableMapper.insert(st);
            Long tableId = st.getId();

            // 列（聚合后批量写入）
            List<SchemaColumn> columnRows = new ArrayList<>();
            for (ColumnMeta cm : columnMetas) {
                SchemaColumn sc = new SchemaColumn();
                sc.setTableId(tableId);
                sc.setColumnName(cm.getColumnName());
                sc.setDataType(cm.getDataType());
                sc.setCanonicalType(resolveCanonicalType(cm.getDataType()));
                sc.setNullable(cm.isNullable() ? 1 : 0);
                sc.setDefaultValue(cm.getDefaultValue());
                sc.setComment(cm.getComment());
                sc.setMaxLength(cm.getMaxLength());
                sc.setNumericPrecision(cm.getNumericPrecision());
                sc.setNumericScale(cm.getNumericScale());
                sc.setOrdinalPosition(cm.getOrdinalPosition());
                sc.setIsPrimaryKey(cm.isPrimaryKey() ? 1 : 0);
                sc.setIsUnique(0);
                sc.setStatus(1);
                columnRows.add(sc);
            }
            if (!columnRows.isEmpty()) {
                schemaColumnMapper.insertBatch(columnRows);
            }
            // 外键（聚合后批量写入）
            List<SchemaForeignKey> fkRows = new ArrayList<>();
            for (ForeignKeyMeta fk : connector.listForeignKeys(config, tm.getSchemaName(), tm.getTableName())) {
                SchemaForeignKey e = new SchemaForeignKey();
                e.setTableId(tableId);
                e.setConstraintName(fk.getConstraintName());
                e.setColumnName(fk.getColumnName());
                e.setRefSchema(fk.getRefSchema());
                e.setRefTable(fk.getRefTable());
                e.setRefColumn(fk.getRefColumn());
                e.setOnDelete(fk.getDeleteRule());
                e.setOnUpdate(fk.getUpdateRule());
                fkRows.add(e);
            }
            if (!fkRows.isEmpty()) {
                schemaForeignKeyMapper.insertBatch(fkRows);
            }
            // 索引（聚合后批量写入）
            List<SchemaIndex> indexRows = new ArrayList<>();
            for (IndexMeta im : connector.listIndexes(config, tm.getSchemaName(), tm.getTableName())) {
                SchemaIndex si = new SchemaIndex();
                si.setTableId(tableId);
                si.setIndexName(im.getIndexName());
                si.setColumnNames(toJsonArray(im.getColumnNames()));
                si.setIsUnique(im.isUnique() ? 1 : 0);
                si.setIndexType(im.getIndexType() == null ? "BTREE" : im.getIndexType());
                indexRows.add(si);
            }
            if (!indexRows.isEmpty()) {
                schemaIndexMapper.insertBatch(indexRows);
            }
        }
        return tables.size();
    }

    /** 浏览：查询数据源已同步的表清单 */
    @Override
    @Transactional(readOnly = true)
    public List<SchemaTableVO> listTables(Long datasourceId) {
        List<SchemaTable> tables = schemaTableMapper.selectListByQuery(QueryWrapper.create()
                .where(SCHEMA_TABLE.DATASOURCE_ID.eq(datasourceId))
                .orderBy(SCHEMA_TABLE.TABLE_NAME.asc()));
        return tables.stream()
                .map(t -> BeanCopyUtil.copyProperties(t, SchemaTableVO.class))
                .collect(Collectors.toList());
    }

    /** 浏览：查询表已同步的列清单 */
    @Override
    @Transactional(readOnly = true)
    public List<SchemaColumnVO> listColumns(Long tableId) {
        List<SchemaColumn> columns = schemaColumnMapper.selectListByQuery(QueryWrapper.create()
                .where(SCHEMA_COLUMN.TABLE_ID.eq(tableId))
                .orderBy(SCHEMA_COLUMN.ORDINAL_POSITION.asc()));
        return columns.stream()
                .map(c -> BeanCopyUtil.copyProperties(c, SchemaColumnVO.class))
                .collect(Collectors.toList());
    }

    /**
     * 加载数据源实体，不存在抛业务异常。
     *
     * @param id 主键ID
     * @return 数据源实体
     */
    private Datasource loadEntity(Long id) {
        Datasource ds = datasourceMapper.selectOneById(id);
        if (ds == null) {
            throw new BusinessException(DatasourceErrorCode.DATASOURCE_NOT_FOUND);
        }
        return ds;
    }

    /**
     * 批量删除一组表的子元数据（列/外键/索引），按 tableId 集合一次性删除，避免逐表 N+1。
     *
     * @param tables 表元数据列表
     */
    private void deleteChildrenByTables(List<SchemaTable> tables) {
        if (tables == null || tables.isEmpty()) {
            return;
        }
        List<Long> tableIds = tables.stream().map(SchemaTable::getId).collect(Collectors.toList());
        schemaColumnMapper.deleteByQuery(QueryWrapper.create().where(SCHEMA_COLUMN.TABLE_ID.in(tableIds)));
        schemaForeignKeyMapper.deleteByQuery(QueryWrapper.create().where(SCHEMA_FOREIGN_KEY.TABLE_ID.in(tableIds)));
        schemaIndexMapper.deleteByQuery(QueryWrapper.create().where(SCHEMA_INDEX.TABLE_ID.in(tableIds)));
    }

    /**
     * 将 DTO 拷贝到实体；密码非空时加密写入 passwordEnc。
     *
     * @param dto            入参
     * @param entity         目标实体
     * @param changePassword 是否更新密码
     */
    private void copyDtoToEntity(DatasourceDTO dto, Datasource entity, boolean changePassword) {
        entity.setCode(dto.getCode());
        entity.setName(dto.getName());
        entity.setType(dto.getType());
        entity.setHost(dto.getHost());
        entity.setPort(dto.getPort());
        entity.setDbName(dto.getDbName());
        entity.setSchemaName(dto.getSchemaName());
        entity.setUsername(dto.getUsername());
        entity.setJdbcUrl(dto.getJdbcUrl());
        entity.setStatus(dto.getStatus() == null ? 1 : dto.getStatus());
        entity.setRemark(dto.getRemark());
        if (changePassword && StringUtils.hasText(dto.getPassword())) {
            entity.setPasswordEnc(stringEncryptor.encrypt(dto.getPassword()));
        }
        // 扩展字段（连接选项/标签/连接池/超时）：更新时 DTO 为 null 则保留原值不覆盖
        if (dto.getOptions() != null) {
            entity.setOptions(dto.getOptions());
        }
        if (dto.getTags() != null) {
            entity.setTags(dto.getTags());
        }
        if (dto.getPoolMaxSize() != null) {
            entity.setPoolMaxSize(dto.getPoolMaxSize());
        }
        if (dto.getPoolMinIdle() != null) {
            entity.setPoolMinIdle(dto.getPoolMinIdle());
        }
        if (dto.getPoolMaxLifeS() != null) {
            entity.setPoolMaxLifeS(dto.getPoolMaxLifeS());
        }
        if (dto.getConnTimeoutS() != null) {
            entity.setConnTimeoutS(dto.getConnTimeoutS());
        }
        if (dto.getQueryTimeoutS() != null) {
            entity.setQueryTimeoutS(dto.getQueryTimeoutS());
        }
        // 连接池/超时默认值
        if (entity.getPoolMaxSize() == null) {
            entity.setPoolMaxSize(10);
        }
        if (entity.getPoolMinIdle() == null) {
            entity.setPoolMinIdle(2);
        }
        if (entity.getPoolMaxLifeS() == null) {
            entity.setPoolMaxLifeS(3600);
        }
        if (entity.getConnTimeoutS() == null) {
            entity.setConnTimeoutS(30);
        }
        if (entity.getQueryTimeoutS() == null) {
            entity.setQueryTimeoutS(60);
        }
    }

    /**
     * 构建指定数据源的连接器配置（含解密后的明文密码），供联邦执行/问数直查复用。
     *
     * @param id 数据源ID
     * @return 连接器配置
     */
    @Override
    public ConnectorConfig getConnectorConfig(Long id) {
        return toConnectorConfig(loadEntity(id));
    }

    /**
     * 由数据源实体构造连接器配置（解密密码）。
     *
     * @param ds 数据源实体
     * @return 连接器配置
     */
    private ConnectorConfig toConnectorConfig(Datasource ds) {
        ConnectorConfig config = new ConnectorConfig();
        config.setType(ds.getType());
        config.setHost(ds.getHost());
        config.setPort(ds.getPort());
        config.setDbName(ds.getDbName());
        config.setSchemaName(ds.getSchemaName());
        config.setUsername(ds.getUsername());
        config.setPassword(decryptPassword(ds));
        config.setJdbcUrl(ds.getJdbcUrl());
        config.setConnTimeoutS(ds.getConnTimeoutS() == null ? 30 : ds.getConnTimeoutS());
        config.setQueryTimeoutS(ds.getQueryTimeoutS() == null ? 60 : ds.getQueryTimeoutS());
        return config;
    }

    /**
     * 解密数据源密码，兼容种子/历史明文数据。
     *
     * <p>{@code password_enc} 正常应存放 Jasypt 密文，但示例种子数据可能直接写入明文，
     * 或加密密钥（{@code JASYPT_ENCRYPTOR_PASSWORD}）变更导致旧密文无法解密。此处优先按密文解密，
     * 失败时回退为原值（视为明文）并记录告警，避免单条脏数据抛
     * {@link EncryptionOperationNotPossibleException} 使测试连接 / 元数据同步整体返回 500。
     *
     * @param ds 数据源实体
     * @return 明文密码；未配置密码时返回 {@code null}
     */
    private String decryptPassword(Datasource ds) {
        String stored = ds.getPasswordEnc();
        if (!StringUtils.hasText(stored)) {
            return null;
        }
        try {
            return stringEncryptor.decrypt(stored);
        } catch (EncryptionOperationNotPossibleException e) {
            log.warn("数据源[id={}, code={}] 的 password_enc 非合法密文（可能为种子明文或加密密钥已变更），"
                    + "已按明文回退；建议重新编辑并保存该数据源以正确加密密码", ds.getId(), ds.getCode());
            return stored;
        }
    }

    /**
     * 将字符串列表序列化为 JSON 数组字符串（用于索引列名存储）。
     *
     * @param list 列名列表
     * @return JSON 数组字符串
     */
    private String toJsonArray(List<String> list) {
        if (list == null || list.isEmpty()) {
            return "[]";
        }
        List<String> quoted = new ArrayList<>();
        for (String s : list) {
            quoted.add("\"" + s.replace("\"", "\\\"") + "\"");
        }
        return "[" + String.join(",", quoted) + "]";
    }

    /**
     * 将源库原始类型归一为内部规范类型。
     *
     * @param dataType 源库类型名
     * @return 归一类型（int/decimal/bool/date/datetime/json/string）
     */
    private String resolveCanonicalType(String dataType) {
        if (dataType == null) {
            return "string";
        }
        String lower = dataType.toLowerCase();
        if (lower.contains("int") || lower.contains("serial")) {
            return "int";
        }
        if (lower.contains("decimal") || lower.contains("numeric") || lower.contains("float")
                || lower.contains("double") || lower.contains("real")) {
            return "decimal";
        }
        if (lower.contains("bool") || lower.contains("bit")) {
            return "bool";
        }
        if (lower.contains("date") && !lower.contains("time")) {
            return "date";
        }
        if (lower.contains("time") || lower.contains("timestamp")) {
            return "datetime";
        }
        if (lower.contains("json")) {
            return "json";
        }
        return "string";
    }

    /**
     * 基于列定义计算表结构指纹（按列名排序后的 列名:类型 拼接取 MD5）。
     *
     * @param columns 列元数据
     * @return 32 位十六进制 MD5 指纹
     */
    private String computeFingerprint(List<ColumnMeta> columns) {
        if (columns == null || columns.isEmpty()) {
            return DigestUtils.md5DigestAsHex("".getBytes(StandardCharsets.UTF_8));
        }
        String raw = columns.stream()
                .sorted(Comparator.comparing(ColumnMeta::getColumnName))
                .map(c -> c.getColumnName() + ":" + c.getDataType())
                .collect(Collectors.joining(","));
        return DigestUtils.md5DigestAsHex(raw.getBytes(StandardCharsets.UTF_8));
    }
}
