import { request } from './request'

// 统一响应
export interface FerryResp<T> {
  success: boolean
  message?: string
  data?: T
  total?: number
  current?: number
  size?: number
}

// ============ 类型定义 ============
export interface Process {
  id: string
  name: string
  icon?: string
  classify?: string
  description?: string
  structure?: any
  formSchema?: any
  notify?: any
  creatorId?: string
  creatorName?: string
  enabled?: boolean
  submitCnt?: number
  createdAt?: string
  updatedAt?: string
}

export interface Classify {
  id: string
  name: string
  parentId?: string
  sortOrder?: number
  creatorId?: string
  createdAt?: string
  updatedAt?: string
}

export interface WorkOrder {
  id: string
  title: string
  processId: string
  processName: string
  priority?: number
  status?: string
  creatorId?: string
  creatorName?: string
  formData?: any
  currentNodeIDs?: string
  urgeCount?: number
  createdAt?: string
  updatedAt?: string
  finishedAt?: string
}

export interface WorkOrderTask {
  id: string
  workOrderId: string
  nodeId: string
  nodeName: string
  assigneeId?: string
  assigneeName?: string
  status?: string
  remark?: string
  operatorId?: string
  operatorName?: string
  createdAt?: string
  handledAt?: string
}

export interface WorkOrderHistory {
  id: string
  workOrderId: string
  nodeId?: string
  nodeName?: string
  action?: string
  remark?: string
  operatorId?: string
  operatorName?: string
  createdAt?: string
}

// ============ 通用列表请求 ============
function buildQuery(params: Record<string, any>): string {
  const arr = Object.keys(params)
    .filter((k) => params[k] !== undefined && params[k] !== null && params[k] !== '')
    .map((k) => encodeURIComponent(k) + '=' + encodeURIComponent(params[k]))
  return arr.length ? '?' + arr.join('&') : ''
}

// ============ 统计 ============
export function ferryStatistics(): Promise<any> {
  return request('/ferry/statistics', { method: 'GET' })
}

// ============ 用户管理 ============
export function ferryListUsers(params: {
  page?: number
  size?: number
  keyword?: string
  status?: string
  role?: string
}): Promise<any> {
  return request('/ferry/users' + buildQuery(params), { method: 'GET' })
}

export function ferryGetUser(id: string): Promise<any> {
  return request('/ferry/users/' + encodeURIComponent(id), { method: 'GET' })
}

export function ferryCreateUser(data: {
  username: string
  password: string
  displayName?: string
  email?: string
  phone?: string
  role?: string
  status?: string
}): Promise<any> {
  return request('/ferry/users', { method: 'POST', data })
}

export function ferryUpdateUser(id: string, data: {
  displayName?: string
  email?: string
  phone?: string
  role?: string
  status?: string
}): Promise<any> {
  return request('/ferry/users/' + encodeURIComponent(id), { method: 'PUT', data })
}

export function ferryDeleteUser(id: string): Promise<any> {
  return request('/ferry/users/' + encodeURIComponent(id), { method: 'DELETE' })
}

export function ferryResetUserPassword(id: string, password: string): Promise<any> {
  return request('/ferry/users/' + encodeURIComponent(id) + '/password', {
    method: 'POST',
    data: { password }
  })
}

// ============ 角色管理 ============
export function ferryListRoles(params: {
  page?: number
  size?: number
  keyword?: string
}): Promise<any> {
  return request('/ferry/roles' + buildQuery(params), { method: 'GET' })
}

export function ferryListAllRoles(): Promise<any> {
  return request('/ferry/roles/all', { method: 'GET' })
}

export function ferryCreateRole(data: {
  name: string
  description?: string
  status?: string
  codes?: string
  permissions?: string[]
}): Promise<any> {
  return request('/ferry/roles', { method: 'POST', data })
}

export function ferryUpdateRole(id: string, data: {
  name?: string
  description?: string
  status?: string
  codes?: string
  permissions?: string[]
}): Promise<any> {
  return request('/ferry/roles/' + encodeURIComponent(id), { method: 'PUT', data })
}

export function ferryDeleteRole(id: string): Promise<any> {
  return request('/ferry/roles/' + encodeURIComponent(id), { method: 'DELETE' })
}

// ============ 岗位管理 ============
export function ferryListPosts(keyword?: string): Promise<any> {
  const q = keyword ? '?keyword=' + encodeURIComponent(keyword) : ''
  return request('/ferry/posts' + q, { method: 'GET' })
}

export function ferryCreatePost(data: {
  name: string
  description?: string
  sortOrder?: number
  status?: string
}): Promise<any> {
  return request('/ferry/posts', { method: 'POST', data })
}

export function ferryUpdatePost(id: string, data: {
  name?: string
  description?: string
  sortOrder?: number
  status?: string
}): Promise<any> {
  return request('/ferry/posts/' + encodeURIComponent(id), { method: 'PUT', data })
}

