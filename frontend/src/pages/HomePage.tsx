import { useAuth } from '@/hooks/useAuth'
import AdminEnvironmentsCard from '@/components/dashboard/AdminEnvironmentsCard'

export default function HomePage() {
  const { user } = useAuth()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground mt-2">
          Manage temporary developer environments with ease.
        </p>
      </div>
      {user?.role === 'admin' && <AdminEnvironmentsCard />}
    </div>
  )
}
