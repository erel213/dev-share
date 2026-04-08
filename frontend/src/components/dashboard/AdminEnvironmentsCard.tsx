import { useCallback, useEffect, useMemo, useState } from 'react'
import { Search } from 'lucide-react'
import { listEnvironments } from '@/lib/environments-api'
import { listTemplates } from '@/lib/templates-api'
import { useAuth } from '@/hooks/useAuth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
import type { ApiError, Environment, Template } from '@/types/api'

const ALL_STATUSES = [
  'pending',
  'initialized',
  'planning',
  'applying',
  'ready',
  'destroying',
  'destroyed',
  'error',
] as const

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(iso))
}

export default function AdminEnvironmentsCard() {
  const { user } = useAuth()
  const [environments, setEnvironments] = useState<Environment[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [templateFilter, setTemplateFilter] = useState('')

  const fetchData = useCallback(async () => {
    setError('')
    setLoading(true)
    try {
      const [envs, tmpls] = await Promise.all([
        listEnvironments({
          scope: 'all',
          search: search || undefined,
          status: statusFilter || undefined,
          template_id: templateFilter || undefined,
        }),
        listTemplates(),
      ])
      setEnvironments(envs ?? [])
      setTemplates(tmpls ?? [])
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load environments')
    } finally {
      setLoading(false)
    }
  }, [search, statusFilter, templateFilter])

  useEffect(() => {
    const timer = setTimeout(fetchData, 300)
    return () => clearTimeout(timer)
  }, [fetchData])

  const stats = useMemo(() => {
    const total = environments.length
    const ready = environments.filter((e) => e.status === 'ready').length
    const errorCount = environments.filter((e) => e.status === 'error').length
    const active = environments.filter((e) =>
      ['planning', 'applying', 'destroying'].includes(e.status),
    ).length
    return { total, ready, error: errorCount, active }
  }, [environments])

  return (
    <Card>
      <CardHeader>
        <CardTitle>All Environments</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Summary stats */}
        <div className="grid grid-cols-4 gap-4">
          {[
            { label: 'Total', value: stats.total },
            { label: 'Ready', value: stats.ready },
            { label: 'Error', value: stats.error },
            { label: 'Active', value: stats.active },
          ].map((stat) => (
            <div
              key={stat.label}
              className="rounded-lg border p-3 text-center"
            >
              <p className="text-2xl font-bold">{stat.value}</p>
              <p className="text-muted-foreground text-xs">{stat.label}</p>
            </div>
          ))}
        </div>

        {/* Filters */}
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Search className="text-muted-foreground absolute left-2.5 top-2.5 h-4 w-4" />
            <Input
              placeholder="Search by name..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8"
            />
          </div>
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              {ALL_STATUSES.map((s) => (
                <SelectItem key={s} value={s}>
                  {s.charAt(0).toUpperCase() + s.slice(1)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={templateFilter} onValueChange={setTemplateFilter}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="Template" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All templates</SelectItem>
              {templates.map((t) => (
                <SelectItem key={t.id} value={t.id}>
                  {t.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {error && <p className="text-sm text-destructive">{error}</p>}

        {/* Table */}
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Template</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Creator</TableHead>
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
                environments.map((env) => (
                  <TableRow key={env.id}>
                    <TableCell className="font-medium">{env.name}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {env.template_name || '—'}
                    </TableCell>
                    <TableCell>
                      <EnvironmentStatusBadge status={env.status} />
                    </TableCell>
                    <TableCell>{env.created_by_name}</TableCell>
                    <TableCell>{formatDate(env.created_at)}</TableCell>
                    <TableCell>
                      <EnvironmentActions
                        environment={env}
                        currentUserId={user?.id ?? ''}
                        isAdmin={true}
                        onRefresh={fetchData}
                      />
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  )
}
