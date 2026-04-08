import { useState } from 'react'
import { MoreHorizontal, Play, Rocket, Flame, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  planEnvironment,
  applyEnvironment,
  destroyEnvironment,
} from '@/lib/environments-api'
import DeleteEnvironmentDialog from '@/components/environments/DeleteEnvironmentDialog'
import type { ApiError, Environment } from '@/types/api'

interface EnvironmentActionsProps {
  environment: Environment
  currentUserId: string
  isAdmin: boolean
  onRefresh: () => void
}

const inProgressStatuses = new Set(['planning', 'applying', 'destroying'])

function getAvailableActions(status: Environment['status']) {
  switch (status) {
    case 'pending':
      return { plan: false, apply: false, destroy: false, delete: true }
    case 'initialized':
      return { plan: true, apply: true, destroy: false, delete: true }
    case 'ready':
      return { plan: true, apply: true, destroy: true, delete: true }
    case 'destroyed':
      return { plan: false, apply: false, destroy: false, delete: true }
    case 'error':
      return { plan: true, apply: true, destroy: true, delete: true }
    default:
      return { plan: false, apply: false, destroy: false, delete: false }
  }
}

export default function EnvironmentActions({
  environment,
  currentUserId,
  isAdmin,
  onRefresh,
}: EnvironmentActionsProps) {
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [error, setError] = useState('')

  const isOwner = environment.created_by === currentUserId
  const canAct = isOwner || isAdmin
  const isInProgress = inProgressStatuses.has(environment.status)

  if (!canAct || isInProgress) return null

  const actions = getAvailableActions(environment.status)
  const hasAnyAction =
    actions.plan || actions.apply || actions.destroy || actions.delete
  if (!hasAnyAction) return null

  async function handleAction(
    action: (id: string) => Promise<Environment>,
  ) {
    setError('')
    try {
      await action(environment.id)
      onRefresh()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Operation failed')
    }
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" title="Actions">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          {actions.plan && (
            <DropdownMenuItem onClick={() => handleAction(planEnvironment)}>
              <Play className="mr-2 h-4 w-4" />
              Plan
            </DropdownMenuItem>
          )}
          {actions.apply && (
            <DropdownMenuItem onClick={() => handleAction(applyEnvironment)}>
              <Rocket className="mr-2 h-4 w-4" />
              Apply
            </DropdownMenuItem>
          )}
          {actions.destroy && (
            <DropdownMenuItem onClick={() => handleAction(destroyEnvironment)}>
              <Flame className="mr-2 h-4 w-4" />
              Destroy
            </DropdownMenuItem>
          )}
          {actions.delete && (
            <DropdownMenuItem
              className="text-destructive focus:text-destructive"
              onClick={() => setDeleteOpen(true)}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {error && (
        <p className="text-xs text-destructive absolute mt-1">{error}</p>
      )}

      <DeleteEnvironmentDialog
        environment={environment}
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        onSuccess={onRefresh}
      />
    </>
  )
}
