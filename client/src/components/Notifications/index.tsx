import React, { useEffect, useMemo } from 'react';
import { withTranslation, Trans } from 'react-i18next';

import Card from '../ui/Card';
import Loading from '../ui/Loading';
import PageTitle from '../ui/PageTitle';

import { Form, FormValues } from './Form';
import { NotificationsState } from '../../initialState';

const MINUTES_TO_MS = 60 * 1000;

const toFormValues = (telegram: NotificationsState['telegram']): FormValues => ({
    enabled: telegram.enabled,
    botToken: telegram.bot_token,
    chatId: telegram.chat_id,
    cpuThreshold: telegram.cpu_threshold,
    memoryThreshold: telegram.memory_threshold,
    diskThreshold: telegram.disk_threshold,
    checkInterval: Math.max(1, Math.round(telegram.check_interval / MINUTES_TO_MS)),
    cooldown: Math.max(1, Math.round(telegram.cooldown / MINUTES_TO_MS)),
    customMessage: telegram.custom_message || '',
});

type Props = {
    notifications: NotificationsState;
    getTelegramConfig: () => void;
    setTelegramConfig: (config: any) => void;
    sendTelegramTest: (message: string) => void;
    t: (...args: unknown[]) => string;
};

const Notifications = ({ notifications, getTelegramConfig, setTelegramConfig, sendTelegramTest, t }: Props) => {
    useEffect(() => {
        getTelegramConfig();
    }, []);

    const formValues = useMemo(() => toFormValues(notifications.telegram), [notifications.telegram]);

    const handleSubmit = (values: FormValues) => {
        const payload = {
            enabled: values.enabled,
            bot_token: values.botToken.trim(),
            chat_id: values.chatId.trim(),
            cpu_threshold: Number(values.cpuThreshold),
            memory_threshold: Number(values.memoryThreshold),
            disk_threshold: Number(values.diskThreshold),
            check_interval: Number(values.checkInterval) * MINUTES_TO_MS,
            cooldown: Number(values.cooldown) * MINUTES_TO_MS,
            custom_message: values.customMessage.trim(),
        };

        setTelegramConfig(payload);
    };

    const handleTest = () => {
        sendTelegramTest(t('telegram_test_default_message'));
    };

    return (
        <>
            <PageTitle title={t('notifications_settings')} />

            {notifications.processingGet && <Loading />}

            {!notifications.processingGet && (
                <div className="content">
                    <div className="row">
                        <div className="col-md-12">
                            <Card
                                title={t('telegram_settings_title')}
                                subtitle={t('telegram_settings_subtitle')}
                                bodyType="card-body box-body--settings"
                            >
                                <div className="form">
                                    <Form
                                        initialValues={formValues}
                                        processing={notifications.processingSave}
                                        processingTest={notifications.processingTest}
                                        onSubmit={handleSubmit}
                                        onTest={handleTest}
                                    />

                                    {!notifications.telegram.enabled && (
                                        <div className="form__desc mt-3">
                                            <Trans>telegram_status_disabled</Trans>
                                        </div>
                                    )}
                                </div>
                            </Card>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
};

export default withTranslation()(Notifications);
