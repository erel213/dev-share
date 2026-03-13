import { useEffect, useState } from 'react'
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
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { updateTemplateVariable } from '@/lib/template-variables-api'
import type { ApiError, TemplateVariable } from '@/types/api'

interface EditVariableDialogProps {
  variable: TemplateVariable | null
  open: boolean
  onOpenChange: (open: boolean) => void
  templateId: string
  onSuccess: () => void
}

const VAR_TYPES = ['string', 'number', 'bool', 'list', 'map']

export default function EditVariableDialog({
  variable,
  open,
  onOpenChange,
  templateId,
  onSuccess,
}: EditVariableDialogProps) {
  const [description, setDescription] = useState('')
  const [varType, setVarType] = useState('string')
  const [defaultValue, setDefaultValue] = useState('')
  const [isSensitive, setIsSensitive] = useState(false)
  const [isRequired, setIsRequired] = useState(false)
  const [validationRegex, setValidationRegex] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (variable) {
      setDescription(variable.description)
      setVarType(variable.var_type)
      setDefaultValue(variable.default_value)
      setIsSensitive(variable.is_sensitive)
      setIsRequired(variable.is_required)
      setValidationRegex(variable.validation_regex)
      setError('')
    }
  }, [variable])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!variable) return
    setError('')
    setLoading(true)

    try {
      await updateTemplateVariable(templateId, variable.id, {
        description,
        var_type: varType,
        default_value: defaultValue,
        is_sensitive: isSensitive,
        is_required: isRequired,
        validation_regex: validationRegex || undefined,
      })
      onOpenChange(false)
      onSuccess()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to update variable')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Variable</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-var-key">Key</Label>
            <Input
              id="edit-var-key"
              value={variable?.key ?? ''}
              disabled
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-var-description">Description</Label>
            <Input
              id="edit-var-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-var-type">Type</Label>
            <Select value={varType} onValueChange={setVarType}>
              <SelectTrigger id="edit-var-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {VAR_TYPES.map((t) => (
                  <SelectItem key={t} value={t}>
                    {t}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-var-default">Default Value</Label>
            <Input
              id="edit-var-default"
              value={defaultValue}
              onChange={(e) => setDefaultValue(e.target.value)}
              placeholder="Optional default"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-var-regex">Validation Regex</Label>
            <Input
              id="edit-var-regex"
              value={validationRegex}
              onChange={(e) => setValidationRegex(e.target.value)}
              placeholder="Optional regex pattern"
            />
          </div>
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2">
              <Checkbox
                id="edit-var-required"
                checked={isRequired}
                onCheckedChange={(v) => setIsRequired(v === true)}
              />
              <Label htmlFor="edit-var-required" className="text-sm font-normal">
                Required
              </Label>
            </div>
            <div className="flex items-center gap-2">
              <Checkbox
                id="edit-var-sensitive"
                checked={isSensitive}
                onCheckedChange={(v) => setIsSensitive(v === true)}
              />
              <Label htmlFor="edit-var-sensitive" className="text-sm font-normal">
                Sensitive
              </Label>
            </div>
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
            <Button type="submit" disabled={loading}>
              {loading ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
