import { Badge } from '@/components/ui/badge'
import type { Environment } from '@/types/api'

type Status = Environment['status']

const statusConfig: Record<
  Status,
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline'; className?: string }
> = {
  pending: { label: 'Pending', variant: 'outline', className: 'border-yellow-500 text-yellow-600' },
  initialized: { label: 'Initialized', variant: 'secondary' },
  planning: { label: 'Planning', variant: 'secondary', className: 'animate-pulse' },
  applying: { label: 'Applying', variant: 'secondary', className: 'animate-pulse' },
  ready: { label: 'Ready', variant: 'default' },
  destroying: { label: 'Destroying', variant: 'outline', className: 'animate-pulse border-orange-500 text-orange-600' },
  destroyed: { label: 'Destroyed', variant: 'outline' },
  error: { label: 'Error', variant: 'destructive' },
}

export default function EnvironmentStatusBadge({ status }: { status: Status }) {
  const config = statusConfig[status] ?? { label: status, variant: 'outline' as const }
  return (
    <Badge variant={config.variant} className={config.className}>
      {config.label}
    </Badge>
  )
}
