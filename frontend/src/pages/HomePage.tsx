import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '@/lib/api'
import type { SystemStatus } from '@/types/api'

export default function HomePage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api
      .get<SystemStatus>('/admin/status')
      .then((res) => {
        if (!res.data.initialized) {
          navigate('/setup', { replace: true })
        } else {
          setLoading(false)
        }
      })
      .catch(() => setLoading(false))
  }, [navigate])

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </main>
    )
  }

  return (
    <main className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold tracking-tight">Dev Share</h1>
        <p className="mt-4 text-muted-foreground">
          Manage temporary developer environments with ease.
        </p>
      </div>
    </main>
  )
}
