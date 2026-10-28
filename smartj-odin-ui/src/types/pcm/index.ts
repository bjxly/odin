// ==================== Customer ====================
export interface CustomerVO {
  id: string
  customerName: string
  customerDesc?: string
  firstCooperateDate?: string
  taxNo?: string
  registeredAddress?: string
  registeredPhone?: string
  bankName?: string
  bankAccount?: string
  contactName?: string
  contactPhone?: string
  status: number
  remark?: string
  createTime: string
}

export interface CustomerDTO {
  id?: string
  customerName: string
  customerDesc?: string
  firstCooperateDate?: string
  taxNo?: string
  registeredAddress?: string
  registeredPhone?: string
  bankName?: string
  bankAccount?: string
  contactName?: string
  contactPhone?: string
  remark?: string
}

export interface CustomerQuery {
  current: number
  size: number
  customerName?: string
  contactName?: string
  status?: number
}

// ==================== Contract ====================
export interface ContractVO {
  id: string
  contractNo: string
  contractName: string
  customerId: string
  customerName?: string
  projectId?: string
  contractAmount: number
  signDate?: string
  startDate?: string
  endDate?: string
  executionStatus: number
  remark?: string
  createTime: string
}

export interface ContractDTO {
  id?: string
  contractNo: string
  contractName: string
  customerId: string
  projectId?: string
  contractAmount?: number
  signDate?: string
  startDate?: string
  endDate?: string
  executionStatus?: number
  remark?: string
}

export interface ContractQuery {
  current: number
  size: number
  contractNo?: string
  contractName?: string
  customerId?: string
  projectId?: string
  executionStatus?: number
}

// ==================== Project ====================
export interface ProjectVO {
  id: string
  projectCode: string
  projectName: string
  customerId: string
  customerName?: string
  contractId?: string
  deptId?: string
  innerManagerId?: string
  innerManagerName?: string
  customerManagerName?: string
  customerManagerPhone?: string
  planStartDate?: string
  planEndDate?: string
  actualStartDate?: string
  actualEndDate?: string
  projectStatus: number
  remark?: string
  createTime: string
}

export interface ProjectDTO {
  id?: string
  projectCode: string
  projectName: string
  customerId: string
  contractId?: string
  deptId?: string
  innerManagerId?: string
  customerManagerName?: string
  customerManagerPhone?: string
  planStartDate?: string
  planEndDate?: string
  actualStartDate?: string
  actualEndDate?: string
  projectStatus?: number
  remark?: string
}

export interface ProjectQuery {
  current: number
  size: number
  projectName?: string
  projectCode?: string
  customerId?: string
  projectStatus?: number
  innerManagerId?: string
}

// ==================== Budget ====================
export interface BudgetVO {
  id: string
  projectId: string
  budgetItemId: string
  budgetItemName?: string
  totalAmount: number
  usedAmount: number
  frozenAmount: number
  availableAmount?: number
  warnRate: number
  executionRate?: number
  status: number
  remark?: string
  createTime: string
}

export interface BudgetDTO {
  id?: string
  projectId: string
  budgetItemId: string
  totalAmount: number
  warnRate?: number
  remark?: string
}

export interface BudgetQuery {
  current: number
  size: number
  projectId?: string
  budgetItemId?: string
}

// ==================== BudgetItem ====================
export interface BudgetItemVO {
  id: string
  parentId: string
  itemCode?: string
  itemName: string
  sortOrder: number
  children?: BudgetItemVO[]
  status: number
  createTime: string
}

export interface BudgetItemDTO {
  id?: string
  parentId: string
  itemCode?: string
  itemName: string
  sortOrder?: number
}

// ==================== BudgetAdjust ====================
export interface BudgetAdjustVO {
  id: string
  budgetId: string
  adjustType: number
  adjustAmount: number
  beforeAmount: number
  afterAmount: number
  approvalStatus: number
  remark?: string
  createTime: string
}

export interface BudgetAdjustDTO {
  budgetId: string
  adjustType: number
  adjustAmount: number
  remark?: string
}

// ==================== WorkReport ====================
export interface WorkReportVO {
  id: string
  projectId: string
  projectName?: string
  memberId: string
  memberName?: string
  userId: string
  workDate: string
  workHours: number
  workType: string
  taskContent?: string
  approvalStatus: number
  auditUserId?: string
  auditUserName?: string
  auditTime?: string
  costAmount?: number
  remark?: string
  createTime: string
}

export interface WorkReportDTO {
  id?: string
  projectId: string
  memberId: string
  workDate: string
  workHours: number
  workType?: string
  taskContent?: string
  remark?: string
}

export interface WorkReportQuery {
  current: number
  size: number
  projectId?: string
  memberId?: string
  approvalStatus?: number
  workType?: string
  startDate?: string
  endDate?: string
}

// ==================== Expense ====================
export interface ExpenseVO {
  id: string
  expenseNo: string
  projectId: string
  projectName?: string
  userId: string
  userName?: string
  deptId?: string
  totalAmount: number
  approvalStatus: number
  payStatus: number
  applyTime?: string
  auditUserId?: string
  auditUserName?: string
  auditTime?: string
  payTime?: string
  remark?: string
  createTime: string
  items?: ExpenseItemVO[]
}

export interface ExpenseItemVO {
  id: string
  expenseId: string
  budgetItemId: string
  budgetItemName?: string
  amount: number
  invoiceUrl?: string
  remark?: string
}

export interface ExpenseDTO {
  id?: string
  projectId: string
  deptId?: string
  totalAmount: number
  remark?: string
  items: ExpenseItemDTO[]
}

