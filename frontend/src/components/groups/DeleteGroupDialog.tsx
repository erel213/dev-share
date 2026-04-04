import { useState } from 'react'
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { deleteGroup } from '@/lib/groups-api'
import type { ApiError, Group } from '@/types/api'

interface DeleteGroupDialogProps {
  group: Group | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export default function DeleteGroupDialog({
  group,
  open,
  onOpenChange,
  onSuccess,
}: DeleteGroupDialogProps) {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleDelete() {
    if (!group) return
    setError('')
    setLoading(true)

    try {
      await deleteGroup(group.id)
      onOpenChange(false)
      onSuccess()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to delete group')
    } finally {
      setLoading(false)
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Group</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete{' '}
            <span className="font-medium">{group?.name}</span>? All
            memberships and template access for this group will be removed.
            This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        {error && <p className="text-sm text-destructive">{error}</p>}
        <AlertDialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={loading}
          >
            {loading ? 'Deleting...' : 'Delete'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
