import React, { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useDispatch, useSelector } from 'react-redux';

import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';
import Loading from '../ui/Loading';

import { getYoutubeConfig, setYoutubeConfig, getYoutubeStatus } from '../../actions/youtube';
import { RootState, YoutubeIPStatus } from '../../initialState';

const BlockYoutube = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    const youtube = useSelector((state: RootState) => state.youtube);

    const [enabled, setEnabled] = useState(false);
    const [routeServer, setRouteServer] = useState('');
    const [blockAds, setBlockAds] = useState(true);
    const [blockTracking, setBlockTracking] = useState(true);
    const [customDomainsText, setCustomDomainsText] = useState('');

    useEffect(() => {
        dispatch(getYoutubeConfig());
        dispatch(getYoutubeStatus());
    }, [dispatch]);

    useEffect(() => {
        if (youtube && !youtube.processingGet) {
            setEnabled(youtube.enabled || false);
            setRouteServer(youtube.route_server || '');
            setBlockAds(youtube.block_ads !== undefined ? youtube.block_ads : true);
            setBlockTracking(youtube.block_tracking !== undefined ? youtube.block_tracking : true);
            setCustomDomainsText((youtube.custom_domains || []).join('\n'));
        }
    }, [youtube?.processingGet]);

    useEffect(() => {
        const interval = setInterval(() => {
            dispatch(getYoutubeStatus());
        }, 15000);

        return () => clearInterval(interval);
    }, [dispatch]);

    const handleRefreshStatus = useCallback(() => {
        dispatch(getYoutubeStatus());
    }, [dispatch]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const customDomains = customDomainsText
            .split('\n')
            .map((d: string) => d.trim())
            .filter((d: string) => d.length > 0);

        dispatch(
            setYoutubeConfig({
                enabled,
                route_server: routeServer,
                block_ads: blockAds,
                block_tracking: blockTracking,
                custom_domains: customDomains,
            }),
        );

        setTimeout(() => {
            dispatch(getYoutubeStatus());
        }, 2000);
    };

    if (youtube?.processingGet) {
        return <Loading />;
    }

    const status = youtube?.status;
    const formatTime = (timeStr: string) => {
        if (!timeStr) return '-';
        try {
            const date = new Date(timeStr);
            return date.toLocaleString();
        } catch {
            return timeStr;
        }
    };

    return (
        <>
            <PageTitle title={t('block_youtube')} subtitle={t('block_youtube_desc')} />

            {/* Dashboard Card */}
            <Card
                title={t('youtube_dashboard')}
                bodyType="card-body box-body--settings">
                <div className="row">
                    {/* Overall Status */}
                    <div className="col-sm-6 col-lg-3 mb-3">
                        <div className="card" style={{
                            border: `2px solid ${status?.active ? '#28a745' : '#6c757d'}`,
                            borderRadius: '8px',
                        }}>
                            <div className="card-body text-center p-3">
                                <div style={{
                                    fontSize: '2rem',
                                    color: status?.active ? '#28a745' : '#6c757d',
                                    marginBottom: '4px',
                                }}>
                                    {status?.active ? '●' : '○'}
                                </div>
                                <div style={{ fontSize: '1.1rem', fontWeight: 600 }}>
                                    {status?.active ? t('youtube_status_active') : t('youtube_status_inactive')}
                                </div>
                                {status?.uptime && (
                                    <div className="text-muted" style={{ fontSize: '0.85rem' }}>
                                        {t('youtube_uptime')}: {status.uptime}
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Route Server Health */}
                    <div className="col-sm-6 col-lg-3 mb-3">
                        <div className="card" style={{
                            border: '2px solid #17a2b8',
                            borderRadius: '8px',
                        }}>
                            <div className="card-body text-center p-3">
                                <div style={{ fontSize: '2rem', color: '#17a2b8', marginBottom: '4px' }}>
                                    {status?.healthy_ips ?? 0}/{status?.total_ips ?? 0}
                                </div>
                                <div style={{ fontSize: '1.1rem', fontWeight: 600 }}>
                                    {t('youtube_healthy_ips')}
                                </div>
                                <div className="text-muted" style={{ fontSize: '0.85rem' }}>
                                    {status?.route_server || '-'}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Blocked Rules */}
                    <div className="col-sm-6 col-lg-3 mb-3">
                        <div className="card" style={{
                            border: '2px solid #dc3545',
                            borderRadius: '8px',
                        }}>
                            <div className="card-body text-center p-3">
                                <div style={{ fontSize: '2rem', color: '#dc3545', marginBottom: '4px' }}>
                                    {status?.blocked_rules ?? 0}
                                </div>
                                <div style={{ fontSize: '1.1rem', fontWeight: 600 }}>
                                    {t('youtube_blocked_rules')}
                                </div>
                                <div className="text-muted" style={{ fontSize: '0.85rem' }}>
                                    {t('youtube_ad_tracking_domains')}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Active Rewrites */}
                    <div className="col-sm-6 col-lg-3 mb-3">
                        <div className="card" style={{
                            border: '2px solid #007bff',
                            borderRadius: '8px',
                        }}>
                            <div className="card-body text-center p-3">
                                <div style={{ fontSize: '2rem', color: '#007bff', marginBottom: '4px' }}>
                                    {status?.active_rewrites ?? 0}
                                </div>
                                <div style={{ fontSize: '1.1rem', fontWeight: 600 }}>
                                    {t('youtube_active_rewrites')}
                                </div>
                                <div className="text-muted" style={{ fontSize: '0.85rem' }}>
                                    {t('youtube_dns_entries')}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Sync Info */}
                <div className="row mt-2">
                    <div className="col-12">
                        <div className="d-flex align-items-center justify-content-between"
                            style={{
                                backgroundColor: '#f8f9fa',
                                borderRadius: '6px',
                                padding: '10px 16px',
                            }}>
                            <div>
                                <span className="text-muted" style={{ fontSize: '0.85rem' }}>
                                    {t('youtube_last_sync')}: {formatTime(status?.last_sync_time || '')}
                                    {' | '}
                                    {t('youtube_total_syncs')}: {status?.total_syncs ?? 0}
                                    {' | '}
                                    {t('youtube_sync_status')}: {status?.last_sync_status || '-'}
                                </span>
                            </div>
                            <button
                                type="button"
                                className="btn btn-sm btn-outline-secondary"
                                onClick={handleRefreshStatus}
                                disabled={youtube?.processingStatus}>
                                {t('youtube_refresh')}
                            </button>
                        </div>
                    </div>
                </div>

                {/* IP Health Table */}
                {status?.ip_statuses && status.ip_statuses.length > 0 && (
                    <div className="mt-3">
                        <h6 className="mb-2">{t('youtube_ip_health')}</h6>
                        <div className="table-responsive">
                            <table className="table table-sm">
                                <thead>
                                    <tr>
                                        <th>{t('youtube_ip_address')}</th>
                                        <th>{t('youtube_health_status')}</th>
                                        <th>{t('youtube_fail_count')}</th>
                                        <th>{t('youtube_last_check')}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {status.ip_statuses.map((ipStatus: YoutubeIPStatus) => (
                                        <tr key={ipStatus.ip}>
                                            <td><code>{ipStatus.ip}</code></td>
                                            <td>
                                                <span className={`badge ${ipStatus.healthy ? 'badge-success' : 'badge-danger'}`}>
                                                    {ipStatus.healthy
                                                        ? t('youtube_ip_healthy')
                                                        : t('youtube_ip_unhealthy')}
                                                </span>
                                            </td>
                                            <td>
                                                <span className={ipStatus.fail_count > 0 ? 'text-warning' : ''}>
                                                    {ipStatus.fail_count}
                                                </span>
                                            </td>
                                            <td style={{ fontSize: '0.85rem' }}>
                                                {formatTime(ipStatus.last_check)}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}
            </Card>

            {/* Status Card */}
            <Card
                title={t('youtube_status')}
                bodyType="card-body box-body--settings">
                <div className="form">
                    <div className="form__group form__group--settings">
                        <label className="form__label form__label--with-desc" htmlFor="youtube_enabled">
                            <span className="form__label-text">
                                {t('youtube_enable')}
                            </span>
                            <span className="form__desc form__desc--top">
                                {t('youtube_enable_desc')}
                            </span>
                        </label>
                        <div className="form__control">
                            <div className="custom-switch">
                                <input
                                    type="checkbox"
                                    id="youtube_enabled"
                                    className="custom-switch__input"
                                    checked={enabled}
                                    onChange={(e) => setEnabled(e.target.checked)}
                                />
                                <label className="custom-switch__label" htmlFor="youtube_enabled">
                                    {enabled ? t('enabled') : t('disabled')}
                                </label>
                            </div>
                        </div>
                    </div>
                </div>
            </Card>

            {/* Configuration Card */}
            <Card
                title={t('youtube_config')}
                bodyType="card-body box-body--settings">
                <form onSubmit={handleSubmit}>
                    {/* Route Server */}
                    <div className="form__group form__group--settings mb-3">
                        <label className="form__label form__label--with-desc" htmlFor="route_server">
                            <span className="form__label-text">
                                {t('youtube_route_server')}
                            </span>
                            <span className="form__desc form__desc--top">
                                {t('youtube_route_server_desc')}
                            </span>
                        </label>
                        <div className="form__control">
                            <input
                                type="text"
                                id="route_server"
                                className="form-control"
                                placeholder="ytb.fzpn.net"
                                value={routeServer}
                                onChange={(e) => setRouteServer(e.target.value)}
                            />
                        </div>
                    </div>

                    {/* Block Ads */}
                    <div className="form__group form__group--settings mb-3">
                        <label className="form__label form__label--with-desc" htmlFor="block_ads">
                            <span className="form__label-text">
                                {t('youtube_block_ads')}
                            </span>
                            <span className="form__desc form__desc--top">
                                {t('youtube_block_ads_desc')}
                            </span>
                        </label>
                        <div className="form__control">
                            <div className="custom-switch">
                                <input
                                    type="checkbox"
                                    id="block_ads"
                                    className="custom-switch__input"
                                    checked={blockAds}
                                    onChange={(e) => setBlockAds(e.target.checked)}
                                />
                                <label className="custom-switch__label" htmlFor="block_ads">
                                    {blockAds ? t('enabled') : t('disabled')}
                                </label>
                            </div>
                        </div>
                    </div>

                    {/* Block Tracking */}
                    <div className="form__group form__group--settings mb-3">
                        <label className="form__label form__label--with-desc" htmlFor="block_tracking">
                            <span className="form__label-text">
                                {t('youtube_block_tracking')}
                            </span>
                            <span className="form__desc form__desc--top">
                                {t('youtube_block_tracking_desc')}
                            </span>
                        </label>
                        <div className="form__control">
                            <div className="custom-switch">
                                <input
                                    type="checkbox"
                                    id="block_tracking"
                                    className="custom-switch__input"
                                    checked={blockTracking}
                                    onChange={(e) => setBlockTracking(e.target.checked)}
                                />
                                <label className="custom-switch__label" htmlFor="block_tracking">
                                    {blockTracking ? t('enabled') : t('disabled')}
                                </label>
                            </div>
                        </div>
                    </div>

                    {/* Custom Domains */}
                    <div className="form__group form__group--settings mb-3">
                        <label className="form__label form__label--with-desc" htmlFor="custom_domains">
                            <span className="form__label-text">
                                {t('youtube_custom_domains')}
                            </span>
                            <span className="form__desc form__desc--top">
                                {t('youtube_custom_domains_desc')}
                            </span>
                        </label>
                        <div className="form__control">
                            <textarea
                                id="custom_domains"
                                className="form-control"
                                rows={4}
                                placeholder={t('youtube_custom_domains_placeholder')}
                                value={customDomainsText}
                                onChange={(e) => setCustomDomainsText(e.target.value)}
                            />
                        </div>
                    </div>

                    <div className="card-actions">
                        <div className="btn-list">
                            <button
                                type="submit"
                                className="btn btn-success btn-standard"
                                disabled={youtube?.processingSet}>
                                {t('save_config')}
                            </button>
                        </div>
                    </div>
                </form>
            </Card>

            {/* YouTube DNS Rewrite Domains */}
            <Card
                title={t('youtube_rewrite_domains')}
                bodyType="card-body box-body--settings">
                <div className="form__desc mb-3">
                    {t('youtube_rewrite_domains_desc')}
                </div>
                <div className="table-responsive">
                    <table className="table">
                        <thead>
                            <tr>
                                <th>{t('domain')}</th>
                                <th>{t('type')}</th>
                            </tr>
                        </thead>
                        <tbody>
                            {(youtube?.rewrite_domains || []).map((domain: string) => (
                                <tr key={`rewrite-${domain}`}>
                                    <td><code>{domain}</code></td>
                                    <td>
                                        <span className="badge badge-primary">
                                            {t('youtube_type_rewrite')}
                                        </span>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </Card>

            {/* Blocked Ad Domains Info */}
            <Card
                title={t('youtube_ad_domains')}
                bodyType="card-body box-body--settings">
                <div className="form__desc mb-3">
                    {t('youtube_ad_domains_desc')}
                </div>
                <div className="table-responsive">
                    <table className="table">
                        <thead>
                            <tr>
                                <th>{t('domain')}</th>
                                <th>{t('type')}</th>
                            </tr>
                        </thead>
                        <tbody>
                            {(youtube?.ad_domains || []).map((domain: string) => (
                                <tr key={`ad-${domain}`}>
                                    <td><code>{domain}</code></td>
                                    <td>
                                        <span className="badge badge-danger">
                                            {t('youtube_type_ad')}
                                        </span>
                                    </td>
                                </tr>
                            ))}
                            {(youtube?.tracking_domains || []).map((domain: string) => (
                                <tr key={`track-${domain}`}>
                                    <td><code>{domain}</code></td>
                                    <td>
                                        <span className="badge badge-warning">
                                            {t('youtube_type_tracking')}
                                        </span>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </Card>

            {/* Setup Guide Card */}
            <Card
                title={t('youtube_setup_guide')}
                bodyType="card-body box-body--settings">
                <div className="markdown-text">
                    <h6>{t('youtube_guide_router_title')}</h6>
                    <ol>
                        <li>{t('youtube_guide_step1')}</li>
                        <li>{t('youtube_guide_step2')}</li>
                        <li>{t('youtube_guide_step3')}</li>
                        <li>{t('youtube_guide_step4')}</li>
                    </ol>

                    <h6>{t('youtube_guide_device_title')}</h6>
                    <ul>
                        <li><strong>Android:</strong> {t('youtube_guide_android')}</li>
                        <li><strong>iPhone/iPad:</strong> {t('youtube_guide_ios')}</li>
                        <li><strong>Windows:</strong> {t('youtube_guide_windows')}</li>
                    </ul>

                    <div className="alert alert-warning mt-3">
                        <strong>{t('youtube_guide_warning_title')}</strong>
                        <br />
                        {t('youtube_guide_warning_text')}
                    </div>
                </div>
            </Card>
        </>
    );
};

export default BlockYoutube;
