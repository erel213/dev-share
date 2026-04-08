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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { createEnvironment } from '@/lib/environments-api'
import { listTemplates } from '@/lib/templates-api'
import type { ApiError, Template } from '@/types/api'

interface CreateEnvironmentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
  preselectedTemplateId?: string
}

const TTL_PRESETS = [
  { label: '1 hour', value: 3600 },
  { label: '4 hours', value: 14400 },
  { label: '24 hours', value: 86400 },
  { label: 'None', value: 0 },
]

export default function CreateEnvironmentDialog({
  open,
  onOpenChange,
  onSuccess,
  preselectedTemplateId,
}: CreateEnvironmentDialogProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [templateId, setTemplateId] = useState('')
  const [ttlPreset, setTtlPreset] = useState(0)
  const [templates, setTemplates] = useState<Template[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const fetchTemplates = useCallback(async () => {
    try {
      const data = await listTemplates()
      setTemplates(data ?? [])
    } catch {
      // Silently fail — user won't see templates but can retry
    }
  }, [])

  useEffect(() => {
    if (open) {
      fetchTemplates()
      if (preselectedTemplateId) {
        setTemplateId(preselectedTemplateId)
      }
    }
  }, [open, fetchTemplates, preselectedTemplateId])

  function reset() {
    setName('')
    setDescription('')
    setTemplateId(preselectedTemplateId ?? '')
    setTtlPreset(0)
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
      await createEnvironment({
        name,
        description: description || undefined,
        template_id: templateId,
        ttl_seconds: ttlPreset > 0 ? ttlPreset : undefined,
      })
      onSuccess()
      handleClose(false)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to create environment')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Environment</DialogTitle>
          <DialogDescription>
            Create a new development environment from a template.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="env-template">Template</Label>
            <Select
              value={templateId}
              onValueChange={setTemplateId}
              disabled={!!preselectedTemplateId}
            >
              <SelectTrigger id="env-template">
                <SelectValue placeholder="Select a template" />
              </SelectTrigger>
              <SelectContent>
                {templates.map((t) => (
                  <SelectItem key={t.id} value={t.id}>
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="env-name">Name</Label>
            <Input
              id="env-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="my-dev-environment"
              required
              minLength={3}
              maxLength={255}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="env-description">Description</Label>
            <Input
              id="env-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              maxLength={1000}
            />
          </div>

          <div className="space-y-2">
            <Label>Time to Live</Label>
            <div className="flex gap-2">
              {TTL_PRESETS.map((preset) => (
                <Button
                  key={preset.value}
                  type="button"
                  variant={ttlPreset === preset.value ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setTtlPreset(preset.value)}
                >
                  {preset.label}
                </Button>
              ))}
            </div>
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
            <Button type="submit" disabled={loading || !templateId}>
              {loading ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
