import React, { Component } from 'react';
import { withTranslation } from 'react-i18next';

import Card from '../../ui/Card';
import { confirmDialog } from '../../ui/ConfirmModal';

import { Form, FormValues } from './Form';
import { HOUR } from '../../../helpers/constants';

interface StatsConfigProps {
    interval: number;
    customInterval?: number;
    ignored: unknown[];
    enabled: boolean;
    processing: boolean;
    processingReset: boolean;
    setStatsConfig: (...args: unknown[]) => unknown;
    resetStats: (...args: unknown[]) => unknown;
    t: (...args: unknown[]) => string;
}

class StatsConfig extends Component<StatsConfigProps> {
    handleFormSubmit = async ({ enabled, interval, ignored, customInterval }: FormValues) => {
        const { t, interval: prevInterval } = this.props;
        const newInterval = customInterval ? customInterval * HOUR : interval;

        const config = {
            enabled,
            interval: newInterval,
            ignored: ignored ? ignored.split('\n') : [],
        };

        if (config.interval < prevInterval) {
            const confirmed = await confirmDialog({
                title: t('confirm_action', 'Confirm'),
                message: t('statistics_retention_confirm'),
                confirmText: t('save_btn'),
                variant: 'danger',
            });
            if (confirmed) {
                this.props.setStatsConfig(config);
            }
        } else {
            this.props.setStatsConfig(config);
        }
    };

    handleReset = async () => {
        const { t, resetStats } = this.props;
        const confirmed = await confirmDialog({
            title: t('statistics_clear'),
            message: t('statistics_clear_confirm'),
            confirmText: t('statistics_clear'),
            variant: 'danger',
        });
        if (confirmed) {
            resetStats();
        }
    };

    render() {
        const {
            t,
            interval,
            customInterval,
            processing,
            processingReset,
            ignored,
            enabled,
        } = this.props;

        return (
            <Card title={t('statistics_configuration')} bodyType="card-body box-body--settings" id="stats-config">
                <div className="form">
                    <Form
                        initialValues={{
                            interval,
                            customInterval,
                            enabled,
                            ignored: ignored?.join('\n'),
                        }}
                        processing={processing}
                        processingReset={processingReset}
                        onSubmit={this.handleFormSubmit}
                        onReset={this.handleReset}
                    />
                </div>
            </Card>
        );
    }
}

export default withTranslation()(StatsConfig);
