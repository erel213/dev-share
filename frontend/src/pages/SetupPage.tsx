import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import api from '@/lib/api'
import type {
  AdminInitRequest,
  AdminInitResponse,
  SystemStatus,
} from '@/types/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

type Step = 'welcome' | 'admin' | 'workspace' | 'review' | 'success'

export default function SetupPage() {
  const navigate = useNavigate()
  const [step, setStep] = useState<Step>('welcome')
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const [adminName, setAdminName] = useState('')
  const [adminEmail, setAdminEmail] = useState('')
  const [adminPassword, setAdminPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [workspaceName, setWorkspaceName] = useState('')
  const [workspaceDescription, setWorkspaceDescription] = useState('')

  useEffect(() => {
    api
      .get<SystemStatus>('/admin/status')
      .then((res) => {
        if (res.data.initialized) {
          navigate('/', { replace: true })
        } else {
          setLoading(false)
        }
      })
      .catch(() => setLoading(false))
  }, [navigate])

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </main>
    )
  }

  const handleSubmit = async () => {
    setError('')
    setSubmitting(true)
    try {
      const body: AdminInitRequest = {
        admin_name: adminName,
        admin_email: adminEmail,
        admin_password: adminPassword,
        workspace_name: workspaceName,
        workspace_description: workspaceDescription || undefined,
      }
      await api.post<AdminInitResponse>('/admin/init', body)
      setStep('success')
    } catch (err: unknown) {
      const apiErr = err as { message?: string }
      setError(apiErr.message ?? 'Failed to initialize system')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        {step === 'welcome' && (
          <>
            <CardHeader>
              <CardTitle className="text-2xl">Welcome to Dev-Share</CardTitle>
              <CardDescription>
                Let's set up your instance. This will create your admin account
                and first workspace.
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <Button className="w-full" onClick={() => setStep('admin')}>
                Get Started
              </Button>
            </CardFooter>
          </>
        )}

        {step === 'admin' && (
          <>
            <CardHeader>
              <CardTitle>Admin Account</CardTitle>
              <CardDescription>
                Create your administrator account.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  value={adminName}
                  onChange={(e) => setAdminName(e.target.value)}
                  placeholder="Your name"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={adminEmail}
                  onChange={(e) => setAdminEmail(e.target.value)}
                  placeholder="admin@example.com"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={adminPassword}
                  onChange={(e) => setAdminPassword(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="confirm-password">Confirm Password</Label>
                <Input
                  id="confirm-password"
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                />
              </div>
            </CardContent>
            <CardFooter className="flex justify-between">
              <Button variant="outline" onClick={() => setStep('welcome')}>
                Back
              </Button>
              <Button
                onClick={() => {
                  if (!adminName || !adminEmail || !adminPassword) {
                    setError('All fields are required')
                    return
                  }
                  if (adminPassword !== confirmPassword) {
                    setError('Passwords do not match')
                    return
                  }
                  setError('')
                  setStep('workspace')
                }}
              >
                Next
              </Button>
            </CardFooter>
          </>
        )}

        {step === 'workspace' && (
          <>
            <CardHeader>
              <CardTitle>Workspace</CardTitle>
              <CardDescription>
                Create your first workspace to organize environments.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="workspace-name">Workspace Name</Label>
                <Input
                  id="workspace-name"
                  value={workspaceName}
                  onChange={(e) => setWorkspaceName(e.target.value)}
                  placeholder="My Workspace"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="workspace-desc">Description (optional)</Label>
                <Input
                  id="workspace-desc"
                  value={workspaceDescription}
                  onChange={(e) => setWorkspaceDescription(e.target.value)}
                  placeholder="A brief description"
                />
              </div>
            </CardContent>
            <CardFooter className="flex justify-between">
              <Button variant="outline" onClick={() => setStep('admin')}>
                Back
              </Button>
              <Button
                onClick={() => {
                  if (!workspaceName) {
                    setError('Workspace name is required')
                    return
                  }
                  setError('')
                  setStep('review')
                }}
              >
                Next
              </Button>
            </CardFooter>
          </>
        )}

        {step === 'review' && (
          <>
            <CardHeader>
              <CardTitle>Review</CardTitle>
              <CardDescription>
                Confirm your setup details before proceeding.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div>
                <span className="text-muted-foreground">Admin Name:</span>{' '}
                {adminName}
              </div>
              <div>
                <span className="text-muted-foreground">Admin Email:</span>{' '}
                {adminEmail}
              </div>
              <div>
                <span className="text-muted-foreground">Workspace:</span>{' '}
                {workspaceName}
              </div>
              {workspaceDescription && (
                <div>
                  <span className="text-muted-foreground">Description:</span>{' '}
                  {workspaceDescription}
                </div>
              )}
            </CardContent>
            <CardFooter className="flex justify-between">
              <Button variant="outline" onClick={() => setStep('workspace')}>
                Back
              </Button>
              <Button onClick={handleSubmit} disabled={submitting}>
                {submitting ? 'Initializing...' : 'Initialize'}
              </Button>
            </CardFooter>
          </>
        )}

        {step === 'success' && (
          <>
            <CardHeader>
              <CardTitle>Setup Complete!</CardTitle>
              <CardDescription>
                Your Dev-Share instance is ready to use.
              </CardDescription>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              <p>
                Admin account <strong>{adminEmail}</strong> and workspace{' '}
                <strong>{workspaceName}</strong> have been created.
              </p>
            </CardContent>
            <CardFooter>
              <Button className="w-full" onClick={() => navigate('/')}>
                Go to Dashboard
              </Button>
            </CardFooter>
          </>
        )}

        {error && (
          <div className="px-6 pb-4">
            <p className="text-sm text-destructive">{error}</p>
          </div>
        )}
      </Card>
    </main>
  )
}
