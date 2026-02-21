import React from 'react';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { InlineField, Select } from '@grafana/ui';

import { DataSource } from '../datasource';
import { WazuhDataSourceOptions, WazuhQuery, WazuhVariableQuery, WazuhVariableQueryType } from '../types';

type Props = QueryEditorProps<DataSource, WazuhQuery, WazuhDataSourceOptions, WazuhVariableQuery>;

const VARIABLE_QUERY_OPTIONS: Array<SelectableValue<WazuhVariableQueryType>> = [
  {
    label: 'Agents',
    value: 'agents',
    description: 'List of registered Wazuh agents — use as $agent variable',
  },
  {
    label: 'Namespaces',
    value: 'namespaces',
    description: 'Distinct Kubernetes namespaces from Wazuh alert data — use as $namespace variable',
  },
];

export function VariableQueryEditor({ query, onChange }: Props) {
  const currentType: WazuhVariableQueryType = query.query ?? 'agents';

  const handleChange = (option: SelectableValue<WazuhVariableQueryType>) => {
    if (option.value) {
      onChange({ ...query, query: option.value, refId: query.refId || 'A' });
    }
  };

  return (
    <InlineField
      label="Variable type"
      labelWidth={16}
      tooltip="Choose what this template variable lists"
    >
      <Select
        inputId="variable-query-editor-type"
        options={VARIABLE_QUERY_OPTIONS}
        value={VARIABLE_QUERY_OPTIONS.find((o) => o.value === currentType)}
        onChange={handleChange}
        width={30}
      />
    </InlineField>
  );
}
