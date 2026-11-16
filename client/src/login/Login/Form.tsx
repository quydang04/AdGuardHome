import React, { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useTranslation, Trans } from 'react-i18next';
import { validateRequiredValue } from '../../helpers/validators';

export type LoginFormValues = {
    username: string;
    password: string;
};

type LoginFormProps = {
    onSubmit: (data: LoginFormValues) => void;
    processing: boolean;
    onToggleForgotPassword: () => void;
    isForgotPasswordVisible: boolean;
};

const EyeIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
        <circle cx="12" cy="12" r="3" />
    </svg>
);

const EyeOffIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
        <line x1="1" y1="1" x2="23" y2="23" />
    </svg>
);

const Form = ({ onSubmit, processing, onToggleForgotPassword, isForgotPasswordVisible }: LoginFormProps) => {
    const { t } = useTranslation();
    const [showPassword, setShowPassword] = useState(false);

    const {
        handleSubmit,
        control,
        formState: { isValid },
    } = useForm<LoginFormValues>({
        mode: 'onChange',
        defaultValues: {
            username: '',
            password: '',
        },
    });

    return (
        <form onSubmit={handleSubmit(onSubmit)} className="login__form-inner">
            <div className="login__field-group">
                <label className="login__label" htmlFor="username">
                    {t('username_label', 'Tên người dùng')}
                </label>
                <Controller
                    name="username"
                    control={control}
                    rules={{ validate: validateRequiredValue }}
                    render={({ field, fieldState }) => (
                        <div className="login__input-wrap">
                            <input
                                {...field}
                                id="username"
                                data-testid="username"
                                type="text"
                                className={`login__input ${fieldState.error ? 'login__input--error' : ''}`}
                                placeholder={t('username_placeholder', 'Nhập tên đăng nhập')}
                                autoComplete="username"
                                autoCapitalize="none"
                                disabled={processing}
                            />
                            {fieldState.error && (
                                <div className="login__field-error">{fieldState.error.message}</div>
                            )}
                        </div>
                    )}
                />
            </div>

            <div className="login__field-group">
                <label className="login__label" htmlFor="password">
                    {t('password_label', 'Mật khẩu')}
                </label>
                <Controller
                    name="password"
                    control={control}
                    rules={{ validate: validateRequiredValue }}
                    render={({ field, fieldState }) => (
                        <div className="login__input-wrap">
                            <div className="login__password-container">
                                <input
                                    {...field}
                                    id="password"
                                    data-testid="password"
                                    type={showPassword ? 'text' : 'password'}
                                    className={`login__input login__input--password ${fieldState.error ? 'login__input--error' : ''}`}
                                    placeholder={t('password_placeholder', 'Nhập mật khẩu')}
                                    autoComplete="current-password"
                                    disabled={processing}
                                />
                                <button
                                    type="button"
                                    className="login__password-toggle"
                                    onClick={() => setShowPassword((prev) => !prev)}
                                    title={showPassword ? t('hide_password', 'Ẩn mật khẩu') : t('show_password', 'Hiện mật khẩu')}
                                    tabIndex={-1}>
                                    {showPassword ? <EyeOffIcon /> : <EyeIcon />}
                                </button>
                            </div>
                            {fieldState.error && (
                                <div className="login__field-error">{fieldState.error.message}</div>
                            )}
                        </div>
                    )}
                />
            </div>

            <div className="login__actions">
                <button
                    data-testid="sign_in"
                    type="submit"
                    className="btn btn-success login__submit-btn"
                    disabled={processing || !isValid}>
                    {t('sign_in', 'Đăng nhập')}
                </button>

                <button
                    type="button"
                    className="login__forgot-btn"
                    onClick={onToggleForgotPassword}>
                    <Trans>forgot_password</Trans>
                </button>
            </div>

            {isForgotPasswordVisible && (
                <div className="login__message">
                    <Trans
                        components={[
                            <a
                                href="https://github.com/AdguardTeam/AdGuardHome/wiki/Configuration#password-reset"
                                key="0"
                                target="_blank"
                                rel="noopener noreferrer">
                                link
                            </a>,
                        ]}>
                        forgot_password_desc
                    </Trans>
                </div>
            )}
        </form>
    );
};

export default Form;
