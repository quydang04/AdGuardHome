import React, { Component, Fragment } from 'react';
import { withTranslation } from 'react-i18next';
import cn from 'classnames';

import i18next from 'i18next';
import StatsConfig from './StatsConfig';
import LogsConfig from './LogsConfig';
import { FiltersConfig } from './FiltersConfig';
import { Checkbox } from '../ui/Controls/Checkbox';
import Loading from '../ui/Loading';
import PageTitle from '../ui/PageTitle';
import Card from '../ui/Card';

import { getObjectKeysSorted, captitalizeWords, setHtmlLangAttr, setUITheme } from '../../helpers/helpers';
import { THEMES } from '../../helpers/constants';
import { LANGUAGES } from '../../helpers/twosky';
import apiClient from '../../api/Api';
import './Settings.css';
import { SettingsData, DashboardData } from '../../initialState';

const ORDER_KEY = 'order';

const SETTINGS = {
    safebrowsing: {
        enabled: false,
        title: i18next.t('use_adguard_browsing_sec'),
        subtitle: i18next.t('use_adguard_browsing_sec_hint'),
        testId: 'safebrowsing',
        [ORDER_KEY]: 0,
    },
    parental: {
        enabled: false,
        title: i18next.t('use_adguard_parental'),
        subtitle: i18next.t('use_adguard_parental_hint'),
        testId: 'parental',
        [ORDER_KEY]: 1,
    },
};

type ThemeName = keyof typeof THEMES;

interface SettingsProps {
    initSettings: (...args: unknown[]) => unknown;
    settings: SettingsData;
    toggleSetting: (...args: unknown[]) => unknown;
    getStatsConfig: (...args: unknown[]) => unknown;
    setStatsConfig: (...args: unknown[]) => unknown;
    resetStats: (...args: unknown[]) => unknown;
    setFiltersConfig: (...args: unknown[]) => unknown;
    getFilteringStatus: (...args: unknown[]) => unknown;
    changeLanguage: (lang: string) => unknown;
    changeTheme: (theme: string) => unknown;
    t: (...args: unknown[]) => string;
    getLogsConfig?: (...args: unknown[]) => unknown;
    setLogsConfig?: (...args: unknown[]) => unknown;
    clearLogs?: (...args: unknown[]) => unknown;
    dashboard?: DashboardData;
    stats?: {
        processingGetConfig?: boolean;
        interval?: number;
        customInterval?: number;
        enabled?: boolean;
        ignored?: unknown[];
        processingSetConfig?: boolean;
        processingReset?: boolean;
    };
    queryLogs?: {
        enabled?: boolean;
        interval?: number;
        customInterval?: number;
        anonymize_client_ip?: boolean;
        processingSetConfig?: boolean;
        processingClear?: boolean;
        processingGetConfig?: boolean;
        ignored?: unknown[];
    };
    filtering?: {
        interval?: number;
        enabled?: boolean;
        processingSetConfig?: boolean;
    };
}

interface SettingsState {
    currentPassword: string;
    newPassword: string;
    confirmPassword: string;
    passwordMessage: string;
    passwordMessageType: 'success' | 'error' | '';
    passwordProcessing: boolean;
    port: string;
    portMessage: string;
    portMessageType: 'success' | 'error' | '';
    portProcessing: boolean;
}

class Settings extends Component<SettingsProps, SettingsState> {
    state: SettingsState = {
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
        passwordMessage: '',
        passwordMessageType: '',
        passwordProcessing: false,
        port: '',
        portMessage: '',
        portMessageType: '',
        portProcessing: false,
    };

    componentDidMount() {
        this.props.initSettings(SETTINGS);
        this.props.getStatsConfig();
        this.props.getLogsConfig();
        this.props.getFilteringStatus();

        const httpPort = this.props.dashboard?.httpPort;
        if (httpPort) {
            this.setState({ port: String(httpPort) });
        }
    }

