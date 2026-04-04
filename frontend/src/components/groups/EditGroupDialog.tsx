import { useEffect, useState } from 'react'
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
import { updateGroup } from '@/lib/groups-api'
import type { ApiError, Group } from '@/types/api'

interface EditGroupDialogProps {
  group: Group | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export default function EditGroupDialog({
  group,
  open,
  onOpenChange,
  onSuccess,
}: EditGroupDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [accessAllTemplates, setAccessAllTemplates] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (group) {
      setName(group.name)
      setDescription(group.description)
      setAccessAllTemplates(group.access_all_templates)
    }
  }, [group])

  function handleClose(openState: boolean) {
    if (!openState) setError('')
    onOpenChange(openState)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!group) return
    setError('')
    setLoading(true)

    try {
      await updateGroup(group.id, {
        name,
        description,
        access_all_templates: accessAllTemplates,
      })
      onSuccess()
      handleClose(false)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to update group')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Group</DialogTitle>
          <DialogDescription>
            Update the group details.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-group-name">Name</Label>
            <Input
              id="edit-group-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-group-description">Description</Label>
            <Input
              id="edit-group-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="edit-access-all"
              checked={accessAllTemplates}
              onCheckedChange={(checked) =>
                setAccessAllTemplates(checked === true)
              }
            />
            <Label htmlFor="edit-access-all" className="cursor-pointer">
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
              {loading ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
