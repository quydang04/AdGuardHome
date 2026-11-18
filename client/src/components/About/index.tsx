import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';
import cn from 'clsx';

import PageTitle from '../ui/PageTitle';
import { getVersion, getUpdate } from '../../actions';
import { MANUAL_UPDATE_LINK, PRIVACY_POLICY_LINK } from '../../helpers/constants';
import { RootState } from '../../initialState';
import './About.css';

const ShieldHeroIcon = () => (
    <svg width="36" height="36" viewBox="0 0 24 24" fill="#67b279" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z" />
    </svg>
);

const RefreshIcon = ({ isSpinning }: { isSpinning?: boolean }) => (
    <svg
        width="15"
        height="15"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={cn('about-page__spin-icon', { 'about-page__spin-icon--active': isSpinning })}>
        <polyline points="23 4 23 10 17 10" />
        <polyline points="1 20 1 14 7 14" />
        <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
    </svg>
);

const ExternalLinkIcon = () => (
    <svg
        width="13"
        height="13"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="about-page__ext-icon">
        <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
        <polyline points="15 3 21 3 21 9" />
        <line x1="10" y1="14" x2="21" y2="3" />
    </svg>
);

const About = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    const {
        dnsVersion,
        processingVersion,
        checkUpdateFlag,
        isUpdateAvailable,
        newVersion,
        announcementUrl,
        canAutoUpdate,
        processingUpdate,
        protectionEnabled,
        dnsPort,
        httpPort,
    } = useSelector((state: RootState) => state?.dashboard || ({} as any), shallowEqual);

    const installVersion = useSelector((state: RootState) => state?.install?.dnsVersion);
    const rawVersion = dnsVersion || installVersion;
    const formattedVersion = rawVersion ? `v${String(rawVersion).replace(/^v/i, '')}` : '–';
    const formattedNewVersion = newVersion ? `v${String(newVersion).replace(/^v/i, '')}` : '';

    const manualUpdateHref =
        announcementUrl && announcementUrl.includes('hub.docker.com') ? announcementUrl : MANUAL_UPDATE_LINK;

    const handleCheckUpdates = () => {
        dispatch(getVersion(true) as any);
    };

    const handleAutoUpdate = () => {
        dispatch(getUpdate() as any);
    };

    const currentYear = new Date().getFullYear();

    return (
        <div className="about-page">
            <PageTitle title={t('about_title', 'Giới thiệu')} containerClass="page-title--overview" />

            {/* 1. Hero / Introduction Card */}
            <div className="about-hero-card">
                <div className="about-hero-card__top">
                    <div className="about-hero-card__brand">
                        <div className="about-hero-card__logo">
                            <ShieldHeroIcon />
                        </div>
                        <div className="about-hero-card__title-group">
                            <div className="about-hero-card__title-row">
                                <h2 className="about-hero-card__name">AdGuard Home</h2>
                                {formattedVersion !== '–' && (
                                    <span className="about-hero-card__version-pill">{formattedVersion}</span>
                                )}
                                <span className="about-hero-card__edition-badge">
                                    {t('custom_build_badge', 'Custom Edition')}
                                </span>
                            </div>
                            <p className="about-hero-card__subtitle">
                                {t(
                                    'about_intro_desc',
                                    'Giải pháp máy chủ DNS chặn quảng cáo, trình theo dõi và tên miền độc hại trên toàn mạng với khả năng giám sát thời gian thực.',
                                )}
                            </p>
                        </div>
                    </div>
                </div>

                <div className="about-hero-card__mod-desc">
                    {t(
                        'custom_build_desc',
                        'Bản dựng AdGuard Home được tùy chỉnh thiết kế giao diện Flat Dark hiện đại và tối ưu trải nghiệm người dùng bởi quydangblog.',
                    )}
                </div>

                <div className="about-hero-card__features">
                    <span className="about-hero-card__feature-tag">
                        <span className="about-hero-card__feature-dot" />
                        {t('about_feature_dns', 'Chặn quảng cáo & theo dõi toàn mạng')}
                    </span>
                    <span className="about-hero-card__feature-tag">
                        <span className="about-hero-card__feature-dot" />
                        {t('about_feature_encryption', 'DNS mã hóa (DoH / DoT / DoQ / DNSCrypt)')}
                    </span>
                    <span className="about-hero-card__feature-tag">
                        <span className="about-hero-card__feature-dot" />
                        {t('about_feature_analytics', 'Nhật ký & Thống kê thời gian thực')}
                    </span>
                    <span className="about-hero-card__feature-tag">
                        <span className="about-hero-card__feature-dot" />
                        {t('about_feature_custom', 'Giao diện Flat Dark & Sao lưu đa phiên bản')}
                    </span>
                </div>
            </div>

            {/* 2. Version Info & Update Check + Build & Links Grid */}
            <div className="about-page__grid">
                {/* Left Card: Version & Update Check */}
                <div className="about-card">
                    <div className="about-card__header">
                        <div className="about-card__title">
                            {t('about_version_card_title', 'Thông tin phiên bản & Cập nhật')}
                        </div>
                        {checkUpdateFlag !== false && (
                            <button
                                type="button"
                                className="about-card__check-btn"
                                onClick={handleCheckUpdates}
                                disabled={processingVersion}
                                data-testid="about_check_updates">
                                <RefreshIcon isSpinning={processingVersion} />
                                <span>{t('check_updates_btn', 'Kiểm tra cập nhật')}</span>
                            </button>
                        )}
                    </div>

                    <ul className="about-card__rows">
                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('about_current_version', 'Phiên bản hiện tại')}
                            </span>
                            <span className="about-card__row-value">
                                <span className="about-card__version-badge">{formattedVersion}</span>
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('about_update_status', 'Trạng thái cập nhật')}
                            </span>
                            <span className="about-card__row-value">
                                {processingVersion ? (
                                    <span className="about-card__status about-card__status--checking">
                                        {t('about_checking_updates', 'Đang kiểm tra...')}
                                    </span>
                                ) : isUpdateAvailable && formattedNewVersion ? (
                                    <span className="about-card__status about-card__status--available">
                                        {t('updates_checked', 'Phiên bản mới của AdGuard Home có sẵn')} ({formattedNewVersion})
                                    </span>
                                ) : checkUpdateFlag === false ? (
                                    <span className="about-card__status about-card__status--muted">
                                        {t('about_updates_disabled', 'Đã tắt kiểm tra cập nhật')}
                                    </span>
                                ) : (
                                    <span className="about-card__status about-card__status--uptodate">
                                        {t('updates_version_equal', 'AdGuard Home đã được cập nhật')}
                                    </span>
                                )}
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('about_protection_status', 'Trạng thái bảo vệ DNS')}
                            </span>
                            <span className="about-card__row-value">
                                <span
                                    className={cn('about-card__pill', {
                                        'about-card__pill--success': protectionEnabled,
                                        'about-card__pill--danger': !protectionEnabled,
                                    })}>
                                    {t(protectionEnabled ? 'on' : 'off')}
                                </span>
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('about_service_ports', 'Cổng dịch vụ (DNS / Web)')}
                            </span>
                            <span className="about-card__row-value about-card__row-value--mono">
                                {dnsPort ? `DNS :${dnsPort}` : '–'}
                                {httpPort ? ` · Web :${httpPort}` : ''}
                            </span>
                        </li>
                    </ul>

                    {isUpdateAvailable && formattedNewVersion && (
                        <div className="about-card__update-box">
                            <div className="about-card__update-info">
                                <Trans
                                    values={{ version: formattedNewVersion }}
                                    components={[
                                        <a
                                            href={announcementUrl}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            key="0"
                                            className="about-card__update-link">
                                            Click here
                                        </a>,
                                    ]}>
                                    update_announcement
                                </Trans>
                            </div>
                            <div className="about-card__update-actions">
                                {canAutoUpdate ? (
                                    <button
                                        type="button"
                                        className="btn btn-sm btn-success"
                                        onClick={handleAutoUpdate}
                                        disabled={processingUpdate}>
                                        <Trans>update_now</Trans>
                                    </button>
                                ) : (
                                    <span className="about-card__manual-update">
                                        <Trans
                                            components={{
                                                a: (
                                                    <a
                                                        href={manualUpdateHref}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        key="0"
                                                        className="about-card__update-link">
                                                        Link
                                                    </a>
                                                ),
                                            }}>
                                            manual_update
                                        </Trans>
                                    </span>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                {/* Right Card: Build Info, Author & Useful Links */}
                <div className="about-card">
                    <div className="about-card__header">
                        <div className="about-card__title">
                            {t('about_links_card_title', 'Thông tin bản dựng & Liên kết')}
                        </div>
                    </div>

                    <ul className="about-card__rows">
                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('developed_by', 'Tác giả')}
                            </span>
                            <span className="about-card__row-value">
                                <a
                                    href="http://quydang.name.vn"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="about-card__ext-link">
                                    <span>quydangblog (quydang.name.vn)</span>
                                    <ExternalLinkIcon />
                                </a>
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('about_github_repo', 'Mã nguồn GitHub')}
                            </span>
                            <span className="about-card__row-value">
                                <a
                                    href="https://github.com/quydang04/AdGuardHome"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="about-card__ext-link">
                                    <span>quydang04/AdGuardHome</span>
                                    <ExternalLinkIcon />
                                </a>
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('about_docker_hub', 'Docker Hub')}
                            </span>
                            <span className="about-card__row-value">
                                <a
                                    href="https://hub.docker.com/r/quydang04/adguardhome"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="about-card__ext-link">
                                    <span>quydang04/adguardhome</span>
                                    <ExternalLinkIcon />
                                </a>
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('homepage', 'Trang chủ')}
                            </span>
                            <span className="about-card__row-value">
                                <a
                                    href="https://link.adtidy.org/forward.html?action=home&from=ui&app=home"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="about-card__ext-link">
                                    <span>AdGuard Home</span>
                                    <ExternalLinkIcon />
                                </a>
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('privacy_policy', 'Chính sách bảo mật')}
                            </span>
                            <span className="about-card__row-value">
                                <a
                                    href={PRIVACY_POLICY_LINK}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="about-card__ext-link">
                                    <span>{t('privacy_policy', 'Chính sách bảo mật')}</span>
                                    <ExternalLinkIcon />
                                </a>
                            </span>
                        </li>

                        <li className="about-card__row">
                            <span className="about-card__row-label">
                                {t('report_an_issue', 'Báo cáo lỗi')}
                            </span>
                            <span className="about-card__row-value">
                                <a
                                    href="https://github.com/quydang04/AdGuardHome/issues"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="about-card__ext-link">
                                    <span>GitHub Issues</span>
                                    <ExternalLinkIcon />
                                </a>
                            </span>
                        </li>
                    </ul>
                </div>
            </div>

            {/* 3. Copyright Banner */}
            <div className="about-copyright-card">
                <span className="about-copyright-card__text">
                    Copyright &copy; {currentYear}{' '}
                    <a
                        target="_blank"
                        rel="noopener noreferrer"
                        href="https://link.adtidy.org/forward.html?action=home&from=ui&app=home">
                        AdGuard
                    </a>
                    {' '}&ndash; Customize by{' '}
                    <a
                        target="_blank"
                        rel="noopener noreferrer"
                        href="http://quydang.name.vn">
                        quydangblog
                    </a>
                </span>
            </div>
        </div>
    );
};

export default About;
