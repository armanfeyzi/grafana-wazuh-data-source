import {
  createDataFrame,
  CustomVariableSupport,
  DataQueryRequest,
  DataQueryResponse,
  FieldType,
  MetricFindValue,
} from '@grafana/data';
import { from, Observable } from 'rxjs';
import { map } from 'rxjs/operators';

import { VariableQueryEditor } from './components/VariableQueryEditor';
import { DataSource } from './datasource';
import { WazuhVariableQuery } from './types';

function toVariableFrame(values: MetricFindValue[]) {
  return createDataFrame({
    fields: [
      {
        name: 'text',
        type: FieldType.string,
        values: values.map((v) => v.text),
      },
      {
        name: 'value',
        type: FieldType.string,
        values: values.map((v) => `${v.value ?? v.text}`),
      },
    ],
  });
}

export class WazuhVariableSupport extends CustomVariableSupport<DataSource, WazuhVariableQuery> {
  editor = VariableQueryEditor;

  constructor(private readonly datasource: DataSource) {
    super();
  }

  query(request: DataQueryRequest<WazuhVariableQuery>): Observable<DataQueryResponse> {
    return from(this.datasource.metricFindQuery()).pipe(
      map((values) => ({
        data: [toVariableFrame(values)],
      }))
    );
  }
}
