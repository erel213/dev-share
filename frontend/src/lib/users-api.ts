import api from '@/lib/api'
import type {
  AdminUser,
  InviteUserRequest,
  InviteUserResponse,
  ResetPasswordResponse,
} from '@/types/api'

export async function listUsers(): Promise<AdminUser[]> {
  const { data } = await api.get<AdminUser[]>('/api/v1/admin/users')
  return data
}

export async function inviteUser(
  request: InviteUserRequest,
): Promise<InviteUserResponse> {
  const { data } = await api.post<InviteUserResponse>(
    '/api/v1/admin/users/invite',
    request,
  )
  return data
}

export async function resetUserPassword(
  userId: string,
): Promise<ResetPasswordResponse> {
  const { data } = await api.post<ResetPasswordResponse>(
    `/api/v1/admin/users/${userId}/reset-password`,
  )
  return data
}

export async function deleteUser(userId: string): Promise<void> {
  await api.delete(`/api/v1/admin/users/${userId}`)
}
