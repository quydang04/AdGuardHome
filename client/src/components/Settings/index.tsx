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
import CustomDropdownSelect from '../ui/CustomDropdownSelect';
import { confirmDialog } from '../ui/ConfirmModal';

import { getObjectKeysSorted, captitalizeWords, setHtmlLangAttr, setUITheme } from '../../helpers/helpers';
import { THEMES, MIN_PASSWORD_LENGTH, MAX_PASSWORD_LENGTH } from '../../helpers/constants';
import { utf8StringLength } from '../../helpers/validators';
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
    getProfile?: () => unknown;
    getDnsStatus?: () => unknown;
    getTlsStatus?: () => unknown;
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
    showCurrentPassword: boolean;
    showNewPassword: boolean;
    showConfirmPassword: boolean;
    passwordMessage: string;
    passwordMessageType: 'success' | 'error' | '';
    passwordProcessing: boolean;
    port: string;
    portMessage: string;
    portMessageType: 'success' | 'error' | '';
    portProcessing: boolean;
    username: string;
    usernameMessage: string;
    usernameMessageType: 'success' | 'error' | '';
    usernameProcessing: boolean;
    importExportMessage: string;
    importExportMessageType: 'success' | 'error' | '';
    importExportProcessing: boolean;
}

class Settings extends Component<SettingsProps, SettingsState> {
    state: SettingsState = {
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
        showCurrentPassword: false,
        showNewPassword: false,
        showConfirmPassword: false,
        passwordMessage: '',
        passwordMessageType: '',
        passwordProcessing: false,
        port: '',
        portMessage: '',
        portMessageType: '',
        portProcessing: false,
        username: '',
        usernameMessage: '',
        usernameMessageType: '',
        usernameProcessing: false,
        importExportMessage: '',
        importExportMessageType: '',
        importExportProcessing: false,
    };

    private hasScrolledToHash = false;

    componentDidMount() {
        this.props.initSettings(SETTINGS);
        this.props.getStatsConfig();
        this.props.getLogsConfig();
        this.props.getFilteringStatus();

        const httpPort = this.props.dashboard?.httpPort;
        if (httpPort) {
            this.setState({ port: String(httpPort) });
        }

        const profileName = this.props.dashboard?.name || '';
        if (profileName) {
            this.setState({ username: profileName });
        }

        this.scrollToHash();
    }

    componentDidUpdate(prevProps: SettingsProps) {
        const httpPort = this.props.dashboard?.httpPort;
        const prevPort = prevProps.dashboard?.httpPort;
        if (httpPort && httpPort !== prevPort && this.state.port === '') {
            this.setState({ port: String(httpPort) });
        }

        const profileName = this.props.dashboard?.name || '';
        const prevName = prevProps.dashboard?.name || '';
        if (profileName && profileName !== prevName && this.state.username === '') {
            this.setState({ username: profileName });
        }

        if (!this.hasScrolledToHash) {
            this.scrollToHash();
        }
    }

