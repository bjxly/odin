package seed

import (
	"gorm.io/gorm"

	"smartg-odin/internal/common/logger"
	"smartg-odin/internal/model"
)

// propSpec 描述一个本体类属性。
type propSpec struct {
	Name     string
	Label    string
	DataType string
	Required bool
	Synonyms string // 同义词（JSON 数组串或逗号分隔），供规则/LLM 语义匹配
}

// seedOntologies 写入全部本体（inventory + naval）。
func seedOntologies(db *gorm.DB, sourceIDs map[string]uint) {
	seedInventoryOntology(db, sourceIDs)
	seedNavalOntology(db, sourceIDs)
}

// ---- 本体 1：inventory ----

func seedInventoryOntology(db *gorm.DB, sourceIDs map[string]uint) {
	ont := model.OntDefinition{
		Code:        "inventory",
		Name:        "库存与销售分析本体",
		Description: "跨 MySQL(ERP) 与 PostgreSQL(仓储物流) 的客户/订单/产品/库存/运单/供应商本体，含 VIP 客户与大额订单两个规则虚拟类，用于演示本体驱动的跨源虚拟集成查询与推理。",
		Version:     "2.0.0",
		Status:      "published",
	}
	if err := db.Create(&ont).Error; err != nil {
		logger.L.Errorf("seed ontology inventory failed: %v", err)
		return
	}
	ontID := ont.ID

	mysqlID := sourceIDs["crm_mysql"] // 客户 / 订单 / 产品
	pgID := sourceIDs["erp_postgres"] // 库存 / 运单 / 供应商

	// ---- 类与属性 ----
	customer, customerProps := createClass(db, ontID, "Customer", "客户", "企业客户主数据（MySQL smartj.customers）", "normal", nil, 80, 220, []propSpec{
		{Name: "name", Label: "客户名称", DataType: "string"},
		{Name: "email", Label: "邮箱", DataType: "string"},
		{Name: "phone", Label: "电话", DataType: "string"},
		{Name: "level", Label: "客户等级", DataType: "string"},
		{Name: "totalOrders", Label: "累计订单数", DataType: "integer"},
		{Name: "totalAmount", Label: "累计金额", DataType: "float"},
		{Name: "address", Label: "地址", DataType: "string"},
		{Name: "createdAt", Label: "创建时间", DataType: "datetime"},
	})
	order, orderProps := createClass(db, ontID, "Order", "订单", "销售订单（MySQL smartj.orders）", "normal", nil, 400, 220, []propSpec{
		{Name: "orderNo", Label: "订单号", DataType: "string", Synonyms: "订单编号,单号"},
		{Name: "quantity", Label: "数量", DataType: "integer", Synonyms: "销量,销售数量,销售量,购买数量"},
		{Name: "unitPrice", Label: "单价", DataType: "float", Synonyms: "价格"},
		{Name: "amount", Label: "金额", DataType: "float", Synonyms: "销售额,销售金额,总金额,总价,订单金额"},
		{Name: "status", Label: "状态", DataType: "string"},
		{Name: "orderDate", Label: "下单时间", DataType: "datetime", Synonyms: "下单日期,订单日期"},
		{Name: "shipDate", Label: "发货时间", DataType: "datetime"},
		{Name: "remark", Label: "备注", DataType: "string"},
	})
	product, productProps := createClass(db, ontID, "Product", "产品", "产品目录（MySQL smartj.products）", "normal", nil, 720, 220, []propSpec{
		{Name: "name", Label: "产品名称", DataType: "string", Synonyms: "产品,品名"},
		{Name: "sku", Label: "SKU", DataType: "string", Synonyms: "产品编号"},
		{Name: "category", Label: "品类", DataType: "string", Synonyms: "类别,分类"},
		{Name: "price", Label: "单价", DataType: "float", Synonyms: "价格"},
		{Name: "stock", Label: "库存", DataType: "integer", Synonyms: "库存量,存货"},
		{Name: "supplier", Label: "供应商", DataType: "string"},
		{Name: "description", Label: "描述", DataType: "string"},
	})
	warehouseInv, warehouseInvProps := createClass(db, ontID, "WarehouseInventory", "仓库库存", "实时仓库库存（PostgreSQL demo.warehouse_inventory）", "normal", nil, 1040, 220, []propSpec{
		{Name: "productSku", Label: "产品SKU", DataType: "string"},
		{Name: "productName", Label: "产品名称", DataType: "string"},
		{Name: "warehouse", Label: "仓库", DataType: "string"},
		{Name: "quantity", Label: "库存数量", DataType: "integer"},
		{Name: "reserved", Label: "预留数量", DataType: "integer"},
		{Name: "available", Label: "可用数量", DataType: "integer"},
		{Name: "minStock", Label: "最低库存", DataType: "integer"},
		{Name: "maxStock", Label: "最高库存", DataType: "integer"},
		{Name: "lastRestockDate", Label: "最近补货日期", DataType: "datetime"},
	})
	shipment, shipmentProps := createClass(db, ontID, "Shipment", "物流运单", "物流运单（PostgreSQL demo.shipments）", "normal", nil, 400, 460, []propSpec{
		{Name: "shipmentNo", Label: "运单号", DataType: "string"},
		{Name: "orderNo", Label: "关联订单号", DataType: "string"},
		{Name: "carrier", Label: "承运商", DataType: "string"},
		{Name: "trackingNo", Label: "物流单号", DataType: "string"},
		{Name: "status", Label: "状态", DataType: "string"},
		{Name: "shipDate", Label: "发货日期", DataType: "datetime"},
		{Name: "deliveryDate", Label: "送达日期", DataType: "datetime"},
		{Name: "originWarehouse", Label: "发货仓库", DataType: "string"},
		{Name: "destination", Label: "目的地", DataType: "string"},
		{Name: "weightKg", Label: "重量(kg)", DataType: "float"},
	})
	supplier, supplierProps := createClass(db, ontID, "Supplier", "供应商", "供应商档案（PostgreSQL demo.suppliers）", "normal", nil, 720, 460, []propSpec{
		{Name: "name", Label: "供应商名称", DataType: "string"},
		{Name: "contactPerson", Label: "联系人", DataType: "string"},
		{Name: "phone", Label: "电话", DataType: "string"},
		{Name: "email", Label: "邮箱", DataType: "string"},
		{Name: "address", Label: "地址", DataType: "string"},
		{Name: "rating", Label: "评分", DataType: "float"},
		{Name: "cooperationYears", Label: "合作年限", DataType: "integer"},
		{Name: "status", Label: "状态", DataType: "string"},
	})
	vip, _ := createClass(db, ontID, "VipCustomer", "VIP客户", "规则虚拟类：level=VIP 的客户", "virtual", &customer.ID, 80, 40, nil)
	big, _ := createClass(db, ontID, "BigOrder", "大额订单", "规则虚拟类：amount>50000 的订单", "virtual", &order.ID, 400, 40, nil)

	// ---- 关系 ----
	createRelation(db, ontID, "places", "下单", customer.ID, order.ID, "association", "1:N",
		`{"from_column":"id","to_column":"customer_id","join_type":"LEFT"}`)
	createRelation(db, ontID, "contains", "包含产品", order.ID, product.ID, "association", "N:1",
		`{"from_column":"product_id","to_column":"id","join_type":"LEFT"}`)
	// 反向关系 Product→Order（供「产品维度聚合订单度量」的跨类聚合，如产品销量 TOP-N）。
	createRelation(db, ontID, "orderedIn", "对应订单", product.ID, order.ID, "association", "1:N",
		`{"from_column":"id","to_column":"product_id","join_type":"LEFT"}`)
	createRelation(db, ontID, "ships", "发运", order.ID, shipment.ID, "association", "1:N",
		`{"from_column":"order_no","to_column":"order_no","join_type":"LEFT"}`)
	createRelation(db, ontID, "stores", "库存于", product.ID, warehouseInv.ID, "association", "1:N",
		`{"from_column":"sku","to_column":"product_sku","join_type":"LEFT"}`)
	createRelation(db, ontID, "suppliedBy", "供应自", product.ID, supplier.ID, "association", "N:1",
		`{"from_column":"supplier","to_column":"name","join_type":"LEFT"}`)
	createRelation(db, ontID, "classifies", "分类为", customer.ID, vip.ID, "inheritance", "1:N", "")

	// ---- 规则 + 动作 ----
	vipRule := createRule(db, ontID, "vip_customer_rule", "VIP客户派生规则", "derivation",
		`{"type":"value_match","property":"level","operator":"eq","value":"VIP"}`,
		`{"type":"expand_class","target_class":"VipCustomer"}`, 10, true)
	createAction(db, vipRule.ID, "expand_class", &vip.ID, "", "")

	bigRule := createRule(db, ontID, "big_order_rule", "大额订单派生规则", "derivation",
		`{"type":"threshold","property":"amount","operator":"gt","value":50000}`,
		`{"type":"expand_class","target_class":"BigOrder"}`, 20, true)
	createAction(db, bigRule.ID, "expand_class", &big.ID, "", "")

	// ---- 映射配置 ----
	customerPairs := [][2]string{
		{"name", "name"}, {"email", "email"}, {"phone", "phone"}, {"level", "level"},
		{"totalOrders", "total_orders"}, {"totalAmount", "total_amount"}, {"address", "address"}, {"createdAt", "created_at"},
	}
	orderPairs := [][2]string{
		{"orderNo", "order_no"}, {"quantity", "quantity"}, {"unitPrice", "unit_price"}, {"amount", "amount"},
		{"status", "status"}, {"orderDate", "order_date"}, {"shipDate", "ship_date"}, {"remark", "remark"},
	}
	productPairs := [][2]string{
		{"name", "name"}, {"sku", "sku"}, {"category", "category"}, {"price", "price"},
		{"stock", "stock"}, {"supplier", "supplier"}, {"description", "description"},
	}
	warehouseInvPairs := [][2]string{
		{"productSku", "product_sku"}, {"productName", "product_name"}, {"warehouse", "warehouse"},
		{"quantity", "quantity"}, {"reserved", "reserved"}, {"available", "available"},
		{"minStock", "min_stock"}, {"maxStock", "max_stock"}, {"lastRestockDate", "last_restock_date"},
	}
	shipmentPairs := [][2]string{
		{"shipmentNo", "shipment_no"}, {"orderNo", "order_no"}, {"carrier", "carrier"}, {"trackingNo", "tracking_no"},
		{"status", "status"}, {"shipDate", "ship_date"}, {"deliveryDate", "delivery_date"},
		{"originWarehouse", "origin_warehouse"}, {"destination", "destination"}, {"weightKg", "weight_kg"},
	}
	supplierPairs := [][2]string{
		{"name", "name"}, {"contactPerson", "contact_person"}, {"phone", "phone"}, {"email", "email"},
		{"address", "address"}, {"rating", "rating"}, {"cooperationYears", "cooperation_years"}, {"status", "status"},
	}

	// MySQL：客户 / 订单 / 产品
	createMapping(db, ontID, customer.ID, mysqlID, "customers", buildMappings(customerProps, customerPairs))
	createMapping(db, ontID, order.ID, mysqlID, "orders", buildMappings(orderProps, orderPairs))
	createMapping(db, ontID, product.ID, mysqlID, "products", buildMappings(productProps, productPairs))
	// PostgreSQL：库存 / 运单 / 供应商
	createMapping(db, ontID, warehouseInv.ID, pgID, "warehouse_inventory", buildMappings(warehouseInvProps, warehouseInvPairs))
	createMapping(db, ontID, shipment.ID, pgID, "shipments", buildMappings(shipmentProps, shipmentPairs))
	createMapping(db, ontID, supplier.ID, pgID, "suppliers", buildMappings(supplierProps, supplierPairs))
	// 虚拟类复用父类的物理映射（附加过滤由规则注入）。
	createMapping(db, ontID, vip.ID, mysqlID, "customers", buildMappings(customerProps, customerPairs))
	createMapping(db, ontID, big.ID, mysqlID, "orders", buildMappings(orderProps, orderPairs))
}

