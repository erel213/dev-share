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
  workspaceId: string
  status: 'pending' | 'running' | 'stopped' | 'failed'
  createdAt: string
  updatedAt: string
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
