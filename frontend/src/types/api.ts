export interface ApiError {
  code: string
  message: string
  metadata?: Record<string, unknown>
}

export interface User {
  id: string
  name: string
  role: string
  workspaceId: string
}

export interface Workspace {
  id: string
  name: string
  slug: string
  ownerId: string
  createdAt: string
  updatedAt: string
}

export interface Environment {
  id: string
  name: string
  description: string
  created_by: string
  created_by_name: string
  workspace_id: string
  template_id: string
  template_name: string
  status:
    | 'pending'
    | 'initialized'
    | 'planning'
    | 'applying'
    | 'ready'
    | 'destroying'
    | 'destroyed'
    | 'error'
  last_applied_at?: string
  last_operation?: string
  last_error?: string
  ttl_seconds?: number
  created_at: string
  updated_at: string
}

export interface CreateEnvironmentRequest {
  name: string
  description?: string
  template_id: string
  ttl_seconds?: number
}

export interface ListEnvironmentsParams {
  scope?: 'user' | 'all'
  status?: string
  template_id?: string
  created_by?: string
  search?: string
  sort_by?: string
  order?: 'ASC' | 'DESC'
  limit?: number
  offset?: number
}

export interface AdminInitRequest {
  admin_name: string
  admin_email: string
  admin_password: string
  workspace_name: string
  workspace_description?: string
}

export interface AdminInitResponse {
  message: string
  workspace_id: string
  admin_user_id: string
}

export interface LoginResponse {
  user_id: string
  name: string
  role: string
  workspace_id: string
}

export interface Template {
  id: string
  name: string
  workspace_id: string
  path: string
  created_at: string
  updated_at: string
}

export interface TemplateFileInfo {
  name: string
  size: number
}

export interface SystemStatus {
  initialized: boolean
}

export interface TemplateVariable {
  id: string
  template_id: string
  key: string
  description: string
  var_type: string
  default_value: string
  is_sensitive: boolean
  is_required: boolean
  validation_regex: string
  is_auto_parsed: boolean
  display_order: number
  created_at: string
  updated_at: string
}

export interface CreateTemplateVariableRequest {
  key: string
  description?: string
  var_type?: string
  default_value?: string
  is_sensitive?: boolean
  is_required?: boolean
  validation_regex?: string
}

export interface UpdateTemplateVariableRequest {
  description?: string
  var_type?: string
  default_value?: string
  is_sensitive?: boolean
  is_required?: boolean
  validation_regex?: string
  display_order?: number
}

export interface ParseVariablesResult {
  variables: TemplateVariable[]
  added: number
  updated: number
  removed: number
}

export interface EnvironmentVariableValue {
  template_variable_id: string
  key: string
  value: string
  is_sensitive: boolean
}

export interface SetVariableValueEntry {
  template_variable_id: string
  value: string
}

export interface Group {
  id: string
  name: string
  description: string
  workspace_id: string
  access_all_templates: boolean
  created_at: string
  updated_at: string
}

export interface CreateGroupRequest {
  name: string
  description?: string
  access_all_templates?: boolean
}

export interface UpdateGroupRequest {
  name?: string
  description?: string
  access_all_templates?: boolean
}

export interface AdminUser {
  id: string
  name: string
  email: string
  role: string
  workspace_id: string
  created_at: string
  updated_at: string
}

export interface InviteUserRequest {
  name: string
  email: string
  role: string
}

export interface InviteUserResponse {
  user_id: string
  name: string
  email: string
  role: string
  password: string
}

export interface ResetPasswordResponse {
  user_id: string
  password: string
}
