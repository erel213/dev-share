import api from '@/lib/api'
import type { Template, TemplateFileInfo } from '@/types/api'
import type { FileWithPath } from '@/components/templates/FileDropzone'

export async function createTemplate(
  name: string,
  workspaceId: string,
  files: FileWithPath[],
): Promise<Template> {
  const formData = new FormData()
  formData.append('name', name)
  formData.append('workspace_id', workspaceId)
  console.log('[createTemplate] files to upload:', files.map(({ file, path }) => ({ path, fileName: file.name, size: file.size })))
  for (const { file, path } of files) {
    formData.append('files', file, path)
    formData.append('paths', path)
  }
  // Log what FormData actually contains
  for (const [key, value] of formData.entries()) {
    if (value instanceof File) {
      console.log(`[createTemplate] FormData entry: key="${key}", filename="${value.name}", size=${value.size}`)
    }
  }
  const { data } = await api.post<Template>('/api/v1/templates', formData, {
    headers: { 'Content-Type': undefined },
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
  files?: FileWithPath[],
): Promise<Template> {
  const formData = new FormData()
  formData.append('name', name)
  if (files) {
    for (const { file, path } of files) {
      formData.append('files', file, path)
      formData.append('paths', path)
    }
  }
  const { data } = await api.put<Template>(`/api/v1/templates/${id}`, formData, {
    headers: { 'Content-Type': undefined },
  })
  return data
}

export async function deleteTemplate(id: string): Promise<void> {
  await api.delete(`/api/v1/templates/${id}`)
}

export async function getTemplateFiles(
  templateId: string,
): Promise<TemplateFileInfo[]> {
  const { data } = await api.get<TemplateFileInfo[]>(
    `/api/v1/templates/${templateId}/files`,
  )
  return data
}

export async function getTemplateFileContent(
  templateId: string,
  filePath: string,
): Promise<string> {
  const { data } = await api.get<string>(
    `/api/v1/templates/${templateId}/files/content`,
    { params: { path: filePath }, responseType: 'text' },
  )
  return data
}
