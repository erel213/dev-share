import { createBrowserRouter } from 'react-router-dom'
import HomePage from '@/pages/HomePage'
import SetupPage from '@/pages/SetupPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <HomePage />,
  },
  {
    path: '/setup',
    element: <SetupPage />,
  },
])
