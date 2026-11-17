import React, { Component } from 'react';
import { withTranslation } from 'react-i18next';

import Card from '../../ui/Card';
import { confirmDialog } from '../../ui/ConfirmModal';

import { Form, FormValues } from './Form';
import { HOUR } from '../../../helpers/constants';

interface LogsConfigProps {
    interval: number;
    customInterval?: number;
    enabled: boolean;
    anonymize_client_ip: boolean;
    processing: boolean;
    ignored: unknown[];
    processingClear: boolean;
    setLogsConfig: (...args: unknown[]) => unknown;
    clearLogs: (...args: unknown[]) => unknown;
    t: (...args: unknown[]) => string;
}

class LogsConfig extends Component<LogsConfigProps> {
    handleFormSubmit = async (values: FormValues) => {
        const { t, interval: prevInterval } = this.props;
        const { interval, customInterval, ...rest } = values;

        const newInterval = customInterval ? customInterval * HOUR : interval;

        const data = {
            ...rest,
            ignored: values.ignored ? values.ignored.split('\n') : [],
            interval: newInterval,
        };

        if (newInterval < prevInterval) {
            const confirmed = await confirmDialog({
                title: t('confirm_action', 'Confirm'),
                message: t('query_log_retention_confirm'),
                confirmText: t('save_btn'),
                variant: 'danger',
            });
            if (confirmed) {
                this.props.setLogsConfig(data);
            }
        } else {
            this.props.setLogsConfig(data);
        }
    };

    handleClear = async () => {
        const { t, clearLogs } = this.props;
        const confirmed = await confirmDialog({
            title: t('query_log_clear'),
            message: t('query_log_confirm_clear'),
            confirmText: t('query_log_clear'),
            variant: 'danger',
        });
        if (confirmed) {
            clearLogs();
        }
    };

    render() {
        const {
            t,
            enabled,
            interval,
            processing,
            processingClear,
            anonymize_client_ip,
            ignored,
            customInterval,
        } = this.props;

        return (
            <Card title={t('query_log_configuration')} bodyType="card-body box-body--settings" id="logs-config">
                <div className="form">
                    <Form
                        initialValues={{
                            enabled,
                            interval,
                            customInterval,
                            anonymize_client_ip,
                            ignored: ignored?.join('\n'),
                        }}
                        processing={processing}
                        processingReset={processingClear}
                        onSubmit={this.handleFormSubmit}
                        onReset={this.handleClear}
                    />
                </div>
            </Card>
        );
    }
}

export default withTranslation()(LogsConfig);