// ---- 本体 2：naval ----

func seedNavalOntology(db *gorm.DB, sourceIDs map[string]uint) {
	ont := model.OntDefinition{
		Code:        "naval",
		Name:        "海军舰艇本体",
		Description: "舰艇与货运本体（MySQL smartj.vessels / smartj.shipments），用于演示跨源关联查询。",
		Version:     "2.0.0",
		Status:      "published",
	}
	if err := db.Create(&ont).Error; err != nil {
		logger.L.Errorf("seed ontology naval failed: %v", err)
		return
	}
	ontID := ont.ID

	vessel, vesselProps := createClass(db, ontID, "Vessel", "舰艇", "舰艇档案（MySQL smartj.vessels）", "normal", nil, 80, 160, []propSpec{
		{Name: "name", Label: "舰名", DataType: "string"},
		{Name: "type", Label: "舰型", DataType: "string"},
		{Name: "capacity", Label: "载重吨", DataType: "float"},
		{Name: "flag", Label: "船旗国", DataType: "string"},
	})
	shipment, shipmentProps := createClass(db, ontID, "Shipment", "货运", "货运记录（MySQL smartj.shipments）", "normal", nil, 380, 160, []propSpec{
		{Name: "shipmentNo", Label: "运单号", DataType: "string"},
		{Name: "carrier", Label: "承运方", DataType: "string"},
		{Name: "cargo", Label: "货物", DataType: "string"},
		{Name: "status", Label: "状态", DataType: "string"},
		{Name: "shipDate", Label: "发货日期", DataType: "datetime"},
		{Name: "deliveryDate", Label: "送达日期", DataType: "datetime"},
		{Name: "originPort", Label: "起运港", DataType: "string"},
		{Name: "destinationPort", Label: "目的港", DataType: "string"},
	})

	createRelation(db, ontID, "carries", "承运", vessel.ID, shipment.ID, "association", "1:N",
		`{"from_column":"id","to_column":"vessel_id","join_type":"LEFT"}`)

	// naval 本体数据已从 analytics_sqlite 迁移至 MySQL（crm_mysql）。
	mysqlID := sourceIDs["crm_mysql"]
	vesselPairs := [][2]string{{"name", "name"}, {"type", "type"}, {"capacity", "capacity"}, {"flag", "flag"}}
	shipmentPairs := [][2]string{
		{"shipmentNo", "shipment_no"}, {"carrier", "carrier"}, {"cargo", "cargo"}, {"status", "status"},
		{"shipDate", "ship_date"}, {"deliveryDate", "delivery_date"},
		{"originPort", "origin_port"}, {"destinationPort", "destination_port"},
	}

	createMapping(db, ontID, vessel.ID, mysqlID, "vessels", buildMappings(vesselProps, vesselPairs))
	createMapping(db, ontID, shipment.ID, mysqlID, "shipments", buildMappings(shipmentProps, shipmentPairs))
}

