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
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { createTemplateVariable } from '@/lib/template-variables-api'
import type { ApiError } from '@/types/api'

interface CreateVariableDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  templateId: string
  onSuccess: () => void
}

const VAR_TYPES = ['string', 'number', 'bool', 'list', 'map']

export default function CreateVariableDialog({
  open,
  onOpenChange,
  templateId,
  onSuccess,
}: CreateVariableDialogProps) {
  const [key, setKey] = useState('')
  const [description, setDescription] = useState('')
  const [varType, setVarType] = useState('string')
  const [defaultValue, setDefaultValue] = useState('')
  const [isSensitive, setIsSensitive] = useState(false)
  const [isRequired, setIsRequired] = useState(false)
  const [validationRegex, setValidationRegex] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  function reset() {
    setKey('')
    setDescription('')
    setVarType('string')
    setDefaultValue('')
    setIsSensitive(false)
    setIsRequired(false)
    setValidationRegex('')
    setError('')
    setLoading(false)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await createTemplateVariable(templateId, {
        key,
        description: description || undefined,
        var_type: varType,
        default_value: defaultValue || undefined,
        is_sensitive: isSensitive,
        is_required: isRequired,
        validation_regex: validationRegex || undefined,
      })
      reset()
      onOpenChange(false)
      onSuccess()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to create variable')
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
          <DialogTitle>Add Variable</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="var-key">Key</Label>
            <Input
              id="var-key"
              value={key}
              onChange={(e) => setKey(e.target.value)}
              placeholder="MY_VARIABLE"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="var-description">Description</Label>
            <Input
              id="var-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="var-type">Type</Label>
            <Select value={varType} onValueChange={setVarType}>
              <SelectTrigger id="var-type">
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
            <Label htmlFor="var-default">Default Value</Label>
            <Input
              id="var-default"
              value={defaultValue}
              onChange={(e) => setDefaultValue(e.target.value)}
              placeholder="Optional default"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="var-regex">Validation Regex</Label>
            <Input
              id="var-regex"
              value={validationRegex}
              onChange={(e) => setValidationRegex(e.target.value)}
              placeholder="Optional regex pattern"
            />
          </div>
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2">
              <Checkbox
                id="var-required"
                checked={isRequired}
                onCheckedChange={(v) => setIsRequired(v === true)}
              />
              <Label htmlFor="var-required" className="text-sm font-normal">
                Required
              </Label>
            </div>
            <div className="flex items-center gap-2">
              <Checkbox
                id="var-sensitive"
                checked={isSensitive}
                onCheckedChange={(v) => setIsSensitive(v === true)}
              />
              <Label htmlFor="var-sensitive" className="text-sm font-normal">
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
              {loading ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
