import React from 'react'
import { HashRouter, Route, Routes } from 'react-router-dom'
import { CssBaseline, CssVarsProvider } from '@mui/joy'
import ChatLayout from './pages/ChatLayout'
import SignInPage from './pages/SignInPage'
import SignUpPage from './pages/SignUpPage'
import { AuthProvider } from './contexts/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import { QueryClientProvider, QueryClient } from '@tanstack/react-query'
import MessagesPane from './pages/MessagesPane'

const queryClient = new QueryClient()

function App() {
  return (
    <AuthProvider>
      <CssVarsProvider>
        <CssBaseline disableTransitionOnChange />
        <AuthProvider>
          <QueryClientProvider client={queryClient}>
            <HashRouter>
              <Routes>
                <Route element={<ProtectedRoute />}>
                  <Route path="/" element={<ChatLayout />} >
                    {/*<Route index element={<WelcomePane />} />*/}
                    <Route path=":conversationId" element={<MessagesPane />} />
                  </Route>
                </Route>
                <Route path="/signin" element={<SignInPage />} />
                <Route path="/signup" element={<SignUpPage />} />
              </Routes>
            </HashRouter>
          </QueryClientProvider>
        </AuthProvider>
      </CssVarsProvider>
    </AuthProvider>
  )
}

export default App
