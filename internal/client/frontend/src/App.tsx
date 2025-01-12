import React from 'react'
import { HashRouter, Route, Routes } from 'react-router-dom'
import { CssBaseline, CssVarsProvider } from '@mui/joy'
import Home from './pages/Home'
import SignIn from './pages/SignIn'
import SignUp from './pages/SignUp'
import { AuthProvider } from './contexts/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'

function App() {
  return (
    <AuthProvider>
      <CssVarsProvider>
        <CssBaseline disableTransitionOnChange/>
        <HashRouter>
          <Routes>
            <Route element={<ProtectedRoute/>}>
              <Route path="/" element={<Home />} />1
            </Route>
            <Route path="/signin" element={<SignIn />} />
            <Route path="/signup" element={<SignUp />} />
          </Routes>
        </HashRouter>
      </CssVarsProvider>
    </AuthProvider>
  );
}

export default App;