    componentDidUpdate(prevProps: SettingsProps) {
        const httpPort = this.props.dashboard?.httpPort;
        const prevPort = prevProps.dashboard?.httpPort;
        if (httpPort && httpPort !== prevPort && this.state.port === '') {
            this.setState({ port: String(httpPort) });
        }
    }

    renderSettings = (settings: any) =>
        getObjectKeysSorted(SETTINGS, ORDER_KEY).map((key: any) => {
            const setting = settings[key];
            const { enabled, title, subtitle, testId } = setting;

            return (
                <div key={key} className="form__group form__group--checkbox">
                    <Checkbox
                        data-testid={testId}
                        value={enabled}
                        title={title}
                        subtitle={subtitle}
                        onChange={(checked) => this.props.toggleSetting(key, !checked)}
                    />
                </div>
            );
        });

    renderSafeSearch = () => {
        const {
            settings: {
                settingsList: { safesearch },
            },
        } = this.props;
        const { enabled } = safesearch || {};
        const searches = { ...(safesearch || {}) };
        delete searches.enabled;

        return (
            <>
                <div className="form__group form__group--checkbox">
                    <Checkbox
                        data-testid="safesearch"
                        value={enabled}
                        title={i18next.t('enforce_safe_search')}
                        subtitle={i18next.t('enforce_save_search_hint')}
                        onChange={(checked) =>
                            this.props.toggleSetting('safesearch', { ...safesearch, enabled: checked })
                        }
                    />
                </div>

                <div className="form__group--inner">
                    {Object.keys(searches).map((searchKey) => (
                        <div key={searchKey} className="form__group form__group--checkbox">
                            <Checkbox
                                value={searches[searchKey]}
                                title={captitalizeWords(searchKey)}
                                disabled={!safesearch.enabled}
                                onChange={(checked) =>
                                    this.props.toggleSetting('safesearch', { ...safesearch, [searchKey]: checked })
                                }
                            />
                        </div>
                    ))}
                </div>
            </>
        );
    };

    onLanguageChange = (language: string) => {
        i18next.changeLanguage(language);
        setHtmlLangAttr(language);
        const profileName = this.props.dashboard?.name || '';
        if (profileName !== '') {
            this.props.changeLanguage(language);
        }
    };

    onThemeChange = (value: string) => {
        const profileName = this.props.dashboard?.name || '';
        if (profileName !== '') {
            this.props.changeTheme(value);
        } else {
            setUITheme(value);
        }
    };

    onPasswordChange = async (e: React.FormEvent) => {
        e.preventDefault();
        const { t } = this.props;
        const { currentPassword, newPassword, confirmPassword } = this.state;

        if (!currentPassword || !newPassword) {
            this.setState({
                passwordMessage: t('password_required') as string,
                passwordMessageType: 'error',
            });
            return;
        }

        if (newPassword !== confirmPassword) {
            this.setState({
                passwordMessage: t('password_mismatch') as string,
                passwordMessageType: 'error',
            });
            return;
        }

        this.setState({ passwordProcessing: true, passwordMessage: '', passwordMessageType: '' });

        try {
            await apiClient.changePassword({
                current_password: currentPassword,
                new_password: newPassword,
            });
            this.setState({
                passwordMessage: t('password_changed') as string,
                passwordMessageType: 'success',
                currentPassword: '',
                newPassword: '',
                confirmPassword: '',
            });
        } catch (error) {
            this.setState({
                passwordMessage: t('password_incorrect') as string,
                passwordMessageType: 'error',
            });
        }

        this.setState({ passwordProcessing: false });
    };

