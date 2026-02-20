import {
  defaultFormatForDataType,
  formatsForDataType,
  isQueryRunnable,
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

  it('rejects unimplemented data types', () => {
    const query: WazuhQuery = { refId: 'A', dataType: 'fim', format: 'table' };
    expect(validateQuery(query)).toMatch(/not available yet/);
    expect(isQueryRunnable(query)).toBe(false);
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
});
