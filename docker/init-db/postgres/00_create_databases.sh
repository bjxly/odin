#!/bin/bash
# ============================================================
# ODIN 演示库建库脚本（PostgreSQL 17）
# 容器：odin-pgsql
# docker-entrypoint-initdb.d 中的 .sh 仅在数据卷为空时执行一次，
# 且执行时 POSTGRES_DB(demo) 已创建完毕。
#
# 本脚本创建三个演示库并导入对应初始化 SQL：
#   crm  → 01_crm_init.sql        客户主数据
#   erp  → 02_erp_init.sql        产品 / 销售订单 / 舰艇 / 货运
#   wms  → 03_wms_init.sql        仓库库存 / 物流运单 / 供应商
# 三个库分别注册为 ODIN 的三个演示数据源，用于跨源联邦查询验证。
# ============================================================
set -euo pipefail

PSQL=(psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER")

for db in crm erp wms; do
  "${PSQL[@]}" --dbname "$POSTGRES_DB" <<-EOSQL
    DROP DATABASE IF EXISTS ${db};
    CREATE DATABASE ${db} OWNER "$POSTGRES_USER" ENCODING 'UTF8' TEMPLATE template0;
EOSQL
done

"${PSQL[@]}" --dbname crm -f /docker-entrypoint-initdb.d/01_crm_init.sql
"${PSQL[@]}" --dbname erp -f /docker-entrypoint-initdb.d/02_erp_init.sql
"${PSQL[@]}" --dbname wms -f /docker-entrypoint-initdb.d/03_wms_init.sql

echo "ODIN demo databases created: crm, erp, wms"
