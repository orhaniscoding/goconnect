import fs from 'node:fs'
import { execSync } from 'node:child_process'
const tpl='docs/README.tpl.md', out='README.md'
let t='# GoConnect\n'
try{ t=fs.readFileSync(tpl,'utf8') }catch{}
let tag='v0.0.0'
try{ tag=execSync('git describe --tags --abbrev=0').toString().trim() }catch{}
const outText=t.replaceAll('{{LATEST_TAG}}',tag).replaceAll('{{RELEASE_DATE}}',new Date().toISOString().slice(0,10))
fs.writeFileSync(out,outText); console.log('README.md updated')
