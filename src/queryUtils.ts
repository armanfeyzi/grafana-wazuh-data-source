import { WazuhDataType, WazuhQuery, WazuhQueryFormat } from './types';

export const IMPLEMENTED_DATA_TYPES: WazuhDataType[] = [
  'alerts',
  'vulnerabilities',
  'fim',
  'sca',
  'agents',
];

export const DATA_TYPE_LABELS: Record<WazuhDataType, string> = {
  alerts: 'Alerts',
  vulnerabilities: 'Vulnerabilities',
  fim: 'FIM',
  sca: 'SCA',
  agents: 'Agent status',
};

export const FORMAT_LABELS: Record<WazuhQueryFormat, string> = {
  time_series: 'Time series',
  table: 'Table',
  stat: 'Stat',
};

const ALERT_FORMATS: WazuhQueryFormat[] = ['time_series', 'table', 'stat'];
const VULNERABILITY_FORMATS: WazuhQueryFormat[] = ['table', 'stat', 'time_series'];
const FIM_FORMATS: WazuhQueryFormat[] = ['time_series', 'table', 'stat'];
const SCA_FORMATS: WazuhQueryFormat[] = ['table', 'time_series', 'stat'];
const AGENT_FORMATS: WazuhQueryFormat[] = ['table'];

export const FORMAT_HINTS: Partial<Record<WazuhDataType, Partial<Record<WazuhQueryFormat, string>>>> = {
  sca: {
    table: 'Current compliance scores (manager API)',
    time_series: 'Historical scan activity (indexer alerts)',
    stat: 'Total SCA scan events in range',
  },
  vulnerabilities: {
    table: 'Affected packages and CVEs',
    stat: 'Total open vulnerabilities',
    time_series: 'Detections over time',
  },
  fim: {
    table: 'Recent file integrity changes',
    time_series: 'FIM events over time',
    stat: 'Total FIM events in range',
  },
};

export function isDataTypeImplemented(dataType: WazuhDataType): boolean {
  return IMPLEMENTED_DATA_TYPES.includes(dataType);
}

export function formatsForDataType(dataType: WazuhDataType): WazuhQueryFormat[] {
  switch (dataType) {
    case 'alerts':
      return ALERT_FORMATS;
    case 'vulnerabilities':
      return VULNERABILITY_FORMATS;
    case 'fim':
      return FIM_FORMATS;
    case 'sca':
      return SCA_FORMATS;
    case 'agents':
      return AGENT_FORMATS;
    default:
      return [];
  }
}

export function defaultFormatForDataType(dataType: WazuhDataType): WazuhQueryFormat {
  switch (dataType) {
    case 'agents':
      return 'table';
    case 'vulnerabilities':
      return 'table';
    case 'fim':
      return 'time_series';
    case 'sca':
      return 'table';
    case 'alerts':
    default:
      return 'time_series';
  }
}

export function normalizeQuery(query: WazuhQuery): WazuhQuery {
  const dataType = query.dataType || 'alerts';
  const allowedFormats = formatsForDataType(dataType);
  let format = query.format || defaultFormatForDataType(dataType);

  if (!allowedFormats.includes(format)) {
    format = defaultFormatForDataType(dataType);
  }

  return {
    ...query,
    dataType,
    format,
    limit: query.limit ?? 100,
    filters: query.filters ?? {},
  };
}

export function validateQuery(query: WazuhQuery): string | undefined {
  if (!query.dataType) {
    return 'Data type is required';
  }

  if (!isDataTypeImplemented(query.dataType)) {
    return `${DATA_TYPE_LABELS[query.dataType]} is not available yet`;
  }

  const allowedFormats = formatsForDataType(query.dataType);
  if (query.format && !allowedFormats.includes(query.format)) {
    return `Format "${FORMAT_LABELS[query.format]}" is not supported for ${DATA_TYPE_LABELS[query.dataType]}`;
  }

  const min = query.filters?.ruleLevelMin;
  const max = query.filters?.ruleLevelMax;
  if (min != null && max != null && min > max) {
    return 'Minimum rule level cannot be greater than maximum rule level';
  }

  if (min != null && (min < 0 || min > 15)) {
    return 'Rule level must be between 0 and 15';
  }

  if (max != null && (max < 0 || max > 15)) {
    return 'Rule level must be between 0 and 15';
  }

  if (query.limit != null && query.limit < 1) {
    return 'Row limit must be at least 1';
  }

  return undefined;
}

export function isQueryRunnable(query: WazuhQuery): boolean {
  return validateQuery(query) === undefined;
}

export function showLimitField(dataType: WazuhDataType, format: WazuhQueryFormat): boolean {
  return format === 'table' && dataType !== 'agents';
}

export function showAlertFilters(dataType: WazuhDataType): boolean {
  return dataType === 'alerts';
}

export function showAgentFilter(dataType: WazuhDataType): boolean {
  return ['alerts', 'vulnerabilities', 'fim', 'sca'].includes(dataType);
}

export function showSeverityFilter(dataType: WazuhDataType): boolean {
  return dataType === 'vulnerabilities';
}

export const VULNERABILITY_SEVERITIES = ['Critical', 'High', 'Medium', 'Low', 'None'];

export function formatHint(dataType: WazuhDataType, format: WazuhQueryFormat): string | undefined {
  return FORMAT_HINTS[dataType]?.[format];
}
