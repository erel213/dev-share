import { useCallback, useEffect, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import {
  getGroupMembers,
  addGroupMembers,
  removeGroupMember,
  listAllUsers,
} from '@/lib/groups-api'
import type { AdminUser, ApiError, Group } from '@/types/api'

interface ManageMembersDialogProps {
  group: Group | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export default function ManageMembersDialog({
  group,
  open,
  onOpenChange,
  onSuccess,
}: ManageMembersDialogProps) {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [memberIds, setMemberIds] = useState<Set<string>>(new Set())
  const [originalMemberIds, setOriginalMemberIds] = useState<Set<string>>(
    new Set(),
  )
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const fetchData = useCallback(async () => {
    if (!group) return
    setLoading(true)
    setError('')

    try {
      const [allUsers, members] = await Promise.all([
        listAllUsers(),
        getGroupMembers(group.id),
      ])
      setUsers(allUsers || [])
      const ids = new Set(members || [])
      setMemberIds(ids)
      setOriginalMemberIds(ids)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load data')
    } finally {
      setLoading(false)
    }
  }, [group])

  useEffect(() => {
    if (open && group) {
      fetchData()
    }
  }, [open, group, fetchData])

  function toggleMember(userId: string) {
    setMemberIds((prev) => {
      const next = new Set(prev)
      if (next.has(userId)) {
        next.delete(userId)
      } else {
        next.add(userId)
      }
      return next
    })
  }

  async function handleSave() {
    if (!group) return
    setSaving(true)
    setError('')

    try {
      const toAdd = [...memberIds].filter((id) => !originalMemberIds.has(id))
      const toRemove = [...originalMemberIds].filter(
        (id) => !memberIds.has(id),
      )

      if (toAdd.length > 0) {
        await addGroupMembers(group.id, toAdd)
      }
      for (const userId of toRemove) {
        await removeGroupMember(group.id, userId)
      }

      onSuccess()
      onOpenChange(false)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to save members')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Manage Members — {group?.name}</DialogTitle>
          <DialogDescription>
            Select users to include in this group.
          </DialogDescription>
        </DialogHeader>

        {error && <p className="text-sm text-destructive">{error}</p>}

        <div className="space-y-2">
          {loading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-8 w-full" />
            ))
          ) : users.length === 0 ? (
            <p className="text-muted-foreground text-sm">No users found.</p>
          ) : (
            users.map((u) => (
              <div key={u.id} className="flex items-center space-x-2">
                <Checkbox
                  id={`member-${u.id}`}
                  checked={memberIds.has(u.id)}
                  onCheckedChange={() => toggleMember(u.id)}
                />
                <Label
                  htmlFor={`member-${u.id}`}
                  className="cursor-pointer text-sm"
                >
                  {u.name}{' '}
                  <span className="text-muted-foreground">({u.email})</span>
                </Label>
              </div>
            ))
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || loading}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
