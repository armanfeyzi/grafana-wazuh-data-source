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
        ...secureJsonData,
        password: event.target.value,
      },
    });
  };

  const onIndexerPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...secureJsonData,
        indexerPassword: event.target.value,
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
        ...secureJsonData,
        password: '',
      },
    });
  };

  const onResetIndexerPassword = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        indexerPassword: false,
      },
      secureJsonData: {
        ...secureJsonData,
        indexerPassword: '',
      },
    });
  };

  return (
    <>
      <InlineField label="Manager URL" labelWidth={16} interactive tooltip="Wazuh manager API base URL">
        <Input
          id="config-editor-manager-url"
          onChange={onJsonChange('managerUrl')}
          value={jsonData.managerUrl || ''}
          placeholder="https://host.containers.internal:55000"
          width={50}
        />
      </InlineField>
      <InlineField label="Indexer URL" labelWidth={16} interactive tooltip="Wazuh indexer (OpenSearch) base URL">
        <Input
          id="config-editor-indexer-url"
          onChange={onJsonChange('indexerUrl')}
          value={jsonData.indexerUrl || ''}
          placeholder="https://host.containers.internal:9200"
          width={50}
        />
      </InlineField>
      <InlineField
        label="API username"
        labelWidth={16}
        interactive
        tooltip="Wazuh manager API user (e.g. wazuh-wui)"
      >
        <Input
          id="config-editor-username"
          onChange={onJsonChange('username')}
          value={jsonData.username || ''}
          placeholder="wazuh-wui"
          width={30}
        />
      </InlineField>
      <InlineField label="API password" labelWidth={16} interactive tooltip="Stored securely; sent to the backend only">
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
      <InlineField
        label="Indexer username"
        labelWidth={16}
        interactive
        tooltip="OpenSearch user; leave empty to reuse API username"
      >
        <Input
          id="config-editor-indexer-username"
          onChange={onJsonChange('indexerUsername')}
          value={jsonData.indexerUsername || ''}
          placeholder="admin"
          width={30}
        />
      </InlineField>
      <InlineField
        label="Indexer password"
        labelWidth={16}
        interactive
        tooltip="Leave empty to reuse API password"
      >
        <SecretInput
          id="config-editor-indexer-password"
          isConfigured={secureJsonFields.indexerPassword}
          value={secureJsonData?.indexerPassword}
          placeholder="Optional"
          width={30}
          onReset={onResetIndexerPassword}
          onChange={onIndexerPasswordChange}
        />
      </InlineField>
      <InlineField label="Skip TLS verify" labelWidth={16} interactive>
        <Checkbox id="config-editor-tls-skip" value={jsonData.tlsSkipVerify} onChange={onTlsSkipVerifyChange} />
      </InlineField>
    </>
  );
}
