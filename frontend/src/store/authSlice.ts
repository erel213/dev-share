import { createSlice } from '@reduxjs/toolkit'
import type { PayloadAction } from '@reduxjs/toolkit'
import type { User } from '@/types/api'
import type { RootState } from '@/store'

export type AuthStatus = 'idle' | 'loading' | 'authenticated' | 'unauthenticated'

interface AuthState {
  user: User | null
  status: AuthStatus
}

const initialState: AuthState = {
  user: null,
  status: 'idle',
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setUser(state, action: PayloadAction<User>) {
      state.user = action.payload
      state.status = 'authenticated'
    },
    clearUser(state) {
      state.user = null
      state.status = 'unauthenticated'
    },
    setAuthStatus(state, action: PayloadAction<AuthStatus>) {
      state.status = action.payload
    },
  },
})

export const { setUser, clearUser, setAuthStatus } = authSlice.actions

export const selectUser = (state: RootState) => state.auth.user
export const selectIsAuthenticated = (state: RootState) =>
  state.auth.status === 'authenticated'
export const selectAuthStatus = (state: RootState) => state.auth.status

export default authSlice.reducer
