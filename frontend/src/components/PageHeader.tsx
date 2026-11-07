type PageHeaderProps = {
  title: string
  description: string
}

export default function PageHeader({ title, description }: PageHeaderProps) {
  return (
    <header className="mb-8">
      <h1 className="text-3xl font-bold tracking-tight text-ink">{title}</h1>
      <p className="mt-2 text-ink-soft">{description}</p>
    </header>
  )
}
