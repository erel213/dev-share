import { useEffect, useState } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import api from '@/lib/api'
import type { SystemStatus } from '@/types/api'
import { Skeleton } from '@/components/ui/skeleton'

export default function ProtectedRoute() {
  const { status, checkAuth } = useAuth()
  const [systemCheck, setSystemCheck] = useState<
    'loading' | 'not_initialized' | 'ready'
  >('loading')

  useEffect(() => {
    api
      .get<SystemStatus>('/admin/status')
      .then((res) => {
        if (!res.data.initialized) {
          setSystemCheck('not_initialized')
        } else {
          setSystemCheck('ready')
        }
      })
      .catch(() => {
        // If status check fails, assume initialized and let auth handle it
        setSystemCheck('ready')
      })
  }, [])

  useEffect(() => {
    if (systemCheck === 'ready' && status === 'idle') {
      checkAuth()
    }
  }, [systemCheck, status, checkAuth])

  if (systemCheck === 'not_initialized') {
    return <Navigate to="/setup" replace />
  }

  if (systemCheck === 'loading' || status === 'idle' || status === 'loading') {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="w-64 space-y-4">
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
        </div>
      </div>
    )
  }

  if (status === 'unauthenticated') {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}
