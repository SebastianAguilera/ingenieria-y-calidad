import { Link } from 'react-router-dom'
import PageHeader from '../components/PageHeader'

export default function NotFoundPage() {
  return (
    <section>
      <PageHeader title="404" description="La página que buscás no existe." />
      <Link
        to="/proyectos"
        className="inline-flex rounded-pill bg-primary px-5 py-2.5 text-sm font-semibold text-[var(--on-primary)] transition hover:bg-[var(--primary-700)]"
      >
        Volver al panel
      </Link>
    </section>
  )
}
