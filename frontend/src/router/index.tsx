import { createBrowserRouter } from 'react-router-dom'
import HomePage from '@/pages/HomePage'
import SetupPage from '@/pages/SetupPage'
import LoginPage from '@/pages/LoginPage'
import WorkspacesPage from '@/pages/WorkspacesPage'
import TemplatesPage from '@/pages/TemplatesPage'
import TemplateBrowserPage from '@/pages/TemplateBrowserPage'
import UsersPage from '@/pages/UsersPage'
import ProtectedRoute from '@/components/layout/ProtectedRoute'
import AppLayout from '@/components/layout/AppLayout'

export const router = createBrowserRouter([
  {
    path: '/setup',
    element: <SetupPage />,
  },
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          {
            path: '/',
            element: <HomePage />,
          },
          {
            path: '/workspaces',
            element: <WorkspacesPage />,
          },
          {
            path: '/templates',
            element: <TemplatesPage />,
          },
          {
            path: '/templates/:id',
            element: <TemplateBrowserPage />,
          },
          {
            path: '/users',
            element: <UsersPage />,
          },
        ],
      },
    ],
  },
])
