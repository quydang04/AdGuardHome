import React from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Trans, useTranslation } from 'react-i18next';

import { Checkbox } from '../ui/Controls/Checkbox';
import { Input } from '../ui/Controls/Input';
import { Textarea } from '../ui/Controls/Textarea';
import { toNumber } from '../../helpers/form';

import '../Settings/FormButton.css';

const MAX_THRESHOLD = 100;
const MIN_THRESHOLD = 0;
const MIN_MINUTES = 1;
const MAX_MINUTES = 24 * 60;

export type FormValues = {
    enabled: boolean;
    botToken: string;
    chatId: string;
    cpuThreshold: number;
    memoryThreshold: number;
    diskThreshold: number;
    checkInterval: number;
    cooldown: number;
    customMessage: string;
};

type Props = {
    initialValues: FormValues;
    processing: boolean;
    processingTest: boolean;
    onSubmit: (values: FormValues) => void;
    onTest: () => void;
};

const defaultValues: FormValues = {
    enabled: false,
    botToken: '',
    chatId: '',
    cpuThreshold: 90,
    memoryThreshold: 90,
    diskThreshold: 90,
    checkInterval: 1,
    cooldown: 5,
    customMessage: '',
};

export const Form = ({ initialValues, processing, processingTest, onSubmit, onTest }: Props) => {
    const { t } = useTranslation();

    const {
        handleSubmit,
        control,
        watch,
        formState: { isSubmitting },
    } = useForm<FormValues>({
        defaultValues: {
            ...defaultValues,
            ...initialValues,
        },
        mode: 'onBlur',
    });

    const enabled = watch('enabled');

    const disableSubmit = isSubmitting || processing;

    const handleTestClick = () => {
        if (!processingTest) {
            onTest();
        }
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <div className="form__group form__group--settings">
                <Controller
                    name="enabled"
                    control={control}
                    render={({ field }) => (
                        <Checkbox
                            {...field}
                            data-testid="telegram_enabled"
                            title={t('telegram_enable')}
                            disabled={processing}
                        />
                    )}
                />
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="botToken"
                    control={control}
                    rules={{
                        validate: (value) => {
                            if (enabled && !value.trim()) {
                                return t('telegram_bot_token_required');
                            }

                            return true;
                        },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            data-testid="telegram_bot_token"
                            label={t('telegram_bot_token')}
                            disabled={processing}
                            error={fieldState.error?.message}
                            trimOnBlur
                        />
                    )}
                />
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="chatId"
                    control={control}
                    rules={{
                        validate: (value) => {
                            if (enabled && !value.trim()) {
                                return t('telegram_chat_id_required');
                            }

                            return true;
                        },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            data-testid="telegram_chat_id"
                            label={t('telegram_chat_id')}
                            disabled={processing}
                            error={fieldState.error?.message}
                            trimOnBlur
                        />
                    )}
                />
            </div>

            <div className="form__label form__label--with-desc">
                <Trans>telegram_thresholds_title</Trans>
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="cpuThreshold"
                    control={control}
                    rules={{
                        min: { value: MIN_THRESHOLD, message: t('telegram_threshold_min_max') },
                        max: { value: MAX_THRESHOLD, message: t('telegram_threshold_min_max') },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            type="number"
                            label={t('telegram_threshold_cpu')}
                            min={MIN_THRESHOLD}
                            max={MAX_THRESHOLD}
                            disabled={processing}
                            error={fieldState.error?.message}
                            onChange={(e) => field.onChange(toNumber(e.target.value))}
                        />
                    )}
                />

                <Controller
                    name="memoryThreshold"
                    control={control}
                    rules={{
                        min: { value: MIN_THRESHOLD, message: t('telegram_threshold_min_max') },
                        max: { value: MAX_THRESHOLD, message: t('telegram_threshold_min_max') },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            type="number"
                            label={t('telegram_threshold_memory')}
                            min={MIN_THRESHOLD}
                            max={MAX_THRESHOLD}
                            disabled={processing}
                            error={fieldState.error?.message}
                            onChange={(e) => field.onChange(toNumber(e.target.value))}
                        />
                    )}
                />

                <Controller
                    name="diskThreshold"
                    control={control}
                    rules={{
                        min: { value: MIN_THRESHOLD, message: t('telegram_threshold_min_max') },
                        max: { value: MAX_THRESHOLD, message: t('telegram_threshold_min_max') },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            type="number"
                            label={t('telegram_threshold_disk')}
                            min={MIN_THRESHOLD}
                            max={MAX_THRESHOLD}
                            disabled={processing}
                            error={fieldState.error?.message}
                            onChange={(e) => field.onChange(toNumber(e.target.value))}
                        />
                    )}
                />
            </div>

            <div className="form__label form__label--with-desc">
                <Trans>telegram_schedule_title</Trans>
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="checkInterval"
                    control={control}
                    rules={{
                        min: { value: MIN_MINUTES, message: t('telegram_interval_min_max') },
                        max: { value: MAX_MINUTES, message: t('telegram_interval_min_max') },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            type="number"
                            label={t('telegram_check_interval')}
                            min={MIN_MINUTES}
                            max={MAX_MINUTES}
                            disabled={processing}
                            error={fieldState.error?.message}
                            onChange={(e) => field.onChange(toNumber(e.target.value))}
                        />
                    )}
                />

                <Controller
                    name="cooldown"
                    control={control}
                    rules={{
                        min: { value: MIN_MINUTES, message: t('telegram_interval_min_max') },
                        max: { value: MAX_MINUTES, message: t('telegram_interval_min_max') },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            type="number"
                            label={t('telegram_cooldown')}
                            min={MIN_MINUTES}
                            max={MAX_MINUTES}
                            disabled={processing}
                            error={fieldState.error?.message}
                            onChange={(e) => field.onChange(toNumber(e.target.value))}
                        />
                    )}
                />
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="customMessage"
                    control={control}
                    render={({ field }) => (
                        <Textarea
                            {...field}
                            data-testid="telegram_custom_message"
                            placeholder={t('telegram_custom_message_placeholder')}
                            label={t('telegram_custom_message')}
                            disabled={processing}
                            trimOnBlur
                        />
                    )}
                />

                <div className="form__desc form__desc--top">
                    <Trans>telegram_custom_message_hint</Trans>
                </div>
            </div>

            <div className="mt-5 d-flex flex-wrap gap-2">
                <button
                    type="submit"
                    className="btn btn-success btn-standard btn-large"
                    data-testid="telegram_save"
                    disabled={disableSubmit}
                >
                    <Trans>save_btn</Trans>
                </button>

                <button
                    type="button"
                    className="btn btn-outline-secondary btn-standard"
                    data-testid="telegram_test"
                    onClick={handleTestClick}
                    disabled={processing || processingTest}
                >
                    <Trans>telegram_test_button</Trans>
                </button>
            </div>
        </form>
    );
};
