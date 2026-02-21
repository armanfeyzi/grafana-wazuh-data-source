import { useEffect } from 'react';
import { QueryEditorProps } from '@grafana/data';

import { DataSource } from '../datasource';
import { WazuhDataSourceOptions, WazuhQuery, WazuhVariableQuery } from '../types';

type Props = QueryEditorProps<DataSource, WazuhQuery, WazuhDataSourceOptions, WazuhVariableQuery>;

export function VariableQueryEditor({ query, onChange }: Props) {
  useEffect(() => {
    if (query.query !== 'agents') {
      onChange({ ...query, query: 'agents', refId: query.refId || 'A' });
    }
  }, [query, onChange]);

  return null;
}
