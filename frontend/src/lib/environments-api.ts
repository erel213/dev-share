import api from '@/lib/api'
import type {
  Environment,
  EnvironmentOutputs,
  CreateEnvironmentRequest,
  ListEnvironmentsParams,
} from '@/types/api'

export async function listEnvironments(
  params?: ListEnvironmentsParams,
): Promise<Environment[]> {
  const { data } = await api.get<Environment[]>('/api/v1/environments', {
    params,
  })
  return data
}

export async function createEnvironment(
  request: CreateEnvironmentRequest,
): Promise<Environment> {
  const { data } = await api.post<Environment>(
    '/api/v1/environments',
    request,
  )
  return data
}

export async function planEnvironment(id: string): Promise<Environment> {
  const { data } = await api.post<Environment>(
    `/api/v1/environments/${id}/plan`,
  )
  return data
}

export async function applyEnvironment(id: string): Promise<Environment> {
  const { data } = await api.post<Environment>(
    `/api/v1/environments/${id}/apply`,
  )
  return data
}

export async function destroyEnvironment(id: string): Promise<Environment> {
  const { data } = await api.post<Environment>(
    `/api/v1/environments/${id}/destroy`,
  )
  return data
}

export async function deleteEnvironment(id: string): Promise<void> {
  await api.delete(`/api/v1/environments/${id}`)
}

export async function getEnvironmentOutputs(
  id: string,
): Promise<EnvironmentOutputs> {
  const { data } = await api.get<EnvironmentOutputs>(
    `/api/v1/environments/${id}/outputs`,
  )
  return data
}
