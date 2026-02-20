import React, { ChangeEvent } from 'react';
import { Checkbox, InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { WazuhDataSourceOptions, WazuhSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<WazuhDataSourceOptions, WazuhSecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onJsonChange = (field: keyof WazuhDataSourceOptions) => (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        [field]: event.target.value,
      },
    });
  };

  const onTlsSkipVerifyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        tlsSkipVerify: event.target.checked,
      },
    });
  };

  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        password: event.target.value,
      },
    });
  };

  const onResetPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        password: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        password: '',
      },
    });
  };

  return (
    <>
      <InlineField label="Manager URL" labelWidth={14} interactive tooltip="Wazuh manager API base URL">
        <Input
          id="config-editor-manager-url"
          onChange={onJsonChange('managerUrl')}
          value={jsonData.managerUrl || ''}
          placeholder="https://wazuh.example.com:55000"
          width={50}
        />
      </InlineField>
      <InlineField label="Indexer URL" labelWidth={14} interactive tooltip="Wazuh indexer (OpenSearch) base URL">
        <Input
          id="config-editor-indexer-url"
          onChange={onJsonChange('indexerUrl')}
          value={jsonData.indexerUrl || ''}
          placeholder="https://indexer.example.com:9200"
          width={50}
        />
      </InlineField>
      <InlineField label="Username" labelWidth={14} interactive>
        <Input
          id="config-editor-username"
          onChange={onJsonChange('username')}
          value={jsonData.username || ''}
          placeholder="wazuh-wui"
          width={30}
        />
      </InlineField>
      <InlineField label="Password" labelWidth={14} interactive tooltip="Stored securely; sent to the backend only">
        <SecretInput
          required
          id="config-editor-password"
          isConfigured={secureJsonFields.password}
          value={secureJsonData?.password}
          placeholder="Password"
          width={30}
          onReset={onResetPassword}
          onChange={onPasswordChange}
        />
      </InlineField>
      <InlineField label="Skip TLS verify" labelWidth={14} interactive>
        <Checkbox id="config-editor-tls-skip" value={jsonData.tlsSkipVerify} onChange={onTlsSkipVerifyChange} />
      </InlineField>
    </>
  );
}
