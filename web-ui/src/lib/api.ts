export async function api(path: string, init?: RequestInit): Promise<any> {
  const base = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'
  const res = await fetch(base + path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  })
  return res.json()
}
