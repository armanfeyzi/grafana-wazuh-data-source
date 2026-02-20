import { DataSourcePlugin } from '@grafana/data';

import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { DataSource } from './datasource';
import { WazuhQuery, WazuhDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, WazuhQuery, WazuhDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
