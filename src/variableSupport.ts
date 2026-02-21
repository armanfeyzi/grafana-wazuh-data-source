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
    // Forward the query type from the first target so metricFindQuery can
    // dispatch to the correct resource endpoint (agents or namespaces).
    const queryType = request.targets[0]?.query;

    return from(this.datasource.metricFindQuery(queryType)).pipe(
      map((values) => ({
        data: [toVariableFrame(values)],
      }))
    );
  }
}
