import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { AgentOption, WazuhQuery, WazuhDataSourceOptions, DEFAULT_QUERY } from './types';
import { isQueryRunnable, normalizeQuery } from './queryUtils';

export class DataSource extends DataSourceWithBackend<WazuhQuery, WazuhDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<WazuhDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<WazuhQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: WazuhQuery): boolean {
    return isQueryRunnable(normalizeQuery(query));
  }

  async getAgents(): Promise<AgentOption[]> {
    return this.getResource<AgentOption[]>('agents');
  }
}
