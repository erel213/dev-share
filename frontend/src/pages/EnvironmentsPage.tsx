import { Fragment, useCallback, useEffect, useState } from 'react'
import { ChevronRight, Plus } from 'lucide-react'
import { listEnvironments } from '@/lib/environments-api'
import { useAuth } from '@/hooks/useAuth'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import EnvironmentStatusBadge from '@/components/environments/EnvironmentStatusBadge'
import EnvironmentActions from '@/components/environments/EnvironmentActions'
import CreateEnvironmentDialog from '@/components/environments/CreateEnvironmentDialog'
import EnvironmentOutputs from '@/components/environments/EnvironmentOutputs'
import { cn } from '@/lib/utils'
import type { ApiError, Environment } from '@/types/api'

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(iso))
}

export default function EnvironmentsPage() {
  const { user } = useAuth()
  const [environments, setEnvironments] = useState<Environment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const fetchEnvironments = useCallback(async () => {
    setError('')
    setLoading(true)
    try {
      const data = await listEnvironments({ scope: 'user' })
      data ? setEnvironments(data) : setEnvironments([])
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load environments')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchEnvironments()
  }, [fetchEnvironments])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Environments</h1>
          <p className="text-muted-foreground mt-2">
            Manage your development environments.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Environment
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Template</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Owner</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="w-[60px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell>
                    <Skeleton className="h-4 w-28" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-24" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-16" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-20" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-28" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-8" />
                  </TableCell>
                </TableRow>
              ))
            ) : environments.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="text-muted-foreground h-24 text-center"
                >
                  No environments found.
                </TableCell>
              </TableRow>
            ) : (
              environments.map((env) => {
                const hasOutputs = !!env.last_applied_at
                const isExpanded = expandedId === env.id
                return (
                  <Fragment key={env.id}>
                    <TableRow
                      className={cn(
                        hasOutputs && 'cursor-pointer hover:bg-muted/50',
                      )}
                      onClick={() =>
                        hasOutputs &&
                        setExpandedId(isExpanded ? null : env.id)
                      }
                    >
                      <TableCell className="font-medium">
                        <span className="flex items-center gap-1">
                          {hasOutputs && (
                            <ChevronRight
                              className={cn(
                                'h-4 w-4 shrink-0 transition-transform',
                                isExpanded && 'rotate-90',
                              )}
                            />
                          )}
                          {env.name}
                        </span>
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {env.template_name || '—'}
                      </TableCell>
                      <TableCell>
                        <EnvironmentStatusBadge status={env.status} />
                      </TableCell>
                      <TableCell>
                        {env.created_by === user?.id
                          ? 'You'
                          : env.created_by_name}
                      </TableCell>
                      <TableCell>{formatDate(env.created_at)}</TableCell>
                      <TableCell>
                        <EnvironmentActions
                          environment={env}
                          currentUserId={user?.id ?? ''}
                          isAdmin={user?.role === 'admin'}
                          onRefresh={fetchEnvironments}
                        />
                      </TableCell>
                    </TableRow>
                    {isExpanded && (
                      <TableRow>
                        <TableCell colSpan={6} className="bg-muted/30 p-4">
                          <EnvironmentOutputs environmentId={env.id} />
                        </TableCell>
                      </TableRow>
                    )}
                  </Fragment>
                )
              })
            )}
          </TableBody>
        </Table>
      </div>

      <CreateEnvironmentDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSuccess={fetchEnvironments}
      />
    </div>
  )
}
