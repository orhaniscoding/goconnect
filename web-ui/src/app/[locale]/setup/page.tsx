import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useT } from '../lib/i18n-context'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '../components/ui/card'
import { getSetupStatus, persistConfig, type SetupStatus, type SetupConfig } from '../lib/api'
import { toast } from 'sonner'
import { Loader2, Check, Server, Shield, Database } from 'lucide-react'

export default function SetupPage() {
  const t = useT()
  const router = useRouter()
  const [status, setStatus] = useState<SetupStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [step, setStep] = useState(1)
  const [restarting, setRestarting] = useState(false)

  // Form State
  const [mode, setMode] = useState<'personal' | 'enterprise'>('personal')
  const [adminEmail, setAdminEmail] = useState('')
  const [adminPassword, setAdminPassword] = useState('')
  const [networkName, setNetworkName] = useState('')

  useEffect(() => {
    checkStatus()
  }, [])

  async function checkStatus() {
    try {
      const s = await getSetupStatus()
      setStatus(s)
      
      // If config is present and valid, redirect to login
      if (s.config_present && s.config_valid) {
        router.push('/login')
        return
      }
      
      setLoading(false)
    } catch (e) {
      console.error(e)
      // If 404, it means setup mode is disabled -> redirect to login
      router.push('/login') 
    }
  }

  async function handleFinish() {
    setSaving(true)
    
    // Prepare config object
    // In a real implementation, we would generate keys here or let the backend do it
    // For now, we send a minimal config and rely on backend defaults/generation
    const config: SetupConfig = {
      server: {
        host: '0.0.0.0',
        port: '8080',
        environment: 'production'
      },
      database: {
        backend: mode === 'personal' ? 'sqlite' : 'postgres',
        sqlite_path: 'data/goconnect.db',
        // Enterprise fields would be collected if mode === 'enterprise'
      },
      jwt: {
        // In a real app, these would be generated securely on the backend
        // or we'd request generation. For this simplified wizard, 
        // we assume the backend handles generation if empty/missing, 
        // OR we generate a random string here.
        secret: generateRandomString(32),
      },
      wireguard: {
        // Similarly, these should be auto-generated
        private_key: 'GENERATE_ME', // Backend should handle this
        server_endpoint: 'auto',
        server_pubkey: 'GENERATE_ME',
        interface_name: 'wg0',
        port: 51820
      }
    }

    try {
      const res = await persistConfig(config, true) // restart=true
      if (res.status === 'ok') {
        setRestarting(true)
        waitForRestart()
      } else {
        toast.error(t('setup.error_saving'))
        setSaving(false)
      }
    } catch (e) {
      toast.error(t('setup.error_generic'))
      setSaving(false)
    }
  }

  async function waitForRestart() {
    // Poll /health every 2 seconds
    const interval = setInterval(async () => {
      try {
        const res = await fetch('/health')
        if (res.ok) {
          clearInterval(interval)
          router.push('/login')
        }
      } catch (e) {
        // Server is down, keep polling
      }
    }, 2000)
  }

  function generateRandomString(length: number) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
    let result = ''
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length))
    }
    return result
  }

  if (loading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    )
  }

  if (restarting) {
    return (
      <div className="flex h-screen w-full flex-col items-center justify-center space-y-4">
        <Loader2 className="h-12 w-12 animate-spin text-primary" />
        <h2 className="text-xl font-semibold">{t('setup.restarting')}</h2>
        <p className="text-muted-foreground">{t('setup.restarting_desc')}</p>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen w-full items-center justify-center bg-muted/40 p-4">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>{t('setup.title')}</CardTitle>
          <CardDescription>{t('setup.subtitle')}</CardDescription>
        </CardHeader>
        <CardContent>
          {step === 1 && (
            <div className="space-y-4">
              <h3 className="font-medium">{t('setup.choose_mode')}</h3>
              <div 
                className={`cursor-pointer rounded-lg border p-4 transition-all hover:border-primary ${mode === 'personal' ? 'border-primary bg-primary/5' : ''}`}
                onClick={() => setMode('personal')}
              >
                <div className="flex items-center space-x-4">
                  <Database className="h-6 w-6 text-primary" />
                  <div>
                    <p className="font-medium">{t('setup.mode_personal')}</p>
                    <p className="text-sm text-muted-foreground">{t('setup.mode_personal_desc')}</p>
                  </div>
                  {mode === 'personal' && <Check className="ml-auto h-5 w-5 text-primary" />}
                </div>
              </div>
              
              <div 
                className={`cursor-pointer rounded-lg border p-4 transition-all hover:border-primary ${mode === 'enterprise' ? 'border-primary bg-primary/5' : ''}`}
                onClick={() => setMode('enterprise')}
              >
                <div className="flex items-center space-x-4">
                  <Server className="h-6 w-6 text-primary" />
                  <div>
                    <p className="font-medium">{t('setup.mode_enterprise')}</p>
                    <p className="text-sm text-muted-foreground">{t('setup.mode_enterprise_desc')}</p>
                  </div>
                  {mode === 'enterprise' && <Check className="ml-auto h-5 w-5 text-primary" />}
                </div>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <h3 className="font-medium">{t('setup.admin_account')}</h3>
              <div className="space-y-2">
                <Label htmlFor="email">{t('auth.email')}</Label>
                <Input 
                  id="email" 
                  type="email" 
                  value={adminEmail} 
                  onChange={(e) => setAdminEmail(e.target.value)} 
                  placeholder="admin@example.com" 
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">{t('auth.password')}</Label>
                <Input 
                  id="password" 
                  type="password" 
                  value={adminPassword} 
                  onChange={(e) => setAdminPassword(e.target.value)} 
                />
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <h3 className="font-medium">{t('setup.first_network')}</h3>
              <div className="space-y-2">
                <Label htmlFor="netname">{t('network.name')}</Label>
                <Input 
                  id="netname" 
                  value={networkName} 
                  onChange={(e) => setNetworkName(e.target.value)} 
                  placeholder="My Private Network" 
                />
              </div>
              <div className="rounded-md bg-blue-50 p-3 text-sm text-blue-700 dark:bg-blue-950 dark:text-blue-200">
                <Shield className="mr-2 inline-block h-4 w-4" />
                {t('setup.keys_auto_generated')}
              </div>
            </div>
          )}
        </CardContent>
        <CardFooter className="flex justify-between">
          <Button 
            variant="ghost" 
            onClick={() => setStep(s => Math.max(1, s - 1))}
            disabled={step === 1}
          >
            {t('common.back')}
          </Button>
          
          {step < 3 ? (
            <Button onClick={() => setStep(s => s + 1)}>
              {t('common.next')}
            </Button>
          ) : (
            <Button onClick={handleFinish} disabled={saving}>
              {saving && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {t('setup.finish')}
            </Button>
          )}
        </CardFooter>
      </Card>
    </div>
  )
}
