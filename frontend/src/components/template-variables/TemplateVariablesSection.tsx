import { useCallback, useEffect, useState } from 'react'
import { Plus, RefreshCw, Pencil, Trash2, Lock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  listTemplateVariables,
  parseTemplateVariables,
} from '@/lib/template-variables-api'
import CreateVariableDialog from '@/components/template-variables/CreateVariableDialog'
import EditVariableDialog from '@/components/template-variables/EditVariableDialog'
import DeleteVariableDialog from '@/components/template-variables/DeleteVariableDialog'
import type { ApiError, TemplateVariable } from '@/types/api'

interface TemplateVariablesSectionProps {
  templateId: string
}

export default function TemplateVariablesSection({
  templateId,
}: TemplateVariablesSectionProps) {
  const [variables, setVariables] = useState<TemplateVariable[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const [createOpen, setCreateOpen] = useState(false)
  const [editVariable, setEditVariable] = useState<TemplateVariable | null>(null)
  const [deleteVariable, setDeleteVariable] = useState<TemplateVariable | null>(null)

  const [parsing, setParsing] = useState(false)
  const [parseMessage, setParseMessage] = useState('')

  const fetchVariables = useCallback(async () => {
    setError('')
    setLoading(true)
    try {
      const data = await listTemplateVariables(templateId)
      data ? setVariables(data) : setVariables([])
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to load variables')
    } finally {
      setLoading(false)
    }
  }, [templateId])

  useEffect(() => {
    fetchVariables()
  }, [fetchVariables])

  async function handleParse() {
    setParsing(true)
    setParseMessage('')
    setError('')
    try {
      const result = await parseTemplateVariables(templateId)
      setParseMessage(
        `Added ${result.added}, Updated ${result.updated}, Removed ${result.removed}`,
      )
      await fetchVariables()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to parse variables')
    } finally {
      setParsing(false)
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleParse}
            disabled={parsing}
          >
            <RefreshCw className={`mr-2 h-4 w-4 ${parsing ? 'animate-spin' : ''}`} />
            {parsing ? 'Parsing...' : 'Parse & Reconcile'}
          </Button>
          {parseMessage && (
            <span className="text-sm text-muted-foreground">{parseMessage}</span>
          )}
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Variable
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Key</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Default</TableHead>
              <TableHead>Required</TableHead>
              <TableHead>Sensitive</TableHead>
              <TableHead>Source</TableHead>
              <TableHead className="w-[100px]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-20" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                  <TableCell><Skeleton className="h-4 w-16" /></TableCell>
                </TableRow>
              ))
            ) : variables.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className="text-muted-foreground h-24 text-center"
                >
                  No variables yet. Use "Parse & Reconcile" to detect variables from
                  template files, or add one manually.
                </TableCell>
              </TableRow>
            ) : (
              variables.map((v) => (
                <TableRow key={v.id}>
                  <TableCell className="font-medium font-mono text-sm">
                    {v.key}
                  </TableCell>
                  <TableCell>{v.var_type}</TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {v.default_value || '—'}
                  </TableCell>
                  <TableCell>
                    {v.is_required && <Badge variant="secondary">Required</Badge>}
                  </TableCell>
                  <TableCell>
                    {v.is_sensitive && (
                      <Badge variant="outline" className="gap-1">
                        <Lock className="h-3 w-3" />
                        Sensitive
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    {v.is_auto_parsed && (
                      <Badge variant="outline">Auto</Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setEditVariable(v)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteVariable(v)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <CreateVariableDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        templateId={templateId}
        onSuccess={fetchVariables}
      />

      <EditVariableDialog
        variable={editVariable}
        open={editVariable !== null}
        onOpenChange={(open) => {
          if (!open) setEditVariable(null)
        }}
        templateId={templateId}
        onSuccess={fetchVariables}
      />

      <DeleteVariableDialog
        variable={deleteVariable}
        open={deleteVariable !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteVariable(null)
        }}
        templateId={templateId}
        onSuccess={fetchVariables}
      />
    </div>
  )
}
