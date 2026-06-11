/**
 * Capture catalog screenshots for plugin.json.
 *
 * Usage:
 *   GRAFANA_URL=http://grafana.example.com \
 *   GRAFANA_USER=admin GRAFANA_PASSWORD=secret \
 *   node scripts/capture-catalog-screenshots.mjs
 */
import { chromium } from '@playwright/test';
import { mkdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const outDir = path.resolve(__dirname, '../src/img');

const baseURL = process.env.GRAFANA_URL ?? 'http://localhost:3000';
const username = process.env.GRAFANA_USER ?? 'admin';
const password = process.env.GRAFANA_PASSWORD ?? 'admin';

const shots = [
  {
    file: 'screenshot-security-overview.png',
    path: '/d/wazuh-security-overview/wazuh-security-overview?orgId=1&from=now-1h&to=now&timezone=browser&var-agent=$__all&refresh=30s',
    width: 1600,
    height: 900,
  },
  {
    file: 'screenshot-datasource-config.png',
    path: '/connections/datasources/edit/wazuh',
    width: 1400,
    height: 900,
  },
  {
    file: 'screenshot-explore-alerts.png',
    path: '/explore?orgId=1&left=%7B%22datasource%22%3A%22wazuh%22%2C%22queries%22%3A%5B%7B%22refId%22%3A%22A%22%2C%22datasource%22%3A%7B%22type%22%3A%22armanfeyzi-wazuh-datasource%22%2C%22uid%22%3A%22wazuh%22%7D%2C%22dataType%22%3A%22alerts%22%2C%22format%22%3A%22table%22%7D%5D%2C%22range%22%3A%7B%22from%22%3A%22now-1h%22%2C%22to%22%3A%22now%22%7D%7D',
    width: 1600,
    height: 900,
  },
];

async function login(page) {
  await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded', timeout: 120000 });
  if (!page.url().includes('/login')) {
    return;
  }
  await page.getByTestId('data-testid Username input field').fill(username);
  await page.getByTestId('data-testid Password input field').fill(password);
  await page.getByTestId('data-testid Login button').click();
  await page.waitForURL((url) => !url.pathname.endsWith('/login'), { timeout: 60000 });
}

async function main() {
  await mkdir(outDir, { recursive: true });

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1600, height: 900 },
    deviceScaleFactor: 1,
  });
  const page = await context.newPage();

  try {
    await login(page);

    for (const shot of shots) {
      console.log(`Capturing ${shot.file} ...`);
      await page.setViewportSize({ width: shot.width, height: shot.height });
      await page.goto(`${baseURL}${shot.path}`, { waitUntil: 'domcontentloaded', timeout: 120000 });

      if (shot.file.includes('security-overview')) {
        await page.waitForTimeout(10000);
      } else if (shot.file.includes('explore')) {
        await page.getByRole('button', { name: /Run query|Refresh/i }).click({ timeout: 15000 }).catch(() => {});
        await page.waitForTimeout(8000);
      } else {
        await page.waitForTimeout(5000);
      }

      await page.evaluate(() => {
        document
          .querySelectorAll('[role="alert"][class*="error"], [data-testid="data-testid Alert error"]')
          .forEach((el) => el.remove());
      });
      await page.waitForTimeout(500);
      const target = path.join(outDir, shot.file);
      await page.screenshot({ path: target, fullPage: false });
      console.log(`  → ${target}`);
    }
  } finally {
    await browser.close();
  }

  console.log('Done. Rebuild the plugin so screenshots are copied into dist/.');
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
