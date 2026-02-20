import { test, expect } from '@grafana/plugin-e2e';
import { WazuhDataSourceOptions, WazuhSecureJsonData } from '../src/types';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByLabel('Manager URL')).toBeVisible();
});

test('"Save & test" reports connectivity errors when Wazuh is unreachable', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<WazuhDataSourceOptions, WazuhSecureJsonData>({
    fileName: 'datasources.yml',
  });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Manager URL' }).fill(ds.jsonData.managerUrl ?? '');
  await page.getByRole('textbox', { name: 'Indexer URL' }).fill(ds.jsonData.indexerUrl ?? '');
  await page.getByRole('textbox', { name: 'Username' }).fill(ds.jsonData.username ?? '');
  await page.getByRole('textbox', { name: 'Password' }).fill(ds.secureJsonData?.password ?? '');
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: /manager API|indexer/i });
});

test('"Save & test" should fail when configuration is invalid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<WazuhDataSourceOptions, WazuhSecureJsonData>({
    fileName: 'datasources.yml',
  });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Manager URL' }).fill(ds.jsonData.managerUrl ?? '');
  await page.getByRole('textbox', { name: 'Indexer URL' }).fill(ds.jsonData.indexerUrl ?? '');
  await page.getByRole('textbox', { name: 'Username' }).fill(ds.jsonData.username ?? '');
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: 'Password is required' });
});
