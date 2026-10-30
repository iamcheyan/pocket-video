import test from 'node:test';
import assert from 'node:assert/strict';
import { isPageName, parseRoute, routeToHash } from '../frontend/src/router.ts';

test('parseRoute understands every page', () => {
  assert.deepEqual(parseRoute(''), { page: 'home' });
  assert.deepEqual(parseRoute('#home'), { page: 'home' });
  assert.deepEqual(parseRoute('#/favorites'), { page: 'favorites' });
  assert.deepEqual(parseRoute('#queue'), { page: 'queue' });
  assert.deepEqual(parseRoute('#settings/'), { page: 'settings' });
  assert.deepEqual(parseRoute('#videos'), { page: 'videos' });
});

test('parseRoute carries ids for deep links', () => {
  assert.deepEqual(parseRoute('#detail/BV1DEMO002'), { page: 'detail', videoId: 'BV1DEMO002' });
  assert.deepEqual(parseRoute('#detail/BV1%20X'), { page: 'detail', videoId: 'BV1 X' });
  assert.deepEqual(parseRoute('#favorites/fav-audio'), { page: 'favorites', favoriteId: 'fav-audio' });
  assert.deepEqual(parseRoute('#detail'), { page: 'videos' });
});

test('unknown routes fall back to home instead of crashing', () => {
  assert.deepEqual(parseRoute('#nope'), { page: 'home' });
  assert.deepEqual(parseRoute('#nope/thing'), { page: 'home' });
});

test('routeToHash round-trips through parseRoute', () => {
  for (const hash of ['#home', '#favorites', '#favorites/fav-ui', '#videos', '#detail/BV1DEMO001', '#queue', '#settings']) {
    assert.equal(routeToHash(parseRoute(hash)), hash);
  }
});

test('isPageName rejects non-pages', () => {
  assert.equal(isPageName('settings'), true);
  assert.equal(isPageName('detail'), true);
  assert.equal(isPageName('nope'), false);
});
