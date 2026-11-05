const API_BASE = '/api'

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`)

  if (!response.ok) {
    throw new Error(`Error ${response.status} al llamar ${path}`)
  }

  return response.json() as Promise<T>
}
