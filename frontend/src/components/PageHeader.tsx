type PageHeaderProps = {
  title: string
  description: string
  className?: string
}

export default function PageHeader({ title, description, className = 'mb-8' }: PageHeaderProps) {
  return (
    <header className={className}>
      <h1 className="text-3xl font-bold tracking-tight text-ink">{title}</h1>
      <p className="mt-2 text-ink-soft">{description}</p>
    </header>
  )
}
