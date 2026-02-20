import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { WazuhQuery, WazuhDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<WazuhQuery, WazuhDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<WazuhDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<WazuhQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: WazuhQuery): boolean {
    return Boolean(query.dataType);
  }
}