export function ferryDeletePost(id: string): Promise<any> {
  return request('/ferry/posts/' + encodeURIComponent(id), { method: 'DELETE' })
}

// ============ 系统配置 ============
export function ferryListSystemSettings(keyword?: string): Promise<any> {
  const q = keyword ? '?keyword=' + encodeURIComponent(keyword) : ''
  return request('/ferry/system-settings' + q, { method: 'GET' })
}

export function ferrySaveSystemSettings(items: { settingKey: string; value: string; remark?: string }[]): Promise<any> {
  return request('/ferry/system-settings', { method: 'POST', data: items })
}

// ============ 部门管理 ============
export function ferryListDepartments(params?: { keyword?: string; status?: string }): Promise<any> {
  const p = params || {}
  return request('/ferry/departments' + buildQuery(p), { method: 'GET' })
}
export function ferryListDepartmentsTree(): Promise<any> {
  return request('/ferry/departments/tree', { method: 'GET' })
}
export function ferryCreateDepartment(data: {
  name: string
  parentId?: string
  leader?: string
  phone?: string
  email?: string
  sortOrder?: number
  status?: string
  remark?: string
}): Promise<any> {
  return request('/ferry/departments', { method: 'POST', data })
}
export function ferryUpdateDepartment(id: string, data: any): Promise<any> {
  return request('/ferry/departments/' + encodeURIComponent(id), { method: 'PUT', data })
}
export function ferryDeleteDepartment(id: string): Promise<any> {
  return request('/ferry/departments/' + encodeURIComponent(id), { method: 'DELETE' })
}

// ============ 菜单管理 ============
export function ferryListMenus(params?: { keyword?: string; status?: string; type?: string }): Promise<any> {
  return request('/ferry/menus' + buildQuery(params || {}), { method: 'GET' })
}
export function ferryCreateMenu(data: any): Promise<any> {
  return request('/ferry/menus', { method: 'POST', data })
}
export function ferryUpdateMenu(id: string, data: any): Promise<any> {
  return request('/ferry/menus/' + encodeURIComponent(id), { method: 'PUT', data })
}
export function ferryDeleteMenu(id: string): Promise<any> {
  return request('/ferry/menus/' + encodeURIComponent(id), { method: 'DELETE' })
}

// ============ 字典管理 ============
export function ferryListDictionaries(params?: { keyword?: string; status?: string; page?: number; size?: number }): Promise<any> {
  return request('/ferry/dictionaries' + buildQuery(params || {}), { method: 'GET' })
}
export function ferryCreateDictionary(data: { name: string; type: string; status?: string; remark?: string }): Promise<any> {
  return request('/ferry/dictionaries', { method: 'POST', data })
}
export function ferryUpdateDictionary(id: string, data: any): Promise<any> {
  return request('/ferry/dictionaries/' + encodeURIComponent(id), { method: 'PUT', data })
}
export function ferryDeleteDictionary(id: string): Promise<any> {
  return request('/ferry/dictionaries/' + encodeURIComponent(id), { method: 'DELETE' })
}
export function ferryListDictItems(params?: { dictType?: string; keyword?: string; status?: string }): Promise<any> {
  return request('/ferry/dict-items' + buildQuery(params || {}), { method: 'GET' })
}
export function ferryCreateDictItem(data: any): Promise<any> {
  return request('/ferry/dict-items', { method: 'POST', data })
}
export function ferryUpdateDictItem(id: string, data: any): Promise<any> {
  return request('/ferry/dict-items/' + encodeURIComponent(id), { method: 'PUT', data })
}
export function ferryDeleteDictItem(id: string): Promise<any> {
  return request('/ferry/dict-items/' + encodeURIComponent(id), { method: 'DELETE' })
}

// ============ 参数配置 ============
export function ferryListParameters(params?: { keyword?: string; status?: string; type?: string; page?: number; size?: number }): Promise<any> {
  return request('/ferry/parameters' + buildQuery(params || {}), { method: 'GET' })
}
export function ferryCreateParameter(data: any): Promise<any> {
  return request('/ferry/parameters', { method: 'POST', data })
}
export function ferryUpdateParameter(id: string, data: any): Promise<any> {
  return request('/ferry/parameters/' + encodeURIComponent(id), { method: 'PUT', data })
}
export function ferryDeleteParameter(id: string): Promise<any> {
  return request('/ferry/parameters/' + encodeURIComponent(id), { method: 'DELETE' })
}

// ============ 流程定义 ============
export function ferryListProcesses(params: {
  page?: number
  size?: number
  keyword?: string
  classify?: string
}): Promise<FerryResp<Process[]>> {
  return request('/ferry/processes' + buildQuery(params), { method: 'GET' })
}

export function ferryListEnabledProcesses(): Promise<FerryResp<Process[]>> {
  return request('/ferry/processes/enabled', { method: 'GET' })
}

export function ferryGetProcess(id: string): Promise<FerryResp<Process>> {
  return request('/ferry/processes/' + encodeURIComponent(id), { method: 'GET' })
}

