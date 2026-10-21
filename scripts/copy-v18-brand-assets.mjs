import { copyFile } from 'node:fs/promises';
import { resolve } from 'node:path';
const root = resolve(import.meta.dirname, '..');
const output = resolve(root, 'build/vite-web');
for (const name of ['nexora-logo-v18.png', 'login-banner-v18.png', 'nexora-poster-v18.png']) {
  await copyFile(resolve(root, 'web', name), resolve(output, name));
}