export interface ExpenseItemDTO {
  budgetItemId: string
  amount: number
  invoiceUrl?: string
  remark?: string
}

export interface ExpenseQuery {
  current: number
  size: number
  projectId?: string
  expenseNo?: string
  approvalStatus?: number
  userId?: string
}

// ==================== Invoice ====================
export interface InvoiceVO {
  id: string
  invoiceNo: string
  projectId: string
  projectName?: string
  customerId: string
  customerName?: string
  contractId?: string
  invoiceType: string
  invoiceAmount: number
  taxRate: number
  taxAmount: number
  invoiceDate: string
  invoiceStatus: number
  remark?: string
  createTime: string
}

export interface InvoiceDTO {
  id?: string
  invoiceNo: string
  projectId: string
  customerId: string
  contractId?: string
  invoiceType?: string
  invoiceAmount: number
  taxRate?: number
  invoiceDate: string
  remark?: string
}

export interface InvoiceQuery {
  current: number
  size: number
  projectId?: string
  invoiceNo?: string
  invoiceStatus?: number
  customerId?: string
}

// ==================== ReceivePlan ====================
export interface ReceivePlanVO {
  id: string
  projectId: string
  projectName?: string
  customerId: string
  customerName?: string
  contractId?: string
  invoiceId?: string
  planAmount: number
  planDate: string
  receivedAmount: number
  pendingAmount?: number
  planStatus: number
  overdueDays?: number
  remark?: string
  createTime: string
}

export interface ReceivePlanDTO {
  id?: string
  projectId: string
  customerId: string
  contractId?: string
  invoiceId?: string
  planAmount: number
  planDate: string
  remark?: string
}

export interface ReceivePlanQuery {
  current: number
  size: number
  projectId?: string
  customerId?: string
  planStatus?: number
}

// ==================== Receive ====================
export interface ReceiveVO {
  id: string
  receiveNo: string
  planId: string
  projectId: string
  projectName?: string
  customerId: string
  customerName?: string
  receiveAmount: number
  receiveDate: string
  receiveType: string
  receiveStatus: number
  remark?: string
  createTime: string
}

export interface ReceiveDTO {
  id?: string
  receiveNo: string
  planId: string
  projectId: string
  customerId: string
  receiveAmount: number
  receiveDate: string
  receiveType?: string
  remark?: string
}

export interface ReceiveQuery {
  current: number
  size: number
  projectId?: string
  planId?: string
  receiveStatus?: number
}

// ==================== Dashboard ====================
export interface DashboardStatsVO {
  totalProjects: number
  runningProjects: number
  monthlyRevenue: number
  pendingReceiveAmount: number
  receiveRate: number
}

export interface ProjectStatusDistVO {
  statusName: string
  count: number
}

export interface BudgetExecTopVO {
  projectName: string
  totalBudget: number
  usedAmount: number
  executionRate: number
}

export interface ReceiveTrendVO {
  month: string
  receiveAmount: number
  invoiceAmount: number
}

export interface EfficiencyRankVO {
  memberName: string
  totalHours: number
  projectCount: number
  costAmount: number
}

export interface OverdueReceiveVO {
  projectName: string
  customerName: string
  planAmount: number
  planDate: string
  overdueDays: number
  managerName: string
}

// ==================== My Workspace ====================
export interface MyWorkspaceStatsVO {
  myProjectCount: number
  weeklyHours: number
  pendingApprovalCount: number
  overdueReceiveCount: number
}

export interface PendingApprovalVO {
  bizId: string
  bizType: string
  bizNo: string
  applicantName: string
  amount?: number
  summary?: string
  submitTime?: string
}

// ==================== ProjectMember ====================
export interface ProjectMemberVO {
  id: string
  projectId: string
  userType: string
  userId?: string
  userName?: string
  customerUserName?: string
  customerUserPhone?: string
  roleType: string
  duty?: string
  costPrice?: number
  joinDate?: string
  leaveDate?: string
  createTime: string
}

export interface ProjectMemberDTO {
  id?: string
  projectId: string
  userType: string
  userId?: string
  customerUserName?: string
  customerUserPhone?: string
  roleType: string
  duty?: string
  costPrice?: number
  joinDate?: string
  leaveDate?: string
  remark?: string
}

// ==================== DocFolder ====================
export interface DocFolderVO {
  id: string
  projectId: string
  parentId: string
  folderName: string
  color?: string
  sortOrder: number
  createTime: string
  children?: DocFolderVO[]
}

export interface DocFolderDTO {
  id?: string
  projectId: string
  parentId?: string
  folderName: string
  color?: string
  sortOrder?: number
}

// ==================== Document ====================
export interface DocumentVO {
  id: string
  projectId: string
  folderId?: string
  docName: string
  docType: string
  secretLevel?: string
  fileUrl: string
  fileSize?: number
  fileSuffix?: string
  docVersion?: number
  isCurrentVersion?: number
  uploadUserId?: string
  uploadUserName?: string
  downloadCount?: number
  remark?: string
  createTime: string
  updateTime?: string
}

export interface DocumentDTO {
  id?: string
  projectId: string
  folderId?: string
  docName: string
  docType: string
  secretLevel?: string
  fileUrl: string
  fileSize?: number
  fileSuffix?: string
  remark?: string
}

export interface DocumentQuery {
  current: number
  size: number
  projectId?: string
  folderId?: string
  docName?: string
  docType?: string
  secretLevel?: string
}
