import { DataSourceInstanceSettings, CoreApp, MetricFindValue, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';

import { AgentOption, WazuhQuery, WazuhDataSourceOptions, WazuhVariableQueryType, DEFAULT_QUERY } from './types';
import { applyTemplateVariablesToQuery, isQueryRunnable, normalizeQuery } from './queryUtils';
import { WazuhVariableSupport } from './variableSupport';

export class DataSource extends DataSourceWithBackend<WazuhQuery, WazuhDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<WazuhDataSourceOptions>) {
    super(instanceSettings);
    this.variables = new WazuhVariableSupport(this);
  }

  getDefaultQuery(_: CoreApp): Partial<WazuhQuery> {
    return DEFAULT_QUERY;
  }

  filterQuery(query: WazuhQuery): boolean {
    return isQueryRunnable(normalizeQuery(query));
  }

  applyTemplateVariables(query: WazuhQuery, scopedVars: ScopedVars): WazuhQuery {
    const templateSrv = getTemplateSrv();
    return applyTemplateVariablesToQuery(query, (target) => templateSrv.replace(target, scopedVars));
  }

  async getAgents(): Promise<AgentOption[]> {
    return this.getResource<AgentOption[]>('agents');
  }

  async getNamespaces(): Promise<AgentOption[]> {
    return this.getResource<AgentOption[]>('namespaces');
  }

  async metricFindQuery(queryType?: WazuhVariableQueryType): Promise<MetricFindValue[]> {
    if (queryType === 'namespaces') {
      const namespaces = await this.getNamespaces();
      return namespaces.map((ns) => ({
        text: ns.label,
        value: ns.value,
      }));
    }

    // Default: 'agents' (backwards-compatible — existing dashboards with no
    // explicit query type continue to get the agent list).
    const agents = await this.getAgents();
    return agents.map((agent) => ({
      text: agent.label,
      value: agent.value,
    }));
  }
}
