import { useEffect, useState } from 'react'
import { Lock } from 'lucide-react'
import { getEnvironmentOutputs } from '@/lib/environments-api'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { ApiError, EnvironmentOutputs as OutputsType } from '@/types/api'

function formatValue(value: unknown): string {
  if (typeof value === 'string') return value
  return JSON.stringify(value, null, 2)
}

interface EnvironmentOutputsProps {
  environmentId: string
}

export default function EnvironmentOutputs({
  environmentId,
}: EnvironmentOutputsProps) {
  const [outputs, setOutputs] = useState<OutputsType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function fetch() {
      setLoading(true)
      setError('')
      try {
        const data = await getEnvironmentOutputs(environmentId)
        if (!cancelled) setOutputs(data)
      } catch (err) {
        if (!cancelled) {
          const apiError = err as ApiError
          setError(apiError.message || 'Failed to load outputs')
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    fetch()
    return () => {
      cancelled = true
    }
  }, [environmentId])

  if (loading) {
    return (
      <div className="space-y-2 py-2">
        <Skeleton className="h-4 w-48" />
        <Skeleton className="h-4 w-64" />
        <Skeleton className="h-4 w-40" />
      </div>
    )
  }

  if (error) {
    return <p className="text-sm text-destructive py-2">{error}</p>
  }

  if (!outputs || Object.keys(outputs).length === 0) {
    return (
      <p className="text-muted-foreground py-2 text-sm">
        No outputs defined for this environment.
      </p>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[200px]">Name</TableHead>
          <TableHead>Value</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Object.entries(outputs).map(([key, output]) => (
          <TableRow key={key}>
            <TableCell className="font-medium">{key}</TableCell>
            <TableCell>
              {output.sensitive ? (
                <span className="text-muted-foreground flex items-center gap-1 italic">
                  <Lock className="h-3 w-3" />
                  (sensitive)
                </span>
              ) : (
                <pre className="whitespace-pre-wrap text-sm">
                  {formatValue(output.value)}
                </pre>
              )}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
