import { useState } from 'react'

const LIMITE_CARACTERES = 60

export default function GoalSprint({ texto }: { texto: string }) {
  const [expandido, setExpandido] = useState(false)
  const esLargo = [...texto].length > LIMITE_CARACTERES

  if (!esLargo) {
    return <span>{texto}</span>
  }

  const visible = expandido ? texto : `${[...texto].slice(0, LIMITE_CARACTERES).join('').trimEnd()}…`

  return (
    <div className="flex max-w-md flex-col gap-1">
      <span className="break-words">{visible}</span>
      <button
        type="button"
        onClick={() => setExpandido((anterior) => !anterior)}
        aria-expanded={expandido}
        className="self-start text-xs font-semibold text-primary hover:underline focus-visible:shadow-[var(--ring)]"
      >
        {expandido ? 'Ver menos' : 'Ver más'}
      </button>
    </div>
  )
}
