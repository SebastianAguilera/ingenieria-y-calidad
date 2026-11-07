import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import RutaProtegida from './components/RutaProtegida'
import AppShell from './layouts/AppShell'
import LoginPage from './pages/LoginPage'
import NotFoundPage from './pages/NotFoundPage'
import ProyectosPage from './pages/ProyectosPage'
import { SesionProvider } from './session/SesionContext'
import { ThemeProvider } from './theme/ThemeProvider'

export default function App() {
  return (
    <ThemeProvider>
      <SesionProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route element={<RutaProtegida />}>
              <Route element={<AppShell />}>
                <Route index element={<Navigate to="/proyectos" replace />} />
                <Route path="proyectos" element={<ProyectosPage />} />
                <Route path="*" element={<NotFoundPage />} />
              </Route>
            </Route>
          </Routes>
        </BrowserRouter>
      </SesionProvider>
    </ThemeProvider>
  )
}
