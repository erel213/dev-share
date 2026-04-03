import { useCallback, useEffect, useState } from 'react'
import { Plus, KeyRound, Trash2 } from 'lucide-react'
import { listUsers } from '@/lib/users-api'
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
import InviteUserDialog from '@/components/users/InviteUserDialog'
import ResetPasswordDialog from '@/components/users/ResetPasswordDialog'
import DeleteUserDialog from '@/components/users/DeleteUserDialog'
import type { AdminUser, ApiError } from '@/types/api'

function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(iso))
}

function roleBadgeVariant(
  role: string,
): 'default' | 'secondary' | 'outline' {
  switch (role) {
    case 'admin':
      return 'default'
    case 'editor':
      return 'secondary'
    default:
      return 'outline'
  }
}

export default function UsersPage() {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [inviteOpen, setInviteOpen] = useState(false)
  const [resetUser, setResetUser] = useState<AdminUser | null>(null)
  const [deleteUserTarget, setDeleteUserTarget] = useState<AdminUser | null>(
    null,
  )

  const fetchUsers = useCallback(async () => {
    setError('')
    setLoading(true)
    try {
      const data = await listUsers()
      data ? setUsers(data) : setUsers([])
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load users')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground mt-2">
            Manage user accounts and permissions.
          </p>
        </div>
        <Button onClick={() => setInviteOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Invite User
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="w-[100px]">Actions</TableHead>
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
                    <Skeleton className="h-4 w-36" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-16" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-28" />
                  </TableCell>
                  <TableCell>
                    <Skeleton className="h-4 w-16" />
                  </TableCell>
                </TableRow>
              ))
            ) : users.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="text-muted-foreground h-24 text-center"
                >
                  No users found.
                </TableCell>
              </TableRow>
            ) : (
              users.map((u) => (
                <TableRow key={u.id}>
                  <TableCell className="font-medium">{u.name}</TableCell>
                  <TableCell>{u.email}</TableCell>
                  <TableCell>
                    <Badge variant={roleBadgeVariant(u.role)}>{u.role}</Badge>
                  </TableCell>
                  <TableCell>{formatDate(u.created_at)}</TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setResetUser(u)}
                        title="Reset password"
                      >
                        <KeyRound className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteUserTarget(u)}
                        title="Delete user"
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

      <InviteUserDialog
        open={inviteOpen}
        onOpenChange={setInviteOpen}
        onSuccess={fetchUsers}
      />

      <ResetPasswordDialog
        user={resetUser}
        open={resetUser !== null}
        onOpenChange={(open) => {
          if (!open) setResetUser(null)
        }}
      />

      <DeleteUserDialog
        user={deleteUserTarget}
        open={deleteUserTarget !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteUserTarget(null)
        }}
        onSuccess={fetchUsers}
      />
    </div>
  )
}