export function ferryCreateProcess(data: Partial<Process>): Promise<FerryResp<Process>> {
  return request('/ferry/processes', { method: 'POST', data })
}

export function ferryUpdateProcess(id: string, data: Partial<Process>): Promise<FerryResp<any>> {
  return request('/ferry/processes/' + encodeURIComponent(id), { method: 'PUT', data })
}

export function ferryDeleteProcess(id: string): Promise<FerryResp<any>> {
  return request('/ferry/processes/' + encodeURIComponent(id), { method: 'DELETE' })
}

export function ferryCloneProcess(id: string): Promise<FerryResp<Process>> {
  return request('/ferry/processes/' + encodeURIComponent(id) + '/clone', { method: 'POST' })
}

export function ferryToggleProcessEnabled(id: string, enabled: boolean): Promise<FerryResp<any>> {
  return request('/ferry/processes/' + encodeURIComponent(id) + '/enabled', { method: 'POST', data: { enabled } })
}

// ============ 流程分类 ============
export function ferryListClassifies(): Promise<FerryResp<Classify[]>> {
  return request('/ferry/classifies', { method: 'GET' })
}

export function ferryCreateClassify(data: Partial<Classify>): Promise<FerryResp<Classify>> {
  return request('/ferry/classifies', { method: 'POST', data })
}

export function ferryUpdateClassify(id: string, data: Partial<Classify>): Promise<FerryResp<any>> {
  return request('/ferry/classifies/' + encodeURIComponent(id), { method: 'PUT', data })
}

export function ferryDeleteClassify(id: string): Promise<FerryResp<any>> {
  return request('/ferry/classifies/' + encodeURIComponent(id), { method: 'DELETE' })
}

// ============ 表单 schema ============
export function ferryGetApplyForm(id: string): Promise<FerryResp<{ process: Process; formSchema: any; structure: any }>> {
  return request('/ferry/processes/' + encodeURIComponent(id) + '/apply-form', { method: 'GET' })
}

// ============ 工单 ============
export function ferrySubmitWorkOrder(data: {
  processId: string
  title: string
  priority?: number
  formData?: any
}): Promise<FerryResp<WorkOrder>> {
  return request('/ferry/work-orders', { method: 'POST', data })
}

export function ferryGetWorkOrderDetail(id: string): Promise<FerryResp<{
  workOrder: WorkOrder
  tasks: WorkOrderTask[]
  histories: WorkOrderHistory[]
}>> {
  return request('/ferry/work-orders/' + encodeURIComponent(id), { method: 'GET' })
}

export function ferryHandleTask(workOrderId: string, taskId: string, action: 'approve' | 'reject', remark?: string): Promise<FerryResp<any>> {
  return request('/ferry/work-orders/' + encodeURIComponent(workOrderId) + '/tasks/' + encodeURIComponent(taskId) + '/handle', {
    method: 'POST',
    data: { taskId, action, remark }
  })
}

export function ferryRevokeWorkOrder(id: string, remark?: string): Promise<FerryResp<any>> {
  return request('/ferry/work-orders/' + encodeURIComponent(id) + '/revoke', { method: 'POST', data: { remark } })
}

export function ferryUrgeWorkOrder(id: string, remark?: string): Promise<FerryResp<any>> {
  return request('/ferry/work-orders/' + encodeURIComponent(id) + '/urge', { method: 'POST', data: { remark } })
}

export function ferryMyTodo(params: { page?: number; size?: number; keyword?: string; status?: string }): Promise<FerryResp<WorkOrderTask[]>> {
  return request('/ferry/work-orders/my/todo' + buildQuery(params), { method: 'GET' })
}

export function ferryMyApply(params: { page?: number; size?: number; keyword?: string; status?: string }): Promise<FerryResp<WorkOrder[]>> {
  return request('/ferry/work-orders/my/apply' + buildQuery(params), { method: 'GET' })
}

export function ferryMyRelated(params: { page?: number; size?: number; keyword?: string }): Promise<FerryResp<WorkOrder[]>> {
  return request('/ferry/work-orders/my/related' + buildQuery(params), { method: 'GET' })
}

export function ferryAllWorkOrders(params: {
  page?: number
  size?: number
  keyword?: string
  status?: string
  processId?: string
}): Promise<FerryResp<WorkOrder[]>> {
  return request('/ferry/work-orders' + buildQuery(params), { method: 'GET' })
}

// ============ 状态映射 ============
export const STATUS_MAP: Record<string, string> = {
  running: '流转中',
  approved: '已通过',
  rejected: '已驳回',
  canceled: '已撤销'
}

export const TASK_STATUS_MAP: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  approved: '已通过',
  rejected: '已驳回',
  skipped: '已跳过'
}

export const PRIORITY_MAP: Record<number, string> = {
  1: '普通',
  2: '紧急',
  3: '非常紧急'
}
