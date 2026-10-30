import test from 'node:test';
import assert from 'node:assert/strict';
import { ApiError, createApiClient } from '../frontend/src/api.ts';

type Call = { url: string; init: RequestInit | undefined };

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

function recorder(response: Response | (() => Response), calls: Call[] = []) {
  return (async (url: string, init?: RequestInit) => {
    calls.push({ url, init });
    return typeof response === 'function' ? response() : response;
  }) as typeof fetch;
}

test('client targets the injected base URL and encodes query filters', async () => {
  const calls: Call[] = [];
  const api = createApiClient({ base: 'http://device:9000/api/', fetchImpl: recorder(jsonResponse({ items: [], total: 0, downloaded: 0, limit: 50, offset: 0 }), calls) });
  const page = await api.videos({ q: '雨 音', favoriteId: 'fav-music', limit: 20 });
  assert.equal(page.total, 0);
  assert.equal(calls[0].url, 'http://device:9000/api/videos?q=%E9%9B%A8+%E9%9F%B3&favorite_id=fav-music&limit=20');
});

test('POST bodies are JSON with the right content type', async () => {
  const calls: Call[] = [];
  const api = createApiClient({ fetchImpl: recorder(jsonResponse({ id: 7, status: 'queued' }), calls) });
  await api.download('BV1DEMO001', 'audio');
  const [call] = calls;
  assert.equal(call.url, '/api/download');
  assert.equal(call.init?.method, 'POST');
  assert.equal((call.init?.headers as Record<string, string>)['Content-Type'], 'application/json');
  assert.deepEqual(JSON.parse(String(call.init?.body)), { video_id: 'BV1DEMO001', kind: 'audio' });
});

test('error envelopes become typed ApiError instances', async () => {
  const api = createApiClient({
    fetchImpl: recorder(jsonResponse({ error: { code: 'invalid_state', message: 'task is not running' } }, 409)),
  });
  await assert.rejects(() => api.pause(3), (error: unknown) => {
    assert.ok(error instanceof ApiError);
    assert.equal(error.code, 'invalid_state');
    assert.equal(error.status, 409);
    assert.equal(error.message, 'task is not running');
    return true;
  });
});

test('network failures never leak a raw fetch error', async () => {
  const api = createApiClient({
    fetchImpl: (async () => {
      throw new TypeError('failed to fetch');
    }) as typeof fetch,
  });
  await assert.rejects(() => api.settings(), (error: unknown) => {
    assert.ok(error instanceof ApiError);
    assert.equal(error.code, 'network');
    return true;
  });
});

test('204 responses resolve without parsing a body', async () => {
  const api = createApiClient({ fetchImpl: recorder(new Response(null, { status: 204 })) });
  assert.equal(await api.removeTask(12), undefined);
});

test('media URLs keep the audio discriminator', () => {
  const api = createApiClient({ base: '/api' });
  assert.equal(api.mediaUrl('BV1DEMO001'), '/api/media/BV1DEMO001');
  assert.equal(api.mediaUrl('BV1DEMO001', 'audio'), '/api/media/BV1DEMO001?kind=audio');
});

test('non-JSON error pages still produce a useful message', async () => {
  const api = createApiClient({ fetchImpl: recorder(new Response('bad gateway', { status: 502 })) });
  await assert.rejects(() => api.health(), (error: unknown) => {
    assert.ok(error instanceof ApiError);
    assert.equal(error.code, 'http_error');
    assert.match(error.message, /502/);
    return true;
  });
});
