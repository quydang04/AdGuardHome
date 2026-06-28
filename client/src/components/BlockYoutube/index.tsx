import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDispatch, useSelector } from 'react-redux';

import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';
import Loading from '../ui/Loading';

import { getYoutubeConfig, setYoutubeConfig } from '../../actions/youtube';
import { RootState } from '../../initialState';

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
    };

    if (youtube?.processingGet) {
        return <Loading />;
    }

    return (
        <>
            <PageTitle title={t('block_youtube')} subtitle={t('block_youtube_desc')} />

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
