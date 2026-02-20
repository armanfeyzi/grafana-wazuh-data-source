import React, { ChangeEvent, useEffect, useMemo, useState } from 'react';
import { Alert, InlineField, Input, MultiSelect, Select, Stack } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import {
  DATA_TYPE_LABELS,
  FORMAT_LABELS,
  VULNERABILITY_SEVERITIES,
  formatHint,
  formatsForDataType,
  isQueryRunnable,
  normalizeQuery,
  showAgentFilter,
  showAlertFilters,
  showLimitField,
  showSeverityFilter,
  validateQuery,
} from '../queryUtils';
import { WazuhDataSourceOptions, WazuhDataType, WazuhQuery, WazuhQueryFormat } from '../types';

type Props = QueryEditorProps<DataSource, WazuhQuery, WazuhDataSourceOptions>;

const ALL_DATA_TYPES: WazuhDataType[] = ['alerts', 'vulnerabilities', 'fim', 'sca', 'agents'];

function parseOptionalInt(value: string): number | undefined {
  if (value.trim() === '') {
    return undefined;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const normalized = normalizeQuery(query);
  const validationError = validateQuery(normalized);
  const [agentOptions, setAgentOptions] = useState<Array<SelectableValue<string>>>([]);

  const dataTypeOptions = useMemo(
    () =>
      ALL_DATA_TYPES.map((value) => ({
        label: DATA_TYPE_LABELS[value],
        value,
      })),
    []
  );

  const formatOptions = useMemo(
    () =>
      formatsForDataType(normalized.dataType).map((value) => ({
        label: FORMAT_LABELS[value],
        value,
        description: formatHint(normalized.dataType, value),
      })),
    [normalized.dataType]
  );

  const selectedFormatHint = formatHint(normalized.dataType, normalized.format);

  useEffect(() => {
    if (!showAgentFilter(normalized.dataType)) {
      return;
    }

    let cancelled = false;
    datasource
      .getAgents()
      .then((agents) => {
        if (cancelled) {
          return;
        }
        setAgentOptions(agents.map((agent) => ({ label: agent.label, value: agent.value })));
      })
      .catch(() => {
        if (!cancelled) {
          setAgentOptions([]);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [datasource, normalized.dataType]);

  const applyQuery = (next: WazuhQuery, run = true) => {
    const updated = normalizeQuery(next);
    onChange(updated);
    if (run && isQueryRunnable(updated)) {
      onRunQuery();
    }
  };

  const onDataTypeChange = (option: SelectableValue<WazuhDataType>) => {
    if (!option.value) {
      return;
    }
    applyQuery({ ...normalized, dataType: option.value });
  };

  const onFormatChange = (option: SelectableValue<WazuhQueryFormat>) => {
    if (!option.value) {
      return;
    }
    applyQuery({ ...normalized, format: option.value });
  };

  const onLimitChange = (event: ChangeEvent<HTMLInputElement>) => {
    applyQuery({ ...normalized, limit: parseOptionalInt(event.target.value) ?? 100 }, false);
  };

  const onAgentChange = (options: Array<SelectableValue<string>>) => {
    applyQuery({
      ...normalized,
      filters: {
        ...normalized.filters,
        agentNames: options.map((option) => option.value).filter((value): value is string => Boolean(value)),
      },
    });
  };

  const onSeverityChange = (options: Array<SelectableValue<string>>) => {
    applyQuery({
      ...normalized,
      filters: {
        ...normalized.filters,
        severity: options.map((option) => option.value).filter((value): value is string => Boolean(value)),
      },
    });
  };

  const onRuleGroupsChange = (groups: string[]) => {
    applyQuery({
      ...normalized,
      filters: {
        ...normalized.filters,
        ruleGroups: groups.length > 0 ? groups : undefined,
      },
    });
  };

  const onRuleLevelMinChange = (event: ChangeEvent<HTMLInputElement>) => {
    applyQuery(
      {
        ...normalized,
        filters: {
          ...normalized.filters,
          ruleLevelMin: parseOptionalInt(event.target.value),
        },
      },
      false
    );
  };

  const onRuleLevelMaxChange = (event: ChangeEvent<HTMLInputElement>) => {
    applyQuery(
      {
        ...normalized,
        filters: {
          ...normalized.filters,
          ruleLevelMax: parseOptionalInt(event.target.value),
        },
      },
      false
    );
  };

  const selectedAgents = (normalized.filters?.agentNames ?? []).map((name) => {
    const match = agentOptions.find((option) => option.value === name);
    return match ?? { label: name, value: name };
  });

  const severityOptions = VULNERABILITY_SEVERITIES.map((severity) => ({
    label: severity,
    value: severity,
  }));

  const selectedSeverities = (normalized.filters?.severity ?? []).map((severity) => ({
    label: severity,
    value: severity,
  }));

  return (
    <Stack gap={1} direction="column">
      {validationError && (
        <Alert title="Query validation" severity="warning">
          {validationError}
        </Alert>
      )}

      <Stack gap={0}>
        <InlineField label="Data type" labelWidth={14} tooltip="Wazuh dataset to query">
          <Select
            inputId="query-editor-data-type"
            options={dataTypeOptions}
            value={dataTypeOptions.find((option) => option.value === normalized.dataType)}
            onChange={onDataTypeChange}
            width={24}
          />
        </InlineField>

        {formatOptions.length > 0 && (
          <InlineField
            label="Format"
            labelWidth={14}
            tooltip={selectedFormatHint ?? 'Choose how results are shaped for the panel'}
          >
            <Select
              inputId="query-editor-format"
              options={formatOptions}
              value={formatOptions.find((option) => option.value === normalized.format)}
              onChange={onFormatChange}
              width={24}
              isDisabled={formatOptions.length === 1}
            />
          </InlineField>
        )}

        {showLimitField(normalized.dataType, normalized.format) && (
          <InlineField label="Limit" labelWidth={14} tooltip="Maximum number of table rows">
            <Input
              type="number"
              width={10}
              min={1}
              max={1000}
              value={normalized.limit ?? 100}
              onChange={onLimitChange}
              onBlur={() => isQueryRunnable(normalized) && onRunQuery()}
            />
          </InlineField>
        )}
      </Stack>

      {showAgentFilter(normalized.dataType) && !showAlertFilters(normalized.dataType) && (
        <Stack gap={0}>
          <InlineField label="Agents" labelWidth={14} tooltip="Filter by agent name. Leave empty for all agents.">
            <MultiSelect
              inputId="query-editor-agents-filter"
              options={agentOptions}
              value={selectedAgents}
              onChange={onAgentChange}
              width={40}
              placeholder="All agents"
              allowCustomValue
            />
          </InlineField>

          {showSeverityFilter(normalized.dataType) && (
            <InlineField label="Severity" labelWidth={14} tooltip="Filter vulnerabilities by severity">
              <MultiSelect
                inputId="query-editor-severity"
                options={severityOptions}
                value={selectedSeverities}
                onChange={onSeverityChange}
                width={24}
                placeholder="All severities"
              />
            </InlineField>
          )}
        </Stack>
      )}

      {showAlertFilters(normalized.dataType) && (
        <Stack gap={0}>
          <InlineField label="Agents" labelWidth={14} tooltip="Filter alerts by agent name. Leave empty for all agents.">
            <MultiSelect
              inputId="query-editor-agents"
              options={agentOptions}
              value={selectedAgents}
              onChange={onAgentChange}
              width={40}
              placeholder="All agents"
              allowCustomValue
            />
          </InlineField>

          <InlineField label="Rule level min" labelWidth={14} tooltip="Wazuh rule level 0–15">
            <Input
              type="number"
              width={8}
              min={0}
              max={15}
              placeholder="Any"
              value={normalized.filters?.ruleLevelMin ?? ''}
              onChange={onRuleLevelMinChange}
              onBlur={() => isQueryRunnable(normalized) && onRunQuery()}
            />
          </InlineField>

          <InlineField label="Rule level max" labelWidth={14} tooltip="Wazuh rule level 0–15">
            <Input
              type="number"
              width={8}
              min={0}
              max={15}
              placeholder="Any"
              value={normalized.filters?.ruleLevelMax ?? ''}
              onChange={onRuleLevelMaxChange}
              onBlur={() => isQueryRunnable(normalized) && onRunQuery()}
            />
          </InlineField>

          <InlineField label="Rule groups" labelWidth={14} tooltip="Filter by Wazuh rule groups, e.g. sshd, syscheck">
            <Input
              width={40}
              placeholder="sshd, authentication (comma-separated)"
              defaultValue={(normalized.filters?.ruleGroups ?? []).join(', ')}
              key={(normalized.filters?.ruleGroups ?? []).join('|')}
              onBlur={(event) =>
                onRuleGroupsChange(
                  event.currentTarget.value
                    .split(',')
                    .map((group) => group.trim())
                    .filter(Boolean)
                )
              }
            />
          </InlineField>
        </Stack>
      )}
    </Stack>
  );
}
