import React from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Trans, useTranslation } from 'react-i18next';

import { Checkbox } from '../../ui/Controls/Checkbox';
import { Input } from '../../ui/Controls/Input';
import { Textarea } from '../../ui/Controls/Textarea';
import { Select } from '../../ui/Controls/Select';
import { toNumber } from '../../../helpers/form';

const MIN_RENEW_BEFORE_DAYS = 1;
const MAX_RENEW_BEFORE_DAYS = 60;

export const CHALLENGE_HTTP01 = 'http-01';
export const CHALLENGE_CLOUDFLARE_DNS01 = 'dns-01-cloudflare';

export type AcmeFormValues = {
    enabled: boolean;
    email: string;
    domains: string;
    challenge: string;
    cloudflareApiToken: string;
    autoRenew: boolean;
    renewBeforeDays: number;
};

type Props = {
    initialValues: AcmeFormValues;
    processingConfig: boolean;
    processingIssue: boolean;
    lastIssuedAt: string | null;
    lastError: string;
    onSave: (values: AcmeFormValues) => void;
    onIssue: (values: AcmeFormValues) => void;
};

const defaultValues: AcmeFormValues = {
    enabled: false,
    email: '',
    domains: '',
    challenge: CHALLENGE_HTTP01,
    cloudflareApiToken: '',
    autoRenew: true,
    renewBeforeDays: 14,
};

export const AcmeForm = ({
    initialValues,
    processingConfig,
    processingIssue,
    lastIssuedAt,
    lastError,
    onSave,
    onIssue,
}: Props) => {
    const { t } = useTranslation();

    const {
        handleSubmit,
        control,
        watch,
        formState: { isSubmitting },
    } = useForm<AcmeFormValues>({
        defaultValues: {
            ...defaultValues,
            ...initialValues,
        },
        mode: 'onBlur',
    });

    const enabled = watch('enabled');
    const challenge = watch('challenge');

    const processing = processingConfig || processingIssue;
    const disableSubmit = isSubmitting || processing;

    return (
        <form onSubmit={handleSubmit(onSave)}>
            <div className="form__group form__group--settings">
                <Controller
                    name="enabled"
                    control={control}
                    render={({ field }) => (
                        <Checkbox
                            {...field}
                            data-testid="acme_enabled"
                            title={t('acme_enable')}
                            disabled={processing}
                        />
                    )}
                />
                <div className="form__desc">
                    <Trans>acme_enable_desc</Trans>
                </div>
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="email"
                    control={control}
                    rules={{
                        validate: (value) => {
                            if (enabled && !value.trim()) {
                                return t('acme_email_required');
                            }

                            return true;
                        },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            data-testid="acme_email"
                            label={t('acme_email')}
                            disabled={processing}
                            error={fieldState.error?.message}
                            trimOnBlur
                        />
                    )}
                />
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="domains"
                    control={control}
                    rules={{
                        validate: (value) => {
                            if (enabled && !value.trim()) {
                                return t('acme_domains_required');
                            }

                            return true;
                        },
                    }}
                    render={({ field }) => (
                        <Textarea
                            {...field}
                            data-testid="acme_domains"
                            placeholder={t('acme_domains_placeholder')}
                            label={t('acme_domains')}
                            disabled={processing}
                            trimOnBlur
                        />
                    )}
                />
                <div className="form__desc form__desc--top">
                    <Trans>acme_domains_desc</Trans>
                </div>
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="challenge"
                    control={control}
                    render={({ field }) => (
                        <Select {...field} data-testid="acme_challenge" label={t('acme_challenge')} disabled={processing}>
                            <option value={CHALLENGE_HTTP01}>{t('acme_challenge_http01')}</option>
                            <option value={CHALLENGE_CLOUDFLARE_DNS01}>{t('acme_challenge_cloudflare')}</option>
                        </Select>
                    )}
                />
                <div className="form__desc form__desc--top">
                    {challenge === CHALLENGE_HTTP01 ? (
                        <Trans>acme_challenge_http01_desc</Trans>
                    ) : (
                        <Trans>acme_challenge_cloudflare_desc</Trans>
                    )}
                </div>
            </div>

            {challenge === CHALLENGE_CLOUDFLARE_DNS01 && (
                <div className="form__group form__group--settings">
                    <Controller
                        name="cloudflareApiToken"
                        control={control}
                        rules={{
                            validate: (value) => {
                                if (enabled && challenge === CHALLENGE_CLOUDFLARE_DNS01 && !value.trim()) {
                                    return t('acme_cloudflare_token_required');
                                }

                                return true;
                            },
                        }}
                        render={({ field, fieldState }) => (
                            <Input
                                {...field}
                                type="password"
                                data-testid="acme_cloudflare_token"
                                label={t('acme_cloudflare_token')}
                                disabled={processing}
                                error={fieldState.error?.message}
                                trimOnBlur
                            />
                        )}
                    />
                </div>
            )}

            <div className="form__group form__group--settings">
                <Controller
                    name="autoRenew"
                    control={control}
                    render={({ field }) => (
                        <Checkbox
                            {...field}
                            data-testid="acme_auto_renew"
                            title={t('acme_auto_renew')}
                            disabled={processing}
                        />
                    )}
                />
                <div className="form__desc">
                    <Trans>acme_auto_renew_desc</Trans>
                </div>
            </div>

            <div className="form__group form__group--settings">
                <Controller
                    name="renewBeforeDays"
                    control={control}
                    rules={{
                        min: { value: MIN_RENEW_BEFORE_DAYS, message: t('acme_renew_before_days_min_max') },
                        max: { value: MAX_RENEW_BEFORE_DAYS, message: t('acme_renew_before_days_min_max') },
                    }}
                    render={({ field, fieldState }) => (
                        <Input
                            {...field}
                            type="number"
                            label={t('acme_renew_before_days')}
                            min={MIN_RENEW_BEFORE_DAYS}
                            max={MAX_RENEW_BEFORE_DAYS}
                            disabled={processing}
                            error={fieldState.error?.message}
                            onChange={(e) => field.onChange(toNumber(e.target.value))}
                        />
                    )}
                />
            </div>

            {(lastIssuedAt || lastError) && (
                <div className="form__group form__group--settings">
                    {lastIssuedAt && (
                        <div className="form__desc">
                            <Trans values={{ value: new Date(lastIssuedAt).toLocaleString() }}>
                                acme_last_issued_at
                            </Trans>
                        </div>
                    )}
                    {lastError && (
                        <div className="form__message form__message--error mt-1">{lastError}</div>
                    )}
                </div>
            )}

            <div className="mt-5 d-flex flex-wrap gap-2">
                <button
                    type="submit"
                    className="btn btn-success btn-standard btn-large"
                    data-testid="acme_save"
                    disabled={disableSubmit}
                >
                    <Trans>save_btn</Trans>
                </button>

                <button
                    type="button"
                    className="btn btn-outline-secondary btn-standard"
                    data-testid="acme_issue"
                    onClick={handleSubmit(onIssue)}
                    disabled={disableSubmit}
                >
                    <Trans>acme_issue_button</Trans>
                </button>
            </div>
        </form>
    );
};
