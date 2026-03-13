import api from '@/lib/api'
import type {
  TemplateVariable,
  CreateTemplateVariableRequest,
  UpdateTemplateVariableRequest,
  ParseVariablesResult,
} from '@/types/api'

export async function listTemplateVariables(
  templateId: string,
): Promise<TemplateVariable[]> {
  const { data } = await api.get<TemplateVariable[]>(
    `/api/v1/templates/${templateId}/variables`,
  )
  return data
}

export async function createTemplateVariable(
  templateId: string,
  req: CreateTemplateVariableRequest,
): Promise<TemplateVariable> {
  const { data } = await api.post<TemplateVariable>(
    `/api/v1/templates/${templateId}/variables`,
    req,
  )
  return data
}

export async function updateTemplateVariable(
  templateId: string,
  varId: string,
  req: UpdateTemplateVariableRequest,
): Promise<TemplateVariable> {
  const { data } = await api.put<TemplateVariable>(
    `/api/v1/templates/${templateId}/variables/${varId}`,
    req,
  )
  return data
}

export async function deleteTemplateVariable(
  templateId: string,
  varId: string,
): Promise<void> {
  await api.delete(`/api/v1/templates/${templateId}/variables/${varId}`)
}

export async function parseTemplateVariables(
  templateId: string,
): Promise<ParseVariablesResult> {
  const { data } = await api.post<ParseVariablesResult>(
    `/api/v1/templates/${templateId}/variables/parse`,
  )
  return data
}
