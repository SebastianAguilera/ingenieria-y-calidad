const CABECERAS_NO_REENVIADAS = new Set([
  'connection',
  'content-length',
  'host',
  'keep-alive',
  'proxy-authenticate',
  'proxy-authorization',
  'te',
  'trailer',
  'transfer-encoding',
  'upgrade',
])

// Queda por debajo del maxDuration de 120 s configurado en vercel.json (D10).
const TIEMPO_MAXIMO_MS = 110_000

export interface SolicitudProxy extends AsyncIterable<Buffer | Uint8Array | string> {
  method?: string
  url?: string
  headers: Record<string, string | string[] | undefined>
}

export interface RespuestaProxy {
  statusCode: number
  setHeader(nombre: string, valor: string): void
  end(cuerpo?: Buffer | string): void
}

// handler reenvia /api/* a la API de Render agregando la clave compartida
// FRONTEND_KEY (D9). La clave vive solo en el entorno del servidor de Vercel,
// nunca en el bundle del navegador.
export default async function handler(req: SolicitudProxy, res: RespuestaProxy): Promise<void> {
  const origen = process.env.API_ORIGIN
  const clave = process.env.FRONTEND_KEY
  if (!origen || !clave) {
    responderError(res, 500, 'PROXY_MAL_CONFIGURADO', 'faltan API_ORIGIN o FRONTEND_KEY')
    return
  }

  const cuerpo = await leerCuerpo(req)

  let respuesta: Response
  try {
    respuesta = await fetch(construirDestino(origen, req.url), {
      method: req.method ?? 'GET',
      headers: construirCabeceras(req.headers, clave),
      body: cuerpo,
      signal: AbortSignal.timeout(TIEMPO_MAXIMO_MS),
    })
  } catch (error) {
    if (error instanceof Error && error.name === 'TimeoutError') {
      responderError(res, 504, 'PROXY_TIMEOUT', 'la API no respondio a tiempo')
      return
    }
    responderError(res, 502, 'PROXY_ERROR', 'no se pudo contactar a la API')
    return
  }

  res.statusCode = respuesta.status
  respuesta.headers.forEach((valor, nombre) => {
    if (!CABECERAS_NO_REENVIADAS.has(nombre.toLowerCase())) {
      res.setHeader(nombre, valor)
    }
  })
  res.end(Buffer.from(await respuesta.arrayBuffer()))
}

async function leerCuerpo(req: SolicitudProxy): Promise<Buffer | undefined> {
  if (req.method === 'GET' || req.method === 'HEAD') {
    return undefined
  }

  const partes: Buffer[] = []
  for await (const parte of req) {
    partes.push(Buffer.isBuffer(parte) ? parte : Buffer.from(parte))
  }
  return partes.length > 0 ? Buffer.concat(partes) : undefined
}

function construirDestino(origen: string, ruta: string | undefined): string {
  const base = origen.replace(/\/+$/, '')
  if (!ruta) {
    return base
  }
  return ruta.startsWith('/') ? `${base}${ruta}` : `${base}/${ruta}`
}

function construirCabeceras(
  entrantes: Record<string, string | string[] | undefined>,
  clave: string,
): Record<string, string> {
  const salida: Record<string, string> = {}
  for (const [nombre, valor] of Object.entries(entrantes)) {
    if (valor === undefined) {
      continue
    }
    const minuscula = nombre.toLowerCase()
    if (CABECERAS_NO_REENVIADAS.has(minuscula) || minuscula === 'x-frontend-key') {
      continue
    }
    salida[nombre] = Array.isArray(valor) ? valor.join(', ') : valor
  }
  salida['X-Frontend-Key'] = clave
  return salida
}

function responderError(res: RespuestaProxy, estado: number, codigo: string, mensaje: string): void {
  res.statusCode = estado
  res.setHeader('Content-Type', 'application/json')
  res.end(JSON.stringify({ error: codigo, mensaje }))
}
