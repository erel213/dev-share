import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
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
import { inviteUser } from '@/lib/users-api'
import type { ApiError, InviteUserResponse } from '@/types/api'

interface InviteUserDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
}

export default function InviteUserDialog({
  open,
  onOpenChange,
  onSuccess,
}: InviteUserDialogProps) {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('user')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<InviteUserResponse | null>(null)
  const [copied, setCopied] = useState(false)

  function reset() {
    setName('')
    setEmail('')
    setRole('user')
    setError('')
    setLoading(false)
    setResult(null)
    setCopied(false)
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
      const response = await inviteUser({ name, email, role })
      setResult(response)
      onSuccess()
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to invite user')
    } finally {
      setLoading(false)
    }
  }

  async function handleCopy() {
    if (!result) return
    await navigator.clipboard.writeText(result.password)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {result ? 'User Invited' : 'Invite User'}
          </DialogTitle>
          <DialogDescription>
            {result
              ? 'Share the generated password with the new user.'
              : 'Create a new user account. A password will be generated automatically.'}
          </DialogDescription>
        </DialogHeader>

        {result ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <p className="text-sm">
                <span className="text-muted-foreground">Name:</span>{' '}
                {result.name}
              </p>
              <p className="text-sm">
                <span className="text-muted-foreground">Email:</span>{' '}
                {result.email}
              </p>
              <p className="text-sm">
                <span className="text-muted-foreground">Role:</span>{' '}
                {result.role}
              </p>
            </div>
            <div className="space-y-2">
              <Label>Generated Password</Label>
              <div className="flex gap-2">
                <Input value={result.password} readOnly className="font-mono" />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={handleCopy}
                  type="button"
                >
                  {copied ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
              <p className="text-destructive text-xs font-medium">
                This password will not be shown again.
              </p>
            </div>
            <DialogFooter>
              <Button onClick={() => handleClose(false)}>Done</Button>
            </DialogFooter>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="invite-name">Name</Label>
              <Input
                id="invite-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="John Doe"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="invite-email">Email</Label>
              <Input
                id="invite-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="john@example.com"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="invite-role">Role</Label>
              <Select value={role} onValueChange={setRole}>
                <SelectTrigger id="invite-role">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="editor">Editor</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
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
                {loading ? 'Inviting...' : 'Invite'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
