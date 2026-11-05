import { Link } from 'react-router-dom'
import PageHeader from '../components/PageHeader'

export default function NotFoundPage() {
  return (
    <section>
      <PageHeader title="404" description="La página que buscás no existe." />
      <Link
        to="/"
        className="inline-flex rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700"
      >
        Volver al inicio
      </Link>
    </section>
  )
}
