import { useCallback, useEffect, useState } from 'react'
import { Plus, Pencil, Trash2, Users, LayoutTemplate } from 'lucide-react'
import { listGroups } from '@/lib/groups-api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import CreateGroupDialog from '@/components/groups/CreateGroupDialog'
import EditGroupDialog from '@/components/groups/EditGroupDialog'
import DeleteGroupDialog from '@/components/groups/DeleteGroupDialog'
import ManageMembersDialog from '@/components/groups/ManageMembersDialog'
import ManageTemplatesDialog from '@/components/groups/ManageTemplatesDialog'
import type { ApiError, Group } from '@/types/api'

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(iso))
}

export default function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [createOpen, setCreateOpen] = useState(false)
  const [editGroup, setEditGroup] = useState<Group | null>(null)
  const [deleteGroup, setDeleteGroup] = useState<Group | null>(null)
  const [membersGroup, setMembersGroup] = useState<Group | null>(null)
  const [templatesGroup, setTemplatesGroup] = useState<Group | null>(null)

  const fetchGroups = useCallback(async () => {
    setError('')
    setLoading(true)
    try {
      const data = await listGroups()
      data ? setGroups(data) : setGroups([])
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load groups')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchGroups()
  }, [fetchGroups])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Groups</h1>
          <p className="text-muted-foreground mt-2">
            Manage groups and their template access permissions.
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Group
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Template Access</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="w-[160px]">Actions</TableHead>
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
                    <Skeleton className="h-4 w-40" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-20" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-28" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-24" />
                  </TableCell>
                </TableRow>
              ))
            ) : groups.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="text-muted-foreground h-24 text-center"
                >
                  No groups found.
                </TableCell>
              </TableRow>
            ) : (
              groups.map((g) => (
                <TableRow key={g.id}>
                  <TableCell className="font-medium">{g.name}</TableCell>
                  <TableCell className="text-muted-foreground">
                    {g.description || '—'}
                  </TableCell>
                  <TableCell>
                    {g.access_all_templates ? (
                      <Badge variant="default">All Templates</Badge>
                    ) : (
                      <Badge variant="secondary">Custom</Badge>
                    )}
                  </TableCell>
                  <TableCell>{formatDate(g.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setEditGroup(g)}
                        title="Edit group"
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setMembersGroup(g)}
                        title="Manage members"
                      >
                        <Users className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setTemplatesGroup(g)}
                        title="Manage templates"
                      >
                        <LayoutTemplate className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteGroup(g)}
                        title="Delete group"
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

      <CreateGroupDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onSuccess={fetchGroups}
      />

      <EditGroupDialog
        group={editGroup}
        open={editGroup !== null}
        onOpenChange={(open) => {
          if (!open) setEditGroup(null)
        }}
        onSuccess={fetchGroups}
      />

      <DeleteGroupDialog
        group={deleteGroup}
        open={deleteGroup !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteGroup(null)
        }}
        onSuccess={fetchGroups}
      />

      <ManageMembersDialog
        group={membersGroup}
        open={membersGroup !== null}
        onOpenChange={(open) => {
          if (!open) setMembersGroup(null)
        }}
        onSuccess={fetchGroups}
      />

      <ManageTemplatesDialog
        group={templatesGroup}
        open={templatesGroup !== null}
        onOpenChange={(open) => {
          if (!open) setTemplatesGroup(null)
        }}
        onSuccess={fetchGroups}
      />
    </div>
  )
}
