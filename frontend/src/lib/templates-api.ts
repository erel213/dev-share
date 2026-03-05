import api from '@/lib/api'
import type { Template } from '@/types/api'

export async function createTemplate(
  name: string,
  workspaceId: string,
  files: File[],
): Promise<Template> {
  const formData = new FormData()
  formData.append('name', name)
  formData.append('workspace_id', workspaceId)
  for (const file of files) {
    formData.append('files', file)
  }
  const { data } = await api.post<Template>('/api/v1/templates', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

export async function listTemplates(params?: {
  limit?: number
  offset?: number
  sort_by?: string
  order?: string
}): Promise<Template[]> {
  const { data } = await api.get<Template[]>('/api/v1/templates', { params })
  return data
}

export async function getTemplate(id: string): Promise<Template> {
  const { data } = await api.get<Template>(`/api/v1/templates/${id}`)
  return data
}

export async function getWorkspaceTemplates(
  workspaceId: string,
): Promise<Template[]> {
  const { data } = await api.get<Template[]>(
    `/api/v1/templates/workspace/${workspaceId}`,
  )
  return data
}

export async function updateTemplate(
  id: string,
  name: string,
  files?: File[],
): Promise<Template> {
  const formData = new FormData()
  formData.append('name', name)
  if (files) {
    for (const file of files) {
      formData.append('files', file)
    }
  }
  const { data } = await api.put<Template>(`/api/v1/templates/${id}`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

export async function deleteTemplate(id: string): Promise<void> {
  await api.delete(`/api/v1/templates/${id}`)
}
