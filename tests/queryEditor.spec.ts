import { test, expect } from '@grafana/plugin-e2e';

// Query editor tests require a live Wazuh datasource connection and are run
// against a real environment. They are skipped in automated CI.
test.skip('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getQueryEditorRow('A').getByText('Data type')).toBeVisible();
});

test.skip('should trigger new query when format is changed', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  const queryReq = panelEditPage.waitForQueryDataRequest();
  await panelEditPage.getQueryEditorRow('A').getByText('Format').locator('..').getByRole('combobox').click();
  await panelEditPage.getByRole('option', { name: 'Table' }).click();
  await expect(await queryReq).toBeTruthy();
});

test.skip('data query should succeed', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');
  await expect(panelEditPage.refreshPanel()).toBeOK();
});
