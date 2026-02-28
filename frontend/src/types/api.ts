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
