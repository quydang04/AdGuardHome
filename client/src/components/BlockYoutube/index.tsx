import React, { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useDispatch, useSelector } from 'react-redux';

import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';
import Loading from '../ui/Loading';

import { getYoutubeConfig, setYoutubeConfig, getYoutubeStatus, getYoutubeStats } from '../../actions/youtube';
import { RootState, YoutubeIPStatus } from '../../initialState';

import YouTubeStats from './YouTubeStats';

import './BlockYoutube.css';

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
        dispatch(getYoutubeStats());
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
            dispatch(getYoutubeStats());
        }, 5000);

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

            {/* Dashboard */}
            <Card
                title={t('youtube_dashboard')}
                bodyType="card-body box-body--settings">
                <div className="yt-dashboard">
                    <div className="yt-stats-grid">
                        <div className={`yt-stat-card ${status?.active ? 'yt-stat-card--success' : 'yt-stat-card--muted'}`}>
                            <div className="yt-stat-card__icon">
                                {status?.active ? '●' : '○'}
                            </div>
                            <div className="yt-stat-card__value">
                                {status?.active ? t('youtube_status_active') : t('youtube_status_inactive')}
                            </div>
                            {status?.uptime && (
                                <div className="yt-stat-card__label">
                                    {t('youtube_uptime')}: {status.uptime}
                                </div>
                            )}
                        </div>

                        <div className="yt-stat-card yt-stat-card--info">
                            <div className="yt-stat-card__number">
                                {status?.healthy_ips ?? 0}/{status?.total_ips ?? 0}
                            </div>
                            <div className="yt-stat-card__value">
                                {t('youtube_healthy_ips')}
                            </div>
                            <div className="yt-stat-card__label">
                                {status?.route_server || '-'}
                            </div>
                        </div>

                        <div className="yt-stat-card yt-stat-card--danger">
                            <div className="yt-stat-card__number">
                                {status?.blocked_rules ?? 0}
                            </div>
                            <div className="yt-stat-card__value">
                                {t('youtube_blocked_rules')}
                            </div>
                            <div className="yt-stat-card__label">
                                {t('youtube_ad_tracking_domains')}
                            </div>
                        </div>

                        <div className="yt-stat-card yt-stat-card--primary">
                            <div className="yt-stat-card__number">
                                {status?.active_rewrites ?? 0}
                            </div>
                            <div className="yt-stat-card__value">
                                {t('youtube_active_rewrites')}
                            </div>
                            <div className="yt-stat-card__label">
                                {t('youtube_dns_entries')}
                            </div>
                        </div>
                    </div>

                    {/* Sync Info Bar */}
                    <div className="yt-sync-bar">
                        <div className="yt-sync-bar__info">
                            <span>{t('youtube_last_sync')}: <strong>{formatTime(status?.last_sync_time || '')}</strong></span>
                            <span>{t('youtube_total_syncs')}: <strong>{status?.total_syncs ?? 0}</strong></span>
                            <span>{t('youtube_sync_status')}: <strong>{status?.last_sync_status || '-'}</strong></span>
                        </div>
                        <button
                            type="button"
                            className="btn btn-sm btn-outline-secondary"
                            onClick={handleRefreshStatus}
                            disabled={youtube?.processingStatus}>
                            {t('youtube_refresh')}
                        </button>
                    </div>

                    {/* IP Health Table */}
                    {status?.ip_statuses && status.ip_statuses.length > 0 && (
                        <div className="yt-ip-health">
                            <h6 className="yt-ip-health__title">{t('youtube_ip_health')}</h6>
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
                                                <td className="yt-ip-health__time">
                                                    {formatTime(ipStatus.last_check)}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    )}
                </div>
            </Card>

            {/* Configuration Card */}
            <Card
                title={t('youtube_config')}
                bodyType="card-body box-body--settings">
                <form onSubmit={handleSubmit}>
                    {/* Enable Toggle */}
                    <div className="yt-setting">
                        <div className="yt-setting__header">
                            <label className="yt-setting__label" htmlFor="youtube_enabled">
                                {t('youtube_enable')}
                            </label>
                            <p className="yt-setting__desc">
                                {t('youtube_enable_desc')}
                            </p>
                        </div>
                        <div className="yt-setting__control">
                            <label className="yt-toggle" htmlFor="youtube_enabled">
                                <input
                                    type="checkbox"
                                    id="youtube_enabled"
                                    className="yt-toggle__input"
                                    checked={enabled}
                                    onChange={(e) => setEnabled(e.target.checked)}
                                />
                                <span className="yt-toggle__slider" />
                                <span className="yt-toggle__text">
                                    {enabled ? t('enabled') : t('disabled')}
                                </span>
                            </label>
                        </div>
                    </div>

                    {/* Route Server */}
                    <div className="yt-setting">
                        <div className="yt-setting__header">
                            <label className="yt-setting__label" htmlFor="route_server">
                                {t('youtube_route_server')}
                            </label>
                            <p className="yt-setting__desc">
                                {t('youtube_route_server_desc')}
                            </p>
                        </div>
                        <div className="yt-setting__control">
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
                    <div className="yt-setting">
                        <div className="yt-setting__header">
                            <label className="yt-setting__label" htmlFor="block_ads">
                                {t('youtube_block_ads')}
                            </label>
                            <p className="yt-setting__desc">
                                {t('youtube_block_ads_desc')}
                            </p>
                        </div>
                        <div className="yt-setting__control">
                            <label className="yt-toggle" htmlFor="block_ads">
                                <input
                                    type="checkbox"
                                    id="block_ads"
                                    className="yt-toggle__input"
                                    checked={blockAds}
                                    onChange={(e) => setBlockAds(e.target.checked)}
                                />
                                <span className="yt-toggle__slider" />
                                <span className="yt-toggle__text">
                                    {blockAds ? t('enabled') : t('disabled')}
                                </span>
                            </label>
                        </div>
                    </div>

                    {/* Block Tracking */}
                    <div className="yt-setting">
                        <div className="yt-setting__header">
                            <label className="yt-setting__label" htmlFor="block_tracking">
                                {t('youtube_block_tracking')}
                            </label>
                            <p className="yt-setting__desc">
                                {t('youtube_block_tracking_desc')}
                            </p>
                        </div>
                        <div className="yt-setting__control">
                            <label className="yt-toggle" htmlFor="block_tracking">
                                <input
                                    type="checkbox"
                                    id="block_tracking"
                                    className="yt-toggle__input"
                                    checked={blockTracking}
                                    onChange={(e) => setBlockTracking(e.target.checked)}
                                />
                                <span className="yt-toggle__slider" />
                                <span className="yt-toggle__text">
                                    {blockTracking ? t('enabled') : t('disabled')}
                                </span>
                            </label>
                        </div>
                    </div>

                    {/* Custom Domains */}
                    <div className="yt-setting yt-setting--stacked">
                        <div className="yt-setting__header">
                            <label className="yt-setting__label" htmlFor="custom_domains">
                                {t('youtube_custom_domains')}
                            </label>
                            <p className="yt-setting__desc">
                                {t('youtube_custom_domains_desc')}
                            </p>
                        </div>
                        <div className="yt-setting__control yt-setting__control--full">
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
                <p className="yt-setting__desc mb-3">
                    {t('youtube_rewrite_domains_desc')}
                </p>
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
                <p className="yt-setting__desc mb-3">
                    {t('youtube_ad_domains_desc')}
                </p>
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

            {/* YouTube Query Statistics */}
            <Card
                title={t('youtube_query_stats')}
                bodyType="card-body box-body--settings">
                <YouTubeStats />
            </Card>

        </>
    );
};

export default BlockYoutube;
