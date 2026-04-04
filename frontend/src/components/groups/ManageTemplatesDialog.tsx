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
  getGroupTemplateAccess,
  addGroupTemplateAccess,
  removeGroupTemplateAccess,
  listAllTemplates,
} from '@/lib/groups-api'
import type { ApiError, Group, Template } from '@/types/api'

interface ManageTemplatesDialogProps {
  group: Group | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export default function ManageTemplatesDialog({
  group,
  open,
  onOpenChange,
  onSuccess,
}: ManageTemplatesDialogProps) {
  const [templates, setTemplates] = useState<Template[]>([])
  const [accessIds, setAccessIds] = useState<Set<string>>(new Set())
  const [originalAccessIds, setOriginalAccessIds] = useState<Set<string>>(
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
      const [allTemplates, access] = await Promise.all([
        listAllTemplates(),
        getGroupTemplateAccess(group.id),
      ])
      setTemplates(allTemplates || [])
      const ids = new Set(access || [])
      setAccessIds(ids)
      setOriginalAccessIds(ids)
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

  function toggleTemplate(templateId: string) {
    setAccessIds((prev) => {
      const next = new Set(prev)
      if (next.has(templateId)) {
        next.delete(templateId)
      } else {
        next.add(templateId)
      }
      return next
    })
  }

  async function handleSave() {
    if (!group) return
    setSaving(true)
    setError('')

    try {
      const toAdd = [...accessIds].filter(
        (id) => !originalAccessIds.has(id),
      )
      const toRemove = [...originalAccessIds].filter(
        (id) => !accessIds.has(id),
      )

      if (toAdd.length > 0) {
        await addGroupTemplateAccess(group.id, toAdd)
      }
      for (const templateId of toRemove) {
        await removeGroupTemplateAccess(group.id, templateId)
      }

      onSuccess()
      onOpenChange(false)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to save template access')
    } finally {
      setSaving(false)
    }
  }

  const isAccessAll = group?.access_all_templates === true

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Manage Templates — {group?.name}</DialogTitle>
          <DialogDescription>
            {isAccessAll
              ? 'This group has access to all templates. Disable "Access all templates" in the group settings to manage individual template access.'
              : 'Select templates this group can access.'}
          </DialogDescription>
        </DialogHeader>

        {error && <p className="text-sm text-destructive">{error}</p>}

        <div className="space-y-2">
          {loading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-8 w-full" />
            ))
          ) : templates.length === 0 ? (
            <p className="text-muted-foreground text-sm">
              No templates found.
            </p>
          ) : (
            templates.map((t) => (
              <div key={t.id} className="flex items-center space-x-2">
                <Checkbox
                  id={`template-${t.id}`}
                  checked={isAccessAll || accessIds.has(t.id)}
                  onCheckedChange={() => toggleTemplate(t.id)}
                  disabled={isAccessAll}
                />
                <Label
                  htmlFor={`template-${t.id}`}
                  className="cursor-pointer text-sm"
                >
                  {t.name}
                </Label>
              </div>
            ))
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={saving || loading || isAccessAll}
          >
            {saving ? 'Saving...' : 'Save'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
