import { copyFile } from 'node:fs/promises';
import { resolve } from 'node:path';
const root = resolve(import.meta.dirname, '..');
const output = resolve(root, 'build/vite-web');
for (const name of ['nexora-logo-v17.png', 'login-banner-v17.png', 'nexora-poster-v17.png']) {
  await copyFile(resolve(root, 'web', name), resolve(output, name));
}