    scrollToHash() {
        const { hash } = window.location;
        if (!hash) {
            return;
        }

        const id = hash.replace('#', '');
        const el = document.getElementById(id);
        if (el) {
            this.hasScrolledToHash = true;
            setTimeout(() => {
                el.scrollIntoView({ behavior: 'smooth', block: 'start' });
            }, 100);
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

    getBackendErrorDetail = (error: unknown): string => {
        const message = error instanceof Error ? error.message : String(error ?? '');
        const parts = message.split(' | ');
        if (parts.length >= 2) {
            return parts[1].trim();
        }
        return message.trim();
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

        const passwordBytes = utf8StringLength(newPassword);
        if (passwordBytes < MIN_PASSWORD_LENGTH || passwordBytes > MAX_PASSWORD_LENGTH) {
            this.setState({
                passwordMessage: t('form_error_password_length', {
                    min: MIN_PASSWORD_LENGTH,
                    max: MAX_PASSWORD_LENGTH,
                }) as string,
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

        if (newPassword === currentPassword) {
            this.setState({
                passwordMessage: t('password_same_as_current') as string,
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
                showCurrentPassword: false,
                showNewPassword: false,
                showConfirmPassword: false,
            });
        } catch (error) {
            const detail = this.getBackendErrorDetail(error);
            let message = t('password_change_failed') as string;
            if (detail.includes('current password is incorrect')) {
                message = t('password_incorrect') as string;
            } else if (detail) {
                message = `${message}: ${detail}`;
            }

            this.setState({
                passwordMessage: message,
                passwordMessageType: 'error',
            });
        }

        this.setState({ passwordProcessing: false });
    };

    onPortSave = async (e: React.FormEvent) => {
        e.preventDefault();
        const { t, dashboard } = this.props;
        const trimmedPort = this.state.port.trim();

        if (!/^\d+$/.test(trimmedPort)) {
            this.setState({
                portMessage: t('port_invalid') as string,
                portMessageType: 'error',
            });
            return;
        }

        const portNum = Number(trimmedPort);
        if (portNum < 1 || portNum > 65535) {
            this.setState({
                portMessage: t('port_invalid') as string,
                portMessageType: 'error',
            });
            return;
        }

        const currentPort = dashboard?.httpPort;
        if (currentPort && portNum === currentPort) {
            this.setState({
                portMessage: t('port_unchanged') as string,
                portMessageType: 'success',
            });
            return;
        }

        this.setState({ portProcessing: true, portMessage: '', portMessageType: '' });

        try {
            await apiClient.changePort({ port: portNum });
            const isPlainHTTP = window.location.protocol === 'http:';
            this.setState({
                portMessage: t(isPlainHTTP ? 'port_changed' : 'port_saved_no_redirect') as string,
                portMessageType: 'success',
            });
            if (isPlainHTTP) {
                setTimeout(() => {
                    const { hostname, pathname, search, hash } = window.location;
                    const portSuffix = portNum === 80 ? '' : `:${portNum}`;
                    window.location.href = `http://${hostname}${portSuffix}${pathname}${search}${hash}`;
                }, 1500);
            }
        } catch (error) {
            const detail = this.getBackendErrorDetail(error);
            let message = t('port_change_failed') as string;
            if (detail.includes('port conflicts with HTTPS port')) {
                message = t('port_conflicts_https') as string;
            } else if (detail) {
                message = `${message}: ${detail}`;
            }

            this.setState({
                portMessage: message,
                portMessageType: 'error',
            });
        }

        this.setState({ portProcessing: false });
    };

    onUsernameSave = async (e: React.FormEvent) => {
        e.preventDefault();
        const { t, dashboard, getProfile } = this.props;
        const trimmedUsername = this.state.username.trim();

        if (!trimmedUsername) {
            this.setState({
                usernameMessage: t('username_required') as string,
                usernameMessageType: 'error',
            });
            return;
        }

        if (dashboard?.name && trimmedUsername === dashboard.name) {
            this.setState({
                username: trimmedUsername,
                usernameMessage: t('username_unchanged') as string,
                usernameMessageType: 'success',
            });
            return;
        }

        this.setState({ usernameProcessing: true, usernameMessage: '', usernameMessageType: '' });

        try {
            await apiClient.changeUsername({ new_username: trimmedUsername });
            this.setState({
                username: trimmedUsername,
                usernameMessage: t('username_changed') as string,
                usernameMessageType: 'success',
            });
            if (getProfile) {
                getProfile();
            }
        } catch (error) {
            const detail = this.getBackendErrorDetail(error);
            let message = t('username_change_failed') as string;
            if (detail.includes('username already exists')) {
                message = t('username_exists') as string;
            } else if (detail) {
                message = `${message}: ${detail}`;
            }

            this.setState({
                usernameMessage: message,
                usernameMessageType: 'error',
            });
        }

        this.setState({ usernameProcessing: false });
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

        const themeLabels: Record<ThemeName, string> = {
            auto: t('theme_auto') as string,
            dark: t('theme_dark') as string,
            light: t('theme_light') as string,
        };

        return (
            <Card
                title={t('appearance') as string}
                type="settings__appearance-card"
                bodyType="card-body box-body--settings">
                <div className="settings__password-form">
                    <div className="form-group">
                        <label className="form__label" htmlFor="themeSelect">
                            {t('theme')}
                        </label>
                        <div className="settings__theme-select">
                            {(Object.values(THEMES) as ThemeName[]).map((theme) => (
                                <button
                                    key={theme}
                                    type="button"
                                    className={cn('settings__theme-option', {
                                        'settings__theme-option--active': currentTheme === theme,
                                    })}
                                    onClick={() => this.onThemeChange(theme)}>
                                    {icons[theme]('settings__theme-option-icon')}
                                    <span>{themeLabels[theme]}</span>
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="form-group">
                        <label className="form__label" htmlFor="languageSelect">
                            {t('language')}
                        </label>
                        <CustomDropdownSelect
                            id="languageSelect"
                            value={
                                languageOptions.includes(i18next.language)
                                    ? i18next.language
                                    : languageOptions.find((l) => i18next.language?.startsWith(l)) || languageOptions[0]
                            }
                            options={languageOptions.map((lang) => ({
                                value: lang,
                                label: LANGUAGES[lang],
                            }))}
                            onChange={(val) => this.onLanguageChange(val)}
                        />
                    </div>
                </div>
            </Card>
        );
    };

    renderPasswordField = (
        id: string,
        label: string,
        value: string,
        visible: boolean,
        autoComplete: string,
        onChange: (val: string) => void,
        onToggleVisible: () => void,
    ) => {
        const { t } = this.props;
        return (
            <div className="form-group">
                <label className="form__label" htmlFor={id}>
                    {label}
                </label>
                <div className="settings__password-input-wrap">
                    <input
                        type={visible ? 'text' : 'password'}
                        className="form-control"
                        id={id}
                        value={value}
                        onChange={(e) => onChange(e.target.value)}
                        autoComplete={autoComplete}
                    />
                    <button
                        type="button"
                        className="settings__password-toggle"
                        onClick={onToggleVisible}
                        title={t(visible ? 'hide_password' : 'show_password') as string}
                        aria-label={t(visible ? 'hide_password' : 'show_password') as string}>
                        {visible ? (
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                                <line x1="1" y1="1" x2="23" y2="23" />
                            </svg>
                        ) : (
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                                <circle cx="12" cy="12" r="3" />
                            </svg>
                        )}
                    </button>
                </div>
            </div>
        );
    };

    renderAccountCard = () => {
        const { t, dashboard } = this.props;
        const {
            currentPassword, newPassword, confirmPassword,
            showCurrentPassword, showNewPassword, showConfirmPassword,
            passwordMessage, passwordMessageType, passwordProcessing,
            port, portMessage, portMessageType, portProcessing,
            username, usernameMessage, usernameMessageType, usernameProcessing,
        } = this.state;
        const profileName = dashboard?.name || '';

        return (
            <Card title={t('account_and_port') as string} bodyType="card-body box-body--settings">
                {profileName && (
                    <>
                        <h6 className="settings__section-title">{t('username')}</h6>
                        <form className="settings__password-form" onSubmit={this.onUsernameSave}>
                            <div className="form-group">
                                <label className="form__label" htmlFor="usernameInput">
                                    {t('username_new')}
                                </label>
                                <input
                                    type="text"
                                    className="form-control"
                                    id="usernameInput"
                                    value={username}
                                    onChange={(e) =>
                                        this.setState({
                                            username: e.target.value,
                                            usernameMessage: '',
                                            usernameMessageType: '',
                                        })
                                    }
                                    placeholder={t('username_placeholder') as string}
                                />
                            </div>
                            <button
                                type="submit"
                                className="btn btn-success btn-standard"
                                disabled={usernameProcessing}>
                                {t('save_btn')}
                            </button>
                            {usernameMessage && (
                                <div
                                    className={cn('settings__message', {
                                        'settings__message--success': usernameMessageType === 'success',
                                        'settings__message--error': usernameMessageType === 'error',
                                    })}>
                                    {usernameMessage}
                                </div>
                            )}
                        </form>
                        <hr className="settings-divider" />
                    </>
                )}

                <h6 className="settings__section-title">{t('adguard_home_port')}</h6>
                <form className="settings__password-form" onSubmit={this.onPortSave}>
                    <div className="form-group">
                        <label className="form__label" htmlFor="portInput">
                            {t('http_port_label')}
                        </label>
                        <input
                            type="number"
                            className="form-control"
                            id="portInput"
                            min="1"
                            max="65535"
                            value={port}
                            onChange={(e) =>
                                this.setState({
                                    port: e.target.value,
                                    portMessage: '',
                                    portMessageType: '',
                                })
                            }
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
                            style={{ whiteSpace: 'pre-line' }}
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
                        <hr className="settings-divider" />
                        <h6 className="settings__section-title">{t('change_password')}</h6>
                        <form className="settings__password-form" onSubmit={this.onPasswordChange}>
                            {this.renderPasswordField(
                                'currentPassword',
                                t('current_password') as string,
                                currentPassword,
                                showCurrentPassword,
                                'current-password',
                                (val) =>
                                    this.setState({
                                        currentPassword: val,
                                        passwordMessage: '',
                                        passwordMessageType: '',
                                    }),
                                () => this.setState((prev) => ({ showCurrentPassword: !prev.showCurrentPassword })),
                            )}
                            {this.renderPasswordField(
                                'newPassword',
                                t('new_password') as string,
                                newPassword,
                                showNewPassword,
                                'new-password',
                                (val) =>
                                    this.setState({
                                        newPassword: val,
                                        passwordMessage: '',
                                        passwordMessageType: '',
                                    }),
                                () => this.setState((prev) => ({ showNewPassword: !prev.showNewPassword })),
                            )}
                            {this.renderPasswordField(
                                'confirmPassword',
                                t('confirm_password') as string,
                                confirmPassword,
                                showConfirmPassword,
                                'new-password',
                                (val) =>
                                    this.setState({
                                        confirmPassword: val,
                                        passwordMessage: '',
                                        passwordMessageType: '',
                                    }),
                                () => this.setState((prev) => ({ showConfirmPassword: !prev.showConfirmPassword })),
                            )}
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

    onExportSettings = async () => {
        const { t } = this.props;
        this.setState({ importExportProcessing: true, importExportMessage: '', importExportMessageType: '' });

        try {
            const data = await apiClient.exportSettings();
            const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            const date = new Date().toISOString().slice(0, 10);
            a.download = `adguardhome-settings-${date}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            this.setState({
                importExportMessage: t('settings_exported') as string,
                importExportMessageType: 'success',
            });
        } catch (error) {
            this.setState({
                importExportMessage: t('settings_export_error') as string,
                importExportMessageType: 'error',
            });
        }

        this.setState({ importExportProcessing: false });
    };

    onImportSettings = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const { t } = this.props;
        const file = e.target.files?.[0];
        if (!file) {
            return;
        }

        // Reset file input so the same file can be selected again.
        e.target.value = '';

        try {
            const rawText = await file.text();
            const text = rawText.replace(/^\uFEFF/, '').trim();
            if (!text) {
                throw new SyntaxError('empty file');
            }
            const data = JSON.parse(text);
            if (!data || typeof data !== 'object' || Array.isArray(data) || Object.keys(data).length === 0) {
                throw new SyntaxError('empty object');
            }

            const confirmed = await confirmDialog({
                title: t('import_settings') as string,
                message: t('settings_import_confirm') as string,
                confirmText: t('import_settings') as string,
                variant: 'danger',
            });
            if (!confirmed) {
                return;
            }

            this.setState({ importExportProcessing: true, importExportMessage: '', importExportMessageType: '' });

            const res: any = await apiClient.importSettings(data);
            const warnings: string[] = Array.isArray(res?.warnings) ? res.warnings : [];

            this.props.initSettings(SETTINGS);
            this.props.getStatsConfig();
            this.props.getLogsConfig();
            this.props.getFilteringStatus();
            this.props.getProfile?.();
            this.props.getDnsStatus?.();
            this.props.getTlsStatus?.();

            this.setState({
                importExportMessage: warnings.length > 0
                    ? `${t('settings_imported_warnings') as string}\n- ${warnings.join('\n- ')}`
                    : t('settings_imported') as string,
                importExportMessageType: 'success',
            });
        } catch (error) {
            this.setState({
                importExportMessage: error instanceof SyntaxError
                    ? t('settings_import_invalid_file') as string
                    : t('settings_import_error') as string,
                importExportMessageType: 'error',
            });
        }

        this.setState({ importExportProcessing: false });
    };

    renderImportExportCard = () => {
        const { t } = this.props;
        const { importExportMessage, importExportMessageType, importExportProcessing } = this.state;

        return (
            <Card title={t('import_export_settings') as string} bodyType="card-body box-body--settings">
                <p className="form__desc form__desc--top">{t('import_export_settings_desc')}</p>
                <div className="settings__import-export">
                    <div className="settings__import-export-section mb-0">
                        <h6 className="settings__section-title">{t('export_settings')}</h6>
                        <p className="form__desc form__desc--top">{t('export_settings_desc')}</p>
                        <button
                            type="button"
                            className="btn btn-success btn-standard"
                            onClick={this.onExportSettings}
                            disabled={importExportProcessing}>
                            {t('export_settings')}
                        </button>
                    </div>
                    <hr className="settings-divider" />
                    <div className="settings__import-export-section">
                        <h6 className="settings__section-title">{t('import_settings')}</h6>
                        <p className="form__desc form__desc--top">{t('import_settings_desc')}</p>
                        <label className="btn btn-outline-secondary btn-standard settings__import-btn" htmlFor="importFile">
                            {importExportProcessing ? '...' : t('select_file')}
                            <input
                                type="file"
                                id="importFile"
                                accept=".json"
                                onChange={this.onImportSettings}
                                disabled={importExportProcessing}
                                className="settings__import-input"
                            />
                        </label>
                    </div>
                    {importExportMessage && (
                        <div
                            style={{ whiteSpace: 'pre-line' }}
                            className={cn('settings__message', {
                                'settings__message--success': importExportMessageType === 'success',
                                'settings__message--error': importExportMessageType === 'error',
                            })}>
                            {importExportMessage}
                        </div>
                    )}
                </div>
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

                            <div className="col-md-12">
                                {this.renderImportExportCard()}
                            </div>
                        </div>
                    </div>
                )}
            </Fragment>
        );
    }
}

export default withTranslation()(Settings);
