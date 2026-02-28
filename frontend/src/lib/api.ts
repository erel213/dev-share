import axios from 'axios'
import type { ApiError } from '@/types/api'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const apiError: ApiError = error.response?.data ?? {
      code: 'UNKNOWN_ERROR',
      message: error.message ?? 'An unexpected error occurred',
    }
    return Promise.reject(apiError)
  },
)

export default api
