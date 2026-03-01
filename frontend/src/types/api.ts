export interface ApiError {
  code: string
  message: string
  metadata?: Record<string, unknown>
}

export interface User {
  id: string
  email: string
  name: string
  createdAt: string
  updatedAt: string
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

export interface SystemStatus {
  initialized: boolean
}
