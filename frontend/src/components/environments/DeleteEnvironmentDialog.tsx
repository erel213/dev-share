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
import { deleteEnvironment } from '@/lib/environments-api'
import type { ApiError, Environment } from '@/types/api'

interface DeleteEnvironmentDialogProps {
  environment: Environment | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export default function DeleteEnvironmentDialog({
  environment,
  open,
  onOpenChange,
  onSuccess,
}: DeleteEnvironmentDialogProps) {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleDelete() {
    if (!environment) return
    setError('')
    setLoading(true)

    try {
      await deleteEnvironment(environment.id)
      onOpenChange(false)
      onSuccess()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to delete environment')
    } finally {
      setLoading(false)
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Environment</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete{' '}
            <span className="font-medium">{environment?.name}</span>? This
            action cannot be undone.
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
