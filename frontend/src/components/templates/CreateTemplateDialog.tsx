import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import FileDropzone, { type FileWithPath } from '@/components/templates/FileDropzone'
import { createTemplate } from '@/lib/templates-api'
import type { ApiError } from '@/types/api'

interface CreateTemplateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  workspaceId: string
  onSuccess: () => void
}

export default function CreateTemplateDialog({
  open,
  onOpenChange,
  workspaceId,
  onSuccess,
}: CreateTemplateDialogProps) {
  const [name, setName] = useState('')
  const [files, setFiles] = useState<FileWithPath[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  function reset() {
    setName('')
    setFiles([])
    setError('')
    setLoading(false)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await createTemplate(name, workspaceId, files)
      reset()
      onOpenChange(false)
      onSuccess()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to create template')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!v) reset()
        onOpenChange(v)
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Template</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="template-name">Name</Label>
            <Input
              id="template-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-template"
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Files</Label>
            <FileDropzone files={files} onFilesChange={setFiles} />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={loading || files.length === 0}>
              {loading ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
