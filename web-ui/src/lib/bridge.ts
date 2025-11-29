import { toast } from 'sonner'

// Configuration
const DAEMON_PORTS = [12345, 12346, 12347] // Ports to scan
const MAX_RETRIES = 3

export interface DaemonStatus {
  running: boolean
  version: string
  paused: boolean
  device?: {
    registered: boolean
    public_key: string
    device_id: string
  }
  wg?: {
    active: boolean
    error?: string
    peers?: number
    total_rx?: number
    total_tx?: number
    last_handshake?: string
  }
}

class BridgeClient {
  private port: number | null = null
  private status: DaemonStatus | null = null

  constructor() {
    // Try to find port on init? Or lazy load?
    // Let's keep it simple and scan on demand or first use.
  }

  /**
   * Scans local ports to find the running daemon.
   */
  async findDaemon(): Promise<number | null> {
    if (this.port) return this.port

    for (const port of DAEMON_PORTS) {
      try {
        const controller = new AbortController()
        const timeoutId = setTimeout(() => controller.abort(), 500) // Fast scan

        const res = await fetch(`http://127.0.0.1:${port}/status`, {
          signal: controller.signal
        }).catch(() => null)

        clearTimeout(timeoutId)

        if (res && res.ok) {
          this.port = port
          return port
        }
      } catch (e) {
        // Ignore errors during scan
      }
    }
    return null
  }

  /**
   * Makes a request to the daemon bridge.
   */
  async request(path: string, init?: RequestInit): Promise<any> {
    const port = await this.findDaemon()
    if (!port) {
      throw new Error('Daemon not found')
    }

    const url = `http://127.0.0.1:${port}${path}`
    try {
      const res = await fetch(url, {
        ...init,
        headers: {
          'Content-Type': 'application/json',
          ...(init?.headers || {}),
        },
      })

      if (!res.ok) {
        throw new Error(`Daemon error: ${res.status}`)
      }

      return res.json()
    } catch (e) {
      this.port = null // Reset port on error to force re-scan next time
      throw e
    }
  }

  /**
   * Checks if daemon is running and returns status.
   */
  async getStatus(): Promise<DaemonStatus | null> {
    try {
      const status = await this.request('/status')
      this.status = status
      return status
    } catch (e) {
      this.status = null
      return null
    }
  }

  /**
   * Registers the device with the daemon using an auth token.
   */
  async register(token: string): Promise<void> {
    await this.request('/register', {
      method: 'POST',
      body: JSON.stringify({ token }),
    })
  }

  /**
   * Connects the VPN.
   */
  async connect(): Promise<void> {
    await this.request('/connect', { method: 'POST' })
  }

  /**
   * Disconnects the VPN.
   */
  async disconnect(): Promise<void> {
    await this.request('/disconnect', { method: 'POST' })
  }
}

export const bridge = new BridgeClient()