import React, { useEffect, useState } from 'react';
import { Trans } from 'react-i18next';
import { useDispatch } from 'react-redux';
import { TOAST_TIMEOUTS } from '../../helpers/constants';

import { removeToast } from '../../actions';

interface ToastProps {
    id: string;
    message: string;
    type: string;
    options?: object;
}

const Toast = ({ id, message, type, options }: ToastProps) => {
    const dispatch = useDispatch();
    const [timerId, setTimerId] = useState(null);

    const clearRemoveToastTimeout = () => clearTimeout(timerId);
    const removeCurrentToast = () => dispatch(removeToast(id));
    const setRemoveToastTimeout = () => {
        const timeout = TOAST_TIMEOUTS[type];
        const timerId = setTimeout(removeCurrentToast, timeout);

        setTimerId(timerId);
    };

    useEffect(() => {
        setRemoveToastTimeout();
    }, []);

    return (
        <div
            className={`toast toast--${type}`}
            onMouseOver={clearRemoveToastTimeout}
            onMouseOut={setRemoveToastTimeout}>
            <span className="toast__icon" aria-hidden="true">
                {type === 'success' && (
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M5 13l4 4L19 7" />
                    </svg>
                )}
                {type !== 'success' && (
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 8v4" />
                        <path d="M12 16h.01" />
                    </svg>
                )}
            </span>
            <p className="toast__content">
                <Trans i18nKey={message} {...options} />
            </p>

            <button className="toast__dismiss" onClick={removeCurrentToast}>
                <svg
                    stroke="#fff"
                    fill="none"
                    width="20"
                    height="20"
                    strokeWidth="2"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg">
                    <path d="m18 6-12 12" />

                    <path d="m6 6 12 12" />
                </svg>
            </button>
        </div>
    );
};

export default Toast;
