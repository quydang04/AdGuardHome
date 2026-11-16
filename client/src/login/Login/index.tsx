import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useTranslation, Trans } from 'react-i18next';

import * as actionCreators from '../../actions/login';
import Toasts from '../../components/Toasts';
import Icons from '../../components/ui/Icons';
import { LanguageDropdown } from '../../components/ui/LanguageDropdown';
import { ThemeDropdown } from '../../components/ui/ThemeDropdown';
import { setUITheme } from '../../helpers/helpers';
import Form, { LoginFormValues } from './Form';

import './Login.css';
import { LoginState } from '../../initialState';

// SVG Icons matching modern UI
const ShieldLogoIcon = () => (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="#67b279" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z" />
    </svg>
);

export const Login = () => {
    const dispatch = useDispatch();
    const { t } = useTranslation();
    const { processingLogin } = useSelector((state: LoginState) => state.login || { processingLogin: false });
    const [isForgotPasswordVisible, setIsForgotPasswordVisible] = useState(false);

    useEffect(() => {
        setUITheme();
    }, []);

    const handleSubmit = ({ username: name, password }: LoginFormValues) => {
        dispatch(actionCreators.processLogin({ name, password }));
    };

    const toggleForgotPassword = () => {
        setIsForgotPasswordVisible((prev) => !prev);
    };

    return (
        <div className="login">
            {/* Top Navigation Bar */}
            <header className="login__header">
                <div className="login__brand">
                    <ShieldLogoIcon />
                    <span className="login__brand-text">adguard home</span>
                </div>

                <div className="login__actions">
                    <ThemeDropdown direction="down" />
                    <LanguageDropdown direction="down" />
                </div>
            </header>

            {/* Main Center Area */}
            <main className="login__content">
                <div className="login__card">
                    <h1 className="login__title">{t('sign_in', 'Đăng nhập')}</h1>

                    <Form
                        onSubmit={handleSubmit}
                        processing={processingLogin}
                        onToggleForgotPassword={toggleForgotPassword}
                        isForgotPasswordVisible={isForgotPasswordVisible}
                    />
                </div>
            </main>

            <Toasts />
            <Icons />
        </div>
    );
};