// ---- 通用写入辅助 ----

// createClass 创建一个本体类及其属性，返回类与「属性名→属性ID」映射。
func createClass(db *gorm.DB, ontID uint, name, label, desc, classType string, parentID *uint, x, y float64, props []propSpec) (model.OntClass, map[string]uint) {
	cls := model.OntClass{
		OntologyID:    ontID,
		Name:          name,
		Label:         label,
		Description:   desc,
		ClassType:     classType,
		ParentClassID: parentID,
		PositionX:     x,
		PositionY:     y,
	}
	if err := db.Create(&cls).Error; err != nil {
		logger.L.Errorf("seed class %s failed: %v", name, err)
	}

	propIDs := make(map[string]uint, len(props))
	for i, p := range props {
		prop := model.OntClassProperty{
			ClassID:    cls.ID,
			OntologyID: ontID,
			Name:       p.Name,
			Label:      p.Label,
			DataType:   p.DataType,
			Required:   p.Required,
			Synonyms:   p.Synonyms,
			SortOrder:  i + 1,
		}
		if err := db.Create(&prop).Error; err != nil {
			logger.L.Errorf("seed property %s.%s failed: %v", name, p.Name, err)
			continue
		}
		propIDs[p.Name] = prop.ID
	}
	return cls, propIDs
}

