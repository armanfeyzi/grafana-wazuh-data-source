import React from 'react';
import { InlineField, Select, Stack } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { WazuhDataSourceOptions, WazuhDataType, WazuhQuery, WazuhQueryFormat } from '../types';

type Props = QueryEditorProps<DataSource, WazuhQuery, WazuhDataSourceOptions>;

const dataTypeOptions: Array<SelectableValue<WazuhDataType>> = [
  { label: 'Alerts', value: 'alerts' },
  { label: 'Vulnerabilities', value: 'vulnerabilities' },
  { label: 'FIM', value: 'fim' },
  { label: 'SCA', value: 'sca' },
  { label: 'Agent status', value: 'agents' },
];

const formatOptions: Array<SelectableValue<WazuhQueryFormat>> = [
  { label: 'Time series', value: 'time_series' },
  { label: 'Table', value: 'table' },
  { label: 'Stat', value: 'stat' },
];

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onDataTypeChange = (option: SelectableValue<WazuhDataType>) => {
    if (!option.value) {
      return;
    }
    onChange({ ...query, dataType: option.value });
    onRunQuery();
  };

  const onFormatChange = (option: SelectableValue<WazuhQueryFormat>) => {
    if (!option.value) {
      return;
    }
    onChange({ ...query, format: option.value });
    onRunQuery();
  };

  return (
    <Stack gap={0}>
      <InlineField label="Data type" labelWidth={14}>
        <Select
          inputId="query-editor-data-type"
          options={dataTypeOptions}
          value={dataTypeOptions.find((o) => o.value === query.dataType)}
          onChange={onDataTypeChange}
          width={24}
        />
      </InlineField>
      <InlineField label="Format" labelWidth={14}>
        <Select
          inputId="query-editor-format"
          options={formatOptions}
          value={formatOptions.find((o) => o.value === query.format)}
          onChange={onFormatChange}
          width={24}
        />
      </InlineField>
    </Stack>
  );
}
