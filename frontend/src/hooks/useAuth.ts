import { useCallback } from 'react'
import api from '@/lib/api'
import type { LoginResponse } from '@/types/api'
import { useAppDispatch, useAppSelector } from '@/store'
import {
  setUser,
  clearUser,
  setAuthStatus,
  selectUser,
  selectAuthStatus,
  selectIsAuthenticated,
} from '@/store/authSlice'

export function useAuth() {
  const dispatch = useAppDispatch()
  const user = useAppSelector(selectUser)
  const status = useAppSelector(selectAuthStatus)
  const isAuthenticated = useAppSelector(selectIsAuthenticated)

  const mapLoginResponse = (data: LoginResponse) => ({
    id: data.user_id,
    name: data.name,
    isAdmin: data.is_admin,
    workspaceId: data.workspace_id,
  })

  const checkAuth = useCallback(async () => {
    dispatch(setAuthStatus('loading'))
    try {
      const response = await api.get<LoginResponse>('/api/v1/me')
      dispatch(setUser(mapLoginResponse(response.data)))
    } catch {
      dispatch(clearUser())
    }
  }, [dispatch])

  const login = useCallback(
    async (credentials: { email: string; password: string }) => {
      dispatch(setAuthStatus('loading'))
      const response = await api.post<LoginResponse>(
        '/api/v1/login',
        credentials,
      )
      dispatch(setUser(mapLoginResponse(response.data)))
    },
    [dispatch],
  )

  const logout = useCallback(() => {
    dispatch(clearUser())
  }, [dispatch])

  return { user, status, isAuthenticated, checkAuth, login, logout }
}
