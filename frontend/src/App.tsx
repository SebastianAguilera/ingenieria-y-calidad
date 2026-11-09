import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import RutaProtegida from './components/RutaProtegida'
import AppShell from './layouts/AppShell'
import ProyectoLayout from './layouts/ProyectoLayout'
import LoginPage from './pages/LoginPage'
import NotFoundPage from './pages/NotFoundPage'
import ProyectosPage from './pages/ProyectosPage'
import UsuariosPage from './pages/UsuariosPage'
import BacklogPage from './pages/proyecto/BacklogPage'
import DefectosPage from './pages/proyecto/DefectosPage'
import EquipoPage from './pages/proyecto/EquipoPage'
import MetricasPage from './pages/proyecto/MetricasPage'
import ResumenPage from './pages/proyecto/ResumenPage'
import SprintsPage from './pages/proyecto/SprintsPage'
import { SesionProvider } from './session/SesionContext'
import { ThemeProvider } from './theme/ThemeProvider'

export function RutasPanel() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<RutaProtegida />}>
        <Route element={<AppShell />}>
          <Route index element={<Navigate to="/proyectos" replace />} />
          <Route path="proyectos" element={<ProyectosPage />} />
          <Route path="usuarios" element={<UsuariosPage />} />
          <Route path="proyectos/:id" element={<ProyectoLayout />}>
            <Route index element={<ResumenPage />} />
            <Route path="backlog" element={<BacklogPage />} />
            <Route path="sprints" element={<SprintsPage />} />
            <Route path="equipo" element={<EquipoPage />} />
            <Route path="defectos" element={<DefectosPage />} />
            <Route path="metricas" element={<MetricasPage />} />
          </Route>
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Route>
    </Routes>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <SesionProvider>
        <BrowserRouter>
          <RutasPanel />
        </BrowserRouter>
      </SesionProvider>
    </ThemeProvider>
  )
}
