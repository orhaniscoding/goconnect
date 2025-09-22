'use client'
import tr from '../locales/tr/common.json'
import en from '../locales/en/common.json'

export default function Footer(){
  const lang = (typeof navigator!=='undefined' && navigator.language?.startsWith('en')) ? 'en' : 'tr'
  const d = lang==='en' ? en : tr
  return (
    <footer style={{padding:16,borderTop:'1px solid #eee',marginTop:24,fontSize:12,opacity:.8}}>
      {d['footer.brand']}
    </footer>
  )
}
