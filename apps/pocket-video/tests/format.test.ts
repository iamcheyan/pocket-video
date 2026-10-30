import test from 'node:test';
import assert from 'node:assert/strict';
import { formatBytes, formatCount, formatDuration, formatSyncTime } from '../frontend/src/format.ts';

test('formatDuration covers seconds, minutes, and hours', () => {
  assert.equal(formatDuration(0), '0:00');
  assert.equal(formatDuration(65), '1:05');
  assert.equal(formatDuration(305), '5:05');
  assert.equal(formatDuration(3600), '1:00:00');
  assert.equal(formatDuration(3661), '1:01:01');
  assert.equal(formatDuration(Number.NaN), '0:00');
  assert.equal(formatDuration(-10), '0:00');
});

test('formatBytes scales through binary units', () => {
  assert.equal(formatBytes(0), '0 B');
  assert.equal(formatBytes(1023), '1023 B');
  assert.equal(formatBytes(2048), '2.0 KB');
  assert.equal(formatBytes(5 * 1024 * 1024), '5.0 MB');
  assert.equal(formatBytes(20 * 1024 * 1024 * 1024), '20 GB');
});

test('formatCount replaces the placeholder once per occurrence', () => {
  assert.equal(formatCount('{count} videos', 4), '4 videos');
  assert.equal(formatCount('{count} / {count}', 3), '3 / 3');
  assert.equal(formatCount('videos', 3), 'videos');
});

test('formatSyncTime tolerates missing and invalid timestamps', () => {
  assert.equal(formatSyncTime(undefined, 'en'), '');
  assert.equal(formatSyncTime('not-a-date', 'en'), '');
  assert.match(formatSyncTime('2026-10-30T09:00:00Z', 'en'), /2026/);
});
