import test from 'node:test';
import assert from 'node:assert/strict';
import { applyTheme, isThemeSetting, resolveTheme } from '../frontend/src/theme.ts';

test('system theme follows the device preference', () => {
  assert.equal(resolveTheme('system', true), 'dark');
  assert.equal(resolveTheme('system', false), 'light');
  assert.equal(resolveTheme('light', true), 'light');
  assert.equal(resolveTheme('dark', false), 'dark');
});

test('isThemeSetting rejects unknown values', () => {
  assert.equal(isThemeSetting('system'), true);
  assert.equal(isThemeSetting('light'), true);
  assert.equal(isThemeSetting('dark'), true);
  assert.equal(isThemeSetting('sepia'), false);
  assert.equal(isThemeSetting(null), false);
});

test('applyTheme writes dataset and color-scheme for CSS', () => {
  const root = { dataset: {} as DOMStringMap, style: {} as CSSStyleDeclaration } as unknown as HTMLElement;
  assert.equal(applyTheme('dark', root, false), 'dark');
  assert.equal(root.dataset.theme, 'dark');
  assert.equal(root.style.colorScheme, 'dark');
  assert.equal(applyTheme('system', root, false), 'light');
  assert.equal(root.dataset.theme, 'light');
});