    onPortSave = async (e: React.FormEvent) => {
        e.preventDefault();
        const { t } = this.props;
        const { port } = this.state;

        const portNum = parseInt(port, 10);
        if (!portNum || portNum < 1 || portNum > 65535) {
            this.setState({
                portMessage: t('port_invalid') as string,
                portMessageType: 'error',
            });
            return;
        }

        this.setState({ portProcessing: true, portMessage: '', portMessageType: '' });

        try {
            await apiClient.changePort({ port: portNum });
            this.setState({
                portMessage: t('port_changed') as string,
                portMessageType: 'success',
            });
            setTimeout(() => {
                const { protocol, hostname } = window.location;
                window.location.href = `${protocol}//${hostname}:${portNum}`;
            }, 1500);
        } catch (error) {
            this.setState({
                portMessage: t('port_change_failed') as string,
                portMessageType: 'error',
            });
        }

        this.setState({ portProcessing: false });
    };

    renderThemeIcons = (): Record<ThemeName, (className: string) => JSX.Element> => ({
        auto: (className) => (
            <svg className={className} viewBox="0 0 24 24" fill="none" aria-hidden="true" focusable="false">
                <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M12 3C16.9706 3 21 7.02944 21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3Z"
                    stroke="currentColor"
                    strokeWidth="1.5"
                />
                <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M12 3V21C16.9706 21 21 16.9706 21 12C21 7.02944 16.9706 3 12 3Z"
                    fill="currentColor"
                    stroke="currentColor"
                    strokeWidth="1.5"
                />
            </svg>
        ),
        dark: (className) => (
            <svg className={className} viewBox="0 0 24 24" fill="none" aria-hidden="true" focusable="false">
                <path
                    d="M3.80737 15.731L3.9895 15.0034C3.71002 14.9335 3.41517 15.0298 3.23088 15.2512C3.0466 15.4727 3.00545 15.7801 3.12501 16.0422L3.80737 15.731ZM14.1926 3.26892L14.3747 2.54137C14.0953 2.47141 13.8004 2.56772 13.6161 2.78917C13.4318 3.01062 13.3907 3.31806 13.5102 3.58018L14.1926 3.26892ZM12 20.2499C8.66479 20.2499 5.79026 18.2708 4.48974 15.4197L3.12501 16.0422C4.66034 19.4081 8.05588 21.7499 12 21.7499V20.2499ZM20.25 11.9999C20.25 16.5563 16.5563 20.2499 12 20.2499V21.7499C17.3848 21.7499 21.75 17.3847 21.75 11.9999H20.25ZM14.0105 3.99647C17.5955 4.89391 20.25 8.13787 20.25 11.9999H21.75C21.75 7.43347 18.6114 3.60193 14.3747 2.54137L14.0105 3.99647ZM13.5102 3.58018C13.9851 4.6211 14.25 5.77857 14.25 6.99995H15.75C15.75 5.5595 15.4371 4.1901 14.875 2.95766L13.5102 3.58018ZM14.25 6.99995C14.25 11.5563 10.5563 15.2499 5.99999 15.2499V16.7499C11.3848 16.7499 15.75 12.3847 15.75 6.99995H14.25ZM5.99999 15.2499C5.30559 15.2499 4.63225 15.1643 3.9895 15.0034L3.62525 16.4585C4.38616 16.649 5.18181 16.7499 5.99999 16.7499V15.2499Z"
                    fill="currentColor"
                />
            </svg>
        ),
        light: (className) => (
            <svg className={className} viewBox="0 0 24 24" fill="none" aria-hidden="true" focusable="false">
                <path
                    d="M12 3.75C16.5563 3.75 20.25 7.44365 20.25 12H21.75C21.75 6.61522 17.3848 2.25 12 2.25V3.75ZM20.25 12C20.25 16.5563 16.5563 20.25 12 20.25V21.75C17.3848 21.75 21.75 17.3848 21.75 12H20.25ZM12 20.25C7.44365 20.25 3.75 16.5563 3.75 12H2.25C2.25 17.3848 6.61522 21.75 12 21.75V20.25ZM3.75 12C3.75 7.44365 7.44365 3.75 12 3.75V2.25C6.61522 2.25 2.25 6.61522 2.25 12H3.75Z"
                    fill="currentColor"
                />
                <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M12 10C10.8954 10 10 10.8954 10 12C10 13.1046 10.8954 14 12 14C13.1046 14 14 13.1046 14 12C13.9987 10.896 13.104 10.0013 12 10Z"
                    fill="currentColor"
                />
            </svg>
        ),
    });

