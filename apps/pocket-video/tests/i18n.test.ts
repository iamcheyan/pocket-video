import test from 'node:test';
import assert from 'node:assert/strict';
import { detectLocale, isLocale, locales, localeNames, t } from '../frontend/src/i18n.ts';

test('all three locales expose exactly the same keys and no empty copy', () => {
  const baseline = Object.keys(t.en).sort();
  assert.ok(baseline.length >= 100, `expected a complete dictionary, got ${baseline.length} keys`);
  for (const locale of locales) {
    assert.deepEqual(Object.keys(t[locale]).sort(), baseline, `${locale} keys differ`);
    for (const [key, value] of Object.entries(t[locale])) {
      assert.equal(typeof value, 'string', `${locale}.${key} is not a string`);
      assert.ok(value.trim().length > 0, `${locale}.${key} is empty`);
    }
  }
  assert.deepEqual(Object.keys(localeNames).sort(), [...locales].sort());
});

test('interpolated strings keep their placeholder in every locale', () => {
  for (const locale of locales) {
    assert.match(t[locale].searchResults, /\{count\}/);
    assert.match(t[locale].collectionCount, /\{count\}/);
    assert.match(t[locale].lastSync, /\{time\}/);
  }
});

test('hero copy keeps the explicit line break', () => {
  for (const locale of locales) {
    assert.ok(t[locale].hero.includes('\n'), `${locale} hero needs two lines`);
  }
});

test('language negotiation picks the closest supported locale', () => {
  assert.equal(detectLocale(['zh-Hans-CN', 'en']), 'zh-CN');
  assert.equal(detectLocale(['ja-JP', 'en-US']), 'ja');
  assert.equal(detectLocale(['en-GB']), 'en');
  assert.equal(detectLocale(['fr-FR']), 'en');
  assert.equal(detectLocale([]), 'en');
  assert.equal(detectLocale(['de-DE', 'zh-TW']), 'zh-CN');
});

test('isLocale guards persisted and injected values', () => {
  assert.equal(isLocale('ja'), true);
  assert.equal(isLocale('zh-CN'), true);
  assert.equal(isLocale('zh-cn'), false);
  assert.equal(isLocale('fr'), false);
  assert.equal(isLocale(null), false);
  assert.equal(isLocale(3), false);
});
