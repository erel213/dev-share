import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { useAppSelector } from '@/store'
import { selectUser } from '@/store/authSlice'
import { getWorkspaceTemplates } from '@/lib/templates-api'
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
import CreateTemplateDialog from '@/components/templates/CreateTemplateDialog'
import EditTemplateDialog from '@/components/templates/EditTemplateDialog'
import DeleteTemplateDialog from '@/components/templates/DeleteTemplateDialog'
import type { ApiError, Template } from '@/types/api'

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(iso))
}

export default function TemplatesPage() {
  const user = useAppSelector(selectUser)

  const [templates, setTemplates] = useState<Template[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [createOpen, setCreateOpen] = useState(false)
  const [editTemplate, setEditTemplate] = useState<Template | null>(null)
  const [deleteTemplate, setDeleteTemplate] = useState<Template | null>(null)

  const fetchTemplates = useCallback(async () => {
    if (!user) return
    setError('')
    setLoading(true)
    try {
      const data = await getWorkspaceTemplates(user.workspaceId)
      data ? setTemplates(data) : setTemplates([])
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load templates')
    } finally {
      setLoading(false)
    }
  }, [user])

  useEffect(() => {
    fetchTemplates()
  }, [fetchTemplates])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Templates</h1>
          <p className="text-muted-foreground mt-2">
            Browse and manage environment templates.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Template
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead className="w-[100px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell>
                    <Skeleton className="h-4 w-32" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-28" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-28" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-16" />
                  </TableCell>
                </TableRow>
              ))
            ) : templates.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={4}
                  className="text-muted-foreground h-24 text-center"
                >
                  No templates yet. Create one to get started.
                </TableCell>
              </TableRow>
            ) : (
              templates.map((t) => (
                <TableRow key={t.id}>
                  <TableCell className="font-medium">
                    <Link
                      to={`/templates/${t.id}`}
                      className="hover:underline"
                    >
                      {t.name}
                    </Link>
                  </TableCell>
                  <TableCell>{formatDate(t.created_at)}</TableCell>
                  <TableCell>{formatDate(t.updated_at)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setEditTemplate(t)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteTemplate(t)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {user && (
        <CreateTemplateDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
          workspaceId={user.workspaceId}
          onSuccess={fetchTemplates}
        />
      )}

      <EditTemplateDialog
        template={editTemplate}
        open={editTemplate !== null}
        onOpenChange={(open) => {
          if (!open) setEditTemplate(null)
        }}
        onSuccess={fetchTemplates}
      />

      <DeleteTemplateDialog
        template={deleteTemplate}
        open={deleteTemplate !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteTemplate(null)
        }}
        onSuccess={fetchTemplates}
      />
    </div>
  )
}
