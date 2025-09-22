export async function bridge(path: string, init?: RequestInit): Promise<any> {
  const origin = process.env.NEXT_PUBLIC_BRIDGE_ORIGIN || 'http://127.0.0.1'
  const port = process.env.NEXT_PUBLIC_BRIDGE_PORT_HINT || '12000'
  const url = origin + ':' + port + path
  const res = await fetch(url, init)
  return res.json()
}
