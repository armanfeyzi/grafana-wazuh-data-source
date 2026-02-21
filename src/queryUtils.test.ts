import {
  applyTemplateVariablesToQuery,
  defaultFormatForDataType,
  formatsForDataType,
  isDataTypeImplemented,
  normalizeQuery,
  validateQuery,
} from './queryUtils';
import { WazuhQuery } from './types';

describe('queryUtils', () => {
  it('returns alert formats for alerts data type', () => {
    expect(formatsForDataType('alerts')).toEqual(['time_series', 'table', 'stat']);
  });

  it('returns table-only format for agents', () => {
    expect(formatsForDataType('agents')).toEqual(['table']);
    expect(defaultFormatForDataType('agents')).toBe('table');
  });

  it('normalizes invalid format for agents', () => {
    const query = normalizeQuery({
      refId: 'A',
      dataType: 'agents',
      format: 'time_series',
    });

    expect(query.format).toBe('table');
  });

  it('rejects invalid rule level range', () => {
    const query: WazuhQuery = {
      refId: 'A',
      dataType: 'alerts',
      format: 'table',
      filters: { ruleLevelMin: 10, ruleLevelMax: 5 },
    };
    expect(validateQuery(query)).toMatch(/Minimum rule level/);
  });

  it('includes all data types as implemented', () => {
    expect(isDataTypeImplemented('fim')).toBe(true);
    expect(formatsForDataType('vulnerabilities')).toContain('table');
  });

  it('expands dashboard template variables in agent filters', () => {
    const query: WazuhQuery = {
      refId: 'A',
      dataType: 'vulnerabilities',
      format: 'stat',
      filters: {
        agentNames: ['$agent'],
        severity: ['$severity'],
      },
    };

    const unresolved = applyTemplateVariablesToQuery(query, (target) => target);
    expect(unresolved.filters?.agentNames).toBeUndefined();
    expect(unresolved.filters?.severity).toBeUndefined();

    const resolved = applyTemplateVariablesToQuery(query, (target) => {
      if (target === '$agent') {
        return 'fedora,wazuh.manager';
      }
      if (target === '$severity') {
        return '$__all';
      }
      return target;
    });

    expect(resolved.filters?.agentNames).toEqual(['fedora', 'wazuh.manager']);
    expect(resolved.filters?.severity).toBeUndefined();
  });
});
