import fs from 'node:fs';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

const source = fs.readFileSync(new URL('../frontend/src/i18n.ts', import.meta.url), 'utf8');
const js = source
  .replace(/export type [\s\S]*?\n};?\n/g, '')
  .replace(/export const locales[^\n]*\n/g, '')
  .replace('export const copy: Record<Locale, Copy> =', 'const copy =')
  .replace('export const copy', 'const copy')
  + '\nexport { copy };\n';
const temp = path.join(fs.mkdtempSync(path.join(import.meta.dirname, '.i18n-')), 'dict.mjs');
fs.writeFileSync(temp, js);
const module = await import(pathToFileURL(temp).href);
fs.rmSync(path.dirname(temp), { recursive: true, force: true });

const locales = Object.keys(module.copy);
if (locales.length !== 3) throw new Error(`Expected 3 locales, found ${locales.length}`);
const baseline = Object.keys(module.copy.en).sort().join(',');
for (const locale of locales) {
  const keys = Object.keys(module.copy[locale]).sort().join(',');
  if (keys !== baseline) throw new Error(`${locale} translation keys differ from en`);
}
const empty = locales.flatMap(locale => Object.entries(module.copy[locale]).filter(([, value]) => !String(value).trim()).map(([key]) => `${locale}.${key}`));
if (empty.length) throw new Error(`Empty translations: ${empty.join(', ')}`);
console.log(`i18n keys complete: ${locales.map(locale => `${locale}=${Object.keys(module.copy[locale]).length}`).join(', ')}`);
