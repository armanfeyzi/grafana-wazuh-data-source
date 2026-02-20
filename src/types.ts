import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export type WazuhDataType = 'alerts' | 'vulnerabilities' | 'fim' | 'sca' | 'agents';

export type WazuhQueryFormat = 'time_series' | 'table' | 'stat';

export interface WazuhQueryFilters {
  agentNames?: string[];
  ruleLevelMin?: number;
  ruleLevelMax?: number;
  severity?: string[];
}

export interface WazuhQuery extends DataQuery {
  dataType: WazuhDataType;
  format: WazuhQueryFormat;
  aggregation?: string;
  groupBy?: string;
  filters?: WazuhQueryFilters;
  limit?: number;
}

export const DEFAULT_QUERY: Partial<WazuhQuery> = {
  dataType: 'alerts',
  format: 'time_series',
  limit: 100,
};

export interface WazuhDataSourceOptions extends DataSourceJsonData {
  managerUrl?: string;
  indexerUrl?: string;
  username?: string;
  tlsSkipVerify?: boolean;
  indexPrefix?: string;
}

export interface WazuhSecureJsonData {
  password?: string;
}
