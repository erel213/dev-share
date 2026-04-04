import api from '@/lib/api'
import type {
  Group,
  CreateGroupRequest,
  UpdateGroupRequest,
  AdminUser,
  Template,
} from '@/types/api'

export async function listGroups(): Promise<Group[]> {
  const { data } = await api.get<Group[]>('/api/v1/groups')
  return data
}

export async function getGroup(id: string): Promise<Group> {
  const { data } = await api.get<Group>(`/api/v1/groups/${id}`)
  return data
}

export async function createGroup(
  request: CreateGroupRequest,
): Promise<Group> {
  const { data } = await api.post<Group>('/api/v1/groups', request)
  return data
}

export async function updateGroup(
  id: string,
  request: UpdateGroupRequest,
): Promise<Group> {
  const { data } = await api.put<Group>(`/api/v1/groups/${id}`, request)
  return data
}

export async function deleteGroup(id: string): Promise<void> {
  await api.delete(`/api/v1/groups/${id}`)
}

export async function getGroupMembers(id: string): Promise<string[]> {
  const { data } = await api.get<string[]>(`/api/v1/groups/${id}/members`)
  return data
}

export async function addGroupMembers(
  id: string,
  userIds: string[],
): Promise<void> {
  await api.post(`/api/v1/groups/${id}/members`, { user_ids: userIds })
}

export async function removeGroupMember(
  groupId: string,
  userId: string,
): Promise<void> {
  await api.delete(`/api/v1/groups/${groupId}/members/${userId}`)
}

export async function getGroupTemplateAccess(
  id: string,
): Promise<string[]> {
  const { data } = await api.get<string[]>(`/api/v1/groups/${id}/templates`)
  return data
}

export async function addGroupTemplateAccess(
  id: string,
  templateIds: string[],
): Promise<void> {
  await api.post(`/api/v1/groups/${id}/templates`, {
    template_ids: templateIds,
  })
}

export async function removeGroupTemplateAccess(
  groupId: string,
  templateId: string,
): Promise<void> {
  await api.delete(`/api/v1/groups/${groupId}/templates/${templateId}`)
}

// Helper: fetch users list (re-export from users-api for convenience)
export async function listAllUsers(): Promise<AdminUser[]> {
  const { data } = await api.get<AdminUser[]>('/api/v1/admin/users')
  return data
}

export async function listAllTemplates(): Promise<Template[]> {
  const { data } = await api.get<Template[]>('/api/v1/templates')
  return data
}
