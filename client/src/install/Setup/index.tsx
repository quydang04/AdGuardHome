import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import debounce from 'lodash/debounce';

import * as actionCreators from '../../actions/install';
import { getWebAddress, setUITheme } from '../../helpers/helpers';
import { INSTALL_TOTAL_STEPS, ALL_INTERFACES_IP, DEBOUNCE_TIMEOUT, THEMES } from '../../helpers/constants';

import Loading from '../../components/ui/Loading';
import Greeting from './Greeting';
import { ConfigType, DnsConfig, Settings, WebConfig } from './Settings';
import { Devices } from './Devices';
import { Submit } from './Submit';
import { Progress } from './Progress';
import { Auth } from './Auth';
import Toasts from '../../components/Toasts';
import Icons from '../../components/ui/Icons';
import { LanguageDropdown } from '../../components/ui/LanguageDropdown';
import { ThemeDropdown } from '../../components/ui/ThemeDropdown';

import './Setup.css';
import '../../components/ui/Tabler.css';
import { InstallInterface, InstallState } from '../../initialState';

// SVG Icons matching modern UI
const ShieldLogoIcon = () => (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="#67b279" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z" />
    </svg>
);

export const Setup = () => {
    const dispatch = useDispatch();

    const install = useSelector((state: InstallState) => state.install);
    const { processingDefault, step, web, dns, staticIp, interfaces } = install;

    useEffect(() => {
        setUITheme();
    }, []);

    useEffect(() => {
        dispatch(actionCreators.getDefaultAddresses());
    }, []);

    const handleFormSubmit = (values: any) => {
        const config = { ...values };
        delete config.staticIp;

        if (web.port && dns.port) {
            dispatch(
                actionCreators.setAllSettings({
                    web,
                    dns,
                    ...config,
                }),
            );
        }
    };

    const checkConfig = debounce((values) => {
        const { web, dns } = values;

        if (values && web.port && dns.port) {
            dispatch(actionCreators.checkConfig({ web, dns, set_static_ip: false }));
        }
    }, DEBOUNCE_TIMEOUT);

    const handleFix = (web: WebConfig, dns: DnsConfig, set_static_ip: boolean) => {
        dispatch(actionCreators.checkConfig({ web, dns, set_static_ip }));
    };

    const openDashboard = (ip: string, port: number) => {
        let address = getWebAddress(ip, port);
        if (ip === ALL_INTERFACES_IP) {
            address = getWebAddress(window.location.hostname, port);
        }
        window.location.replace(address);
    };

    const handleNextStep = () => {
        if (step < INSTALL_TOTAL_STEPS) {
            dispatch(actionCreators.nextStep());
        }
    };

    const renderPage = (step: number, config: ConfigType, interfaces: InstallInterface[]) => {
        switch (step) {
            case 1:
                return <Greeting />;
            case 2:
                return (
                    <Settings
                        config={config}
                        initialValues={config}
                        interfaces={interfaces}
                        handleSubmit={handleNextStep}
                        validateForm={checkConfig}
                        handleFix={handleFix}
                    />
                );
            case 3:
                return <Auth onAuthSubmit={handleFormSubmit} />;
            case 4:
                return <Devices interfaces={interfaces} dnsConfig={dns} />;
            case 5:
                return <Submit openDashboard={openDashboard} webConfig={web} />;
            default:
                return false;
        }
    };

    if (processingDefault) {
        return <Loading />;
    }

    return (
        <>
            <div className="setup">
                {/* Top Navigation Bar */}
                <header className="setup__header">
                    <div className="setup__brand">
                        <ShieldLogoIcon />
                        <span className="setup__brand-text">adguard home</span>
                    </div>

                    <div className="setup__actions">
                        <ThemeDropdown direction="down" />
                        <LanguageDropdown direction="down" />
                    </div>
                </header>

                {/* Content Area with Centered Card */}
                <main className="setup__content">
                    <div className="setup__container">
                        {step === 1 && (
                            <div className="setup__card-logo">
                                <svg width="56" height="56" viewBox="0 0 24 24" fill="#67b279" xmlns="http://www.w3.org/2000/svg">
                                    <path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z" />
                                </svg>
                            </div>
                        )}
                        {renderPage(step, { web, dns, staticIp }, interfaces)}
                        <Progress step={step} />
                    </div>
                </main>

                <footer className="setup__footer">
                    <span>
                        Copyright &copy; {new Date().getFullYear()}{' '}
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
                </footer>
            </div>

            <Toasts />
            <Icons />
        </>
    );
};