    renderAppearanceCard = () => {
        const { t, dashboard } = this.props;
        const currentTheme = dashboard?.theme || THEMES.auto;
        const languageOptions = Object.keys(LANGUAGES);
        const icons = this.renderThemeIcons();

        const themeContent: Record<ThemeName, { desc: string }> = {
            auto: { desc: t('theme_auto_desc') as string },
            dark: { desc: t('theme_dark_desc') as string },
            light: { desc: t('theme_light_desc') as string },
        };

        return (
            <Card title={t('appearance') as string} bodyType="card-body box-body--settings">
                <div className="form__group">
                    <label className="form__label form__label--bold">{t('theme')}</label>
                    <div className="settings__theme-buttons">
                        {(Object.values(THEMES) as ThemeName[]).map((theme) => (
                            <button
                                key={theme}
                                type="button"
                                className="btn btn-sm btn-secondary settings__theme-button"
                                onClick={() => this.onThemeChange(theme)}
                                title={themeContent[theme].desc}>
                                {icons[theme](
                                    cn('settings__theme-icon', {
                                        'settings__theme-icon--active': currentTheme === theme,
                                    }),
                                )}
                            </button>
                        ))}
                    </div>
                </div>

                <div className="form__group">
                    <label className="form__label form__label--bold">{t('language')}</label>
                    <div className="settings__language-buttons">
                        {languageOptions.map((lang) => {
                            const active = i18next.language === lang;
                            return (
                                <button
                                    key={lang}
                                    type="button"
                                    className={cn('btn btn-sm', {
                                        'btn-secondary': !active,
                                        'btn-primary': active,
                                    })}
                                    onClick={() => this.onLanguageChange(lang)}>
                                    {LANGUAGES[lang]}
                                </button>
                            );
                        })}
                    </div>
                </div>
            </Card>
        );
    };

    renderAccountCard = () => {
        const { t, dashboard } = this.props;
        const {
            currentPassword, newPassword, confirmPassword,
            passwordMessage, passwordMessageType, passwordProcessing,
            port, portMessage, portMessageType, portProcessing,
        } = this.state;
        const profileName = dashboard?.name || '';

        return (
            <Card title={t('system_info') as string} bodyType="card-body box-body--settings">
                <div className="settings__info-row">
                    <span className="settings__info-label">{t('username')}</span>
                    <span className="settings__info-value">{profileName || '—'}</span>
                </div>

                <hr />
                <h6 className="settings__section-title">{t('adguard_home_port')}</h6>
                <form className="settings__password-form" onSubmit={this.onPortSave}>
                    <div className="form-group">
                        <label className="form__label" htmlFor="portInput">
                            Port
                        </label>
                        <input
                            type="number"
                            className="form-control"
                            id="portInput"
                            min="1"
                            max="65535"
                            value={port}
                            onChange={(e) => this.setState({ port: e.target.value })}
                        />
                    </div>
                    <button
                        type="submit"
                        className="btn btn-success btn-standard"
                        disabled={portProcessing}>
                        {t('save_btn')}
                    </button>
                    {portMessage && (
                        <div
                            className={cn('settings__message', {
                                'settings__message--success': portMessageType === 'success',
                                'settings__message--error': portMessageType === 'error',
                            })}>
                            {portMessage}
                        </div>
                    )}
                </form>

                {profileName && (
                    <>
                        <hr />
                        <h6 className="settings__section-title">{t('change_password')}</h6>
                        <form className="settings__password-form" onSubmit={this.onPasswordChange}>
                            <div className="form-group">
                                <label className="form__label" htmlFor="currentPassword">
                                    {t('current_password')}
                                </label>
                                <input
                                    type="password"
                                    className="form-control"
                                    id="currentPassword"
                                    value={currentPassword}
                                    onChange={(e) => this.setState({ currentPassword: e.target.value })}
                                    autoComplete="current-password"
                                />
                            </div>
                            <div className="form-group">
                                <label className="form__label" htmlFor="newPassword">
                                    {t('new_password')}
                                </label>
                                <input
                                    type="password"
                                    className="form-control"
                                    id="newPassword"
                                    value={newPassword}
                                    onChange={(e) => this.setState({ newPassword: e.target.value })}
                                    autoComplete="new-password"
                                />
                            </div>
                            <div className="form-group">
                                <label className="form__label" htmlFor="confirmPassword">
                                    {t('confirm_password')}
                                </label>
                                <input
                                    type="password"
                                    className="form-control"
                                    id="confirmPassword"
                                    value={confirmPassword}
                                    onChange={(e) => this.setState({ confirmPassword: e.target.value })}
                                    autoComplete="new-password"
                                />
                            </div>
                            <button
                                type="submit"
                                className="btn btn-success btn-standard"
                                disabled={passwordProcessing}>
                                {t('save_btn')}
                            </button>
                            {passwordMessage && (
                                <div
                                    className={cn('settings__message', {
                                        'settings__message--success': passwordMessageType === 'success',
                                        'settings__message--error': passwordMessageType === 'error',
                                    })}>
                                    {passwordMessage}
                                </div>
                            )}
                        </form>
                    </>
                )}
            </Card>
        );
    };

