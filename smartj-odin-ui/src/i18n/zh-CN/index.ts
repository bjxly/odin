import common from './framework/common'
import components from './framework/components'
import menu from './framework/menu'
import layout from './framework/layout'
import auth from './pages/auth'
import dashboard from './pages/dashboard'
import gateway from './pages/gateway'
import sysUser from './pages/sys/user'
import sysRole from './pages/sys/role'
import sysMenu from './pages/sys/menu'
import sysDept from './pages/sys/dept'
import sysDict from './pages/sys/dict'
import sysPost from './pages/sys/post'
import sysConfig from './pages/sys/config'
import sysCloudRoute from './pages/sys/cloud-route'
import sysProfile from './pages/sys/profile'
import sysNotice from './pages/sys/notice'
import sysSms from './pages/sys/sms'
import logAudit from './pages/log/audit'
import sysOperLog from './pages/log/operation'
import tenant from './pages/tenant'
import platform from './pages/platform'
import error from './pages/error'
import sso from './pages/sso'
import pay from './pages/pay'
import odinDatasource from './pages/odin/datasource'
import odinOntology from './pages/odin/ontology'
import odinMapping from './pages/odin/mapping'
import odinQuery from './pages/odin/query'
import odinAssistant from './pages/odin/assistant'
import odinInsight from './pages/odin/insight'
import odinConnector from './pages/odin/connector'
import odinReason from './pages/odin/reason'
import odinWorkspace from './pages/odin/workspace'

export default {
  ...common,
  ...components,
  ...menu,
  ...layout,
  ...auth,
  ...gateway,
  ...sysUser,
  ...sysRole,
  ...sysMenu,
  ...sysDept,
  ...sysDict,
  ...sysPost,
  ...sysConfig,
  ...sysCloudRoute,
  ...sysProfile,
  ...sysNotice,
  ...sysSms,
  ...logAudit,
  ...sysOperLog,
  ...tenant,
  ...platform,
  ...error,
  ...sso,
  ...pay,
  ...odinDatasource,
  ...odinOntology,
  ...odinMapping,
  ...odinQuery,
  ...odinAssistant,
  ...odinInsight,
  ...odinConnector,
  ...odinReason,
  ...odinWorkspace,
  // 与gateway i18n key命名存在相同，覆盖gateway
  ...dashboard,
}
