import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { createGroup } from '@/lib/groups-api'
import type { ApiError } from '@/types/api'

interface CreateGroupDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export default function CreateGroupDialog({
  open,
  onOpenChange,
  onSuccess,
}: CreateGroupDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [accessAllTemplates, setAccessAllTemplates] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  function reset() {
    setName('')
    setDescription('')
    setAccessAllTemplates(false)
    setError('')
    setLoading(false)
  }

  function handleClose(openState: boolean) {
    if (!openState) reset()
    onOpenChange(openState)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await createGroup({
        name,
        description: description || undefined,
        access_all_templates: accessAllTemplates,
      })
      onSuccess()
      handleClose(false)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to create group')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Group</DialogTitle>
          <DialogDescription>
            Create a new group to manage template access for users.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="group-name">Name</Label>
            <Input
              id="group-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Engineering"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="group-description">Description</Label>
            <Input
              id="group-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Backend and frontend engineers"
            />
          </div>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="access-all"
              checked={accessAllTemplates}
              onCheckedChange={(checked) =>
                setAccessAllTemplates(checked === true)
              }
            />
            <Label htmlFor="access-all" className="cursor-pointer">
              Access all templates
            </Label>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleClose(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
