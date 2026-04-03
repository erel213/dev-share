import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { resetUserPassword } from '@/lib/users-api'
import type { AdminUser, ApiError } from '@/types/api'

interface ResetPasswordDialogProps {
  user: AdminUser | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function ResetPasswordDialog({
  user,
  open,
  onOpenChange,
}: ResetPasswordDialogProps) {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [newPassword, setNewPassword] = useState('')
  const [copied, setCopied] = useState(false)

  function reset() {
    setError('')
    setLoading(false)
    setNewPassword('')
    setCopied(false)
  }

  function handleClose(openState: boolean) {
    if (!openState) reset()
    onOpenChange(openState)
  }

  async function handleReset() {
    if (!user) return
    setError('')
    setLoading(true)

    try {
      const response = await resetUserPassword(user.id)
      setNewPassword(response.password)
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to reset password')
    } finally {
      setLoading(false)
    }
  }

  async function handleCopy() {
    await navigator.clipboard.writeText(newPassword)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {newPassword ? 'Password Reset' : 'Reset Password'}
          </DialogTitle>
          <DialogDescription>
            {newPassword
              ? 'Share the new password with the user.'
              : `Reset the password for ${user?.name}?`}
          </DialogDescription>
        </DialogHeader>

        {newPassword ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>New Password</Label>
              <div className="flex gap-2">
                <Input value={newPassword} readOnly className="font-mono" />
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
          <div className="space-y-4">
            {error && <p className="text-sm text-destructive">{error}</p>}
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => handleClose(false)}
              >
                Cancel
              </Button>
              <Button onClick={handleReset} disabled={loading}>
                {loading ? 'Resetting...' : 'Reset Password'}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