    render() {
        const {
            settings,
            setStatsConfig,
            resetStats,
            stats,
            queryLogs,
            setLogsConfig,
            clearLogs,
            filtering,
            setFiltersConfig,
            t,
        } = this.props;

        const isDataReady = !settings.processing && !stats.processingGetConfig && !queryLogs.processingGetConfig;

        return (
            <Fragment>
                <PageTitle title={t('general_settings')} />

                {!isDataReady && <Loading />}

                {isDataReady && (
                    <div className="content">
                        <div className="row">
                            <div className="col-md-12">
                                {this.renderAppearanceCard()}
                            </div>

                            <div className="col-md-12">
                                {this.renderAccountCard()}
                            </div>

                            <div className="col-md-12">
                                <Card bodyType="card-body box-body--settings">
                                    <div className="form">
                                        <FiltersConfig
                                            initialValues={{
                                                interval: filtering.interval,
                                                enabled: filtering.enabled,
                                            }}
                                            processing={filtering.processingSetConfig}
                                            setFiltersConfig={setFiltersConfig}
                                        />
                                        {this.renderSettings(settings.settingsList)}
                                        {this.renderSafeSearch()}
                                    </div>
                                </Card>
                            </div>

                            <div className="col-md-12">
                                <LogsConfig
                                    enabled={queryLogs.enabled}
                                    ignored={queryLogs.ignored}
                                    interval={queryLogs.interval}
                                    customInterval={queryLogs.customInterval}
                                    anonymize_client_ip={queryLogs.anonymize_client_ip}
                                    processing={queryLogs.processingSetConfig}
                                    processingClear={queryLogs.processingClear}
                                    setLogsConfig={setLogsConfig}
                                    clearLogs={clearLogs}
                                />
                            </div>

                            <div className="col-md-12">
                                <StatsConfig
                                    interval={stats.interval}
                                    customInterval={stats.customInterval}
                                    ignored={stats.ignored}
                                    enabled={stats.enabled}
                                    processing={stats.processingSetConfig}
                                    processingReset={stats.processingReset}
                                    setStatsConfig={setStatsConfig}
                                    resetStats={resetStats}
                                />
                            </div>
                        </div>
                    </div>
                )}
            </Fragment>
        );
    }
}

export default withTranslation()(Settings);
