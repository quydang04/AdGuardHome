import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';
import cn from 'clsx';

import Card from '../../../ui/Card';
import { Checkbox } from '../../../ui/Controls/Checkbox';
import { setDnsConfig } from '../../../../actions/dnsConfig';
import { BLOCKING_MODES } from '../../../../helpers/constants';
import { RootState } from '../../../../initialState';
import '../../FormButton.css';
import './BlockPage.css';

const PREV_MODE_STORAGE_KEY = 'adgh_prev_blocking_mode';

const PREVIEW_SCENARIOS = [
    { key: 'ads', labelKey: 'block_page_scenario_ads' },
    { key: 'malware', labelKey: 'block_page_scenario_malware' },
    { key: 'service', labelKey: 'block_page_scenario_service' },
    { key: 'parental', labelKey: 'block_page_scenario_parental' },
    { key: 'custom', labelKey: 'block_page_scenario_custom' },
] as const;

const readPrevBlockingMode = (): string => {
    try {
        const stored = window.localStorage.getItem(PREV_MODE_STORAGE_KEY);
        if (
            stored &&
            stored !== BLOCKING_MODES.block_page &&
            Object.values(BLOCKING_MODES).includes(stored)
        ) {
            return stored;
        }
    } catch {
        // Ignore storage errors in restricted browser contexts.
    }

    return BLOCKING_MODES.default;
};

const normalizeDomainInput = (raw: string): string => {
    const trimmed = raw.trim();
    if (!trimmed) {
        return '';
    }

    try {
        const parsed = trimmed.includes('://') ? new URL(trimmed) : new URL(`http://${trimmed}`);

        return parsed.hostname.replace(/\.+$/, '').toLowerCase();
    } catch {
        return trimmed
            .replace(/^https?:\/\//i, '')
            .split('/')[0]
            .split(':')[0]
            .replace(/\.+$/, '')
            .toLowerCase();
    }
};

const BlockPage = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();
    const { blocking_mode, blocking_ipv4, blocking_ipv6, processingSetConfig } = useSelector(
        (state: RootState) => ({
            blocking_mode: state.dnsConfig.blocking_mode,
            blocking_ipv4: state.dnsConfig.blocking_ipv4,
            blocking_ipv6: state.dnsConfig.blocking_ipv6,
            processingSetConfig: state.dnsConfig.processingSetConfig,
        }),
        shallowEqual,
    );

    const [domainQuery, setDomainQuery] = useState<string>('');
    const isEnabled = blocking_mode === BLOCKING_MODES.block_page;

    useEffect(() => {
        if (blocking_mode && blocking_mode !== BLOCKING_MODES.block_page) {
            try {
                window.localStorage.setItem(PREV_MODE_STORAGE_KEY, blocking_mode);
            } catch {
                // Ignore storage errors.
            }
        }
    }, [blocking_mode]);

    const handleToggle = (nextEnabled: boolean) => {
        if (nextEnabled) {
            dispatch(setDnsConfig({ blocking_mode: BLOCKING_MODES.block_page }));

            return;
        }

        const prevMode = readPrevBlockingMode();
        if (
            prevMode === BLOCKING_MODES.custom_ip &&
            blocking_ipv4 &&
            blocking_ipv4 !== '0.0.0.0' &&
            blocking_ipv6
        ) {
            dispatch(
                setDnsConfig({
                    blocking_mode: BLOCKING_MODES.custom_ip,
                    blocking_ipv4,
                    blocking_ipv6,
                }),
            );

            return;
        }

        dispatch(
            setDnsConfig({
                blocking_mode: prevMode === BLOCKING_MODES.custom_ip ? BLOCKING_MODES.default : prevMode,
            }),
        );
    };

    const handleCheckSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const cleanDomain = normalizeDomainInput(domainQuery);
        if (!cleanDomain) {
            return;
        }

        window.open(`/block-page-preview?domain=${encodeURIComponent(cleanDomain)}`, '_blank', 'noopener,noreferrer');
    };

    const statusBadge = (
        <span
            className={cn('block-page__status', {
                'block-page__status--active': isEnabled,
            })}
            data-testid="block_page_status">
            <span className="block-page__status-dot" aria-hidden="true" />
            {isEnabled ? t('block_page_status_on') : t('block_page_status_off')}
        </span>
    );

    return (
        <Card
            title={t('block_page_title')}
            subtitle={t('block_page_subtitle')}
            headerExtra={statusBadge}
            bodyType="card-body box-body--settings"
            id="block-page">
            <div className="row">
                <div className="col-12 col-md-9">
                    <div className="form__group form__group--settings">
                        <Checkbox
                            name="block_page_enabled"
                            data-testid="block_page_toggle"
                            value={isEnabled}
                            title={t('block_page_enable')}
                            subtitle={t('block_page_desc')}
                            disabled={processingSetConfig}
                            onChange={handleToggle}
                        />
                    </div>

                    <div className="block-page__note">{t('block_page_http_note')}</div>
                </div>
            </div>

            <div className="block-page__divider" />

            <div className="mb-4">
                <span className="block-page__section-label">{t('block_page_preview_label')}</span>
                <div className="block-page__scenarios">
                    {PREVIEW_SCENARIOS.map(({ key, labelKey }) => (
                        <a
                            key={key}
                            href={`/block-page-preview?scenario=${key}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="btn btn-sm btn-outline-secondary">
                            {t(labelKey)}
                        </a>
                    ))}
                </div>
            </div>

            <form onSubmit={handleCheckSubmit}>
                <label className="block-page__section-label" htmlFor="block-page-domain-check">
                    {t('block_page_check_label')}
                </label>
                <div className="block-page__checker">
                    <input
                        id="block-page-domain-check"
                        type="text"
                        className="form-control block-page__checker-input"
                        placeholder={t('block_page_check_placeholder')}
                        value={domainQuery}
                        onChange={(e) => setDomainQuery(e.target.value)}
                    />
                    <button
                        type="submit"
                        className="btn btn-outline-secondary btn-standard block-page__checker-btn"
                        disabled={!normalizeDomainInput(domainQuery)}>
                        {t('block_page_check_btn')}
                    </button>
                </div>
            </form>
        </Card>
    );
};

export default BlockPage;