// createRelation 创建本体关系。
func createRelation(db *gorm.DB, ontID uint, name, label string, fromID, toID uint, relType, cardinality, joinJSON string) {
	rel := model.OntRelation{
		OntologyID:        ontID,
		Name:              name,
		Label:             label,
		FromClassID:       fromID,
		ToClassID:         toID,
		RelationType:      relType,
		Cardinality:       cardinality,
		JoinConditionJSON: joinJSON,
	}
	if err := db.Create(&rel).Error; err != nil {
		logger.L.Errorf("seed relation %s failed: %v", name, err)
	}
}

// createRule 创建推理规则并返回。
func createRule(db *gorm.DB, ontID uint, name, desc, ruleType, condJSON, actJSON string, priority int, enabled bool) model.OntRule {
	r := model.OntRule{
		OntologyID:    ontID,
		Name:          name,
		Description:   desc,
		RuleType:      ruleType,
		ConditionJSON: condJSON,
		ActionJSON:    actJSON,
		Priority:      priority,
		Enabled:       enabled,
	}
	if err := db.Create(&r).Error; err != nil {
		logger.L.Errorf("seed rule %s failed: %v", name, err)
	}
	return r
}

// createAction 创建规则动作。
func createAction(db *gorm.DB, ruleID uint, actionType string, targetClassID *uint, targetProp, valueExpr string) {
	a := model.OntAction{
		RuleID:          ruleID,
		ActionType:      actionType,
		TargetClassID:   targetClassID,
		TargetProperty:  targetProp,
		ValueExpression: valueExpr,
	}
	if err := db.Create(&a).Error; err != nil {
		logger.L.Errorf("seed action for rule %d failed: %v", ruleID, err)
	}
}

// createMapping 创建类到物理表的映射配置。
func createMapping(db *gorm.DB, ontID, classID, dsID uint, table string, mappings []propertyMapping) {
	mc := model.OntMappingConfig{
		OntologyID:           ontID,
		ClassID:              classID,
		DataSourceID:         dsID,
		SourceTable:          table,
		PropertyMappingsJSON: mustJSON(mappings),
		Status:               "active",
	}
	if err := db.Create(&mc).Error; err != nil {
		logger.L.Errorf("seed mapping for class %d failed: %v", classID, err)
	}
}

// buildMappings 依据「属性名→列名」对与属性 ID 映射构建 propertyMapping 列表。
func buildMappings(propIDs map[string]uint, pairs [][2]string) []propertyMapping {
	out := make([]propertyMapping, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, propertyMapping{
			PropertyID:       propIDs[p[0]],
			PropertyName:     p[0],
			ColumnName:       p[1],
			Confidence:       1.0,
			ValidationStatus: "validated",
		})
	}
	return out
}
