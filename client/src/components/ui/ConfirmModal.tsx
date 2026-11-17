import React, { useCallback, useEffect, useRef, useState } from 'react';
import ReactModal from 'react-modal';
import { useTranslation } from 'react-i18next';
import classNames from 'classnames';

import './Modal.css';
import './ConfirmModal.css';

export interface ConfirmModalOptions {
    title?: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    variant?: 'danger' | 'success' | 'primary';
}

interface PendingConfirmRequest {
    options: ConfirmModalOptions;
    resolve: (confirmed: boolean) => void;
}

type ConfirmListener = (request: PendingConfirmRequest) => void;

let activeListener: ConfirmListener | null = null;

/**
 * Opens the app's styled ReactModal confirmation dialog and resolves to true
 * when confirmed or false when cancelled/dismissed.
 */
export const confirmDialog = (input: string | ConfirmModalOptions): Promise<boolean> => {
    const options: ConfirmModalOptions =
        typeof input === 'string' ? { message: input } : input;

    if (!activeListener) {
        // Fallback for isolated test environments where ConfirmModalHost isn't mounted.
        return Promise.resolve(window.confirm(options.message));
    }

    return new Promise<boolean>((resolve) => {
        activeListener?.({ options, resolve });
    });
};

export const ConfirmModalHost = () => {
    const { t } = useTranslation();
    const [queue, setQueue] = useState<PendingConfirmRequest[]>([]);
    const currentRef = useRef<PendingConfirmRequest | null>(null);

    const current = queue[0] || null;
    currentRef.current = current;

    useEffect(() => {
        if (typeof document !== 'undefined' && document.getElementById('root')) {
            ReactModal.setAppElement('#root');
        }

        const listener: ConfirmListener = (request) => {
            setQueue((prev) => [...prev, request]);
        };

        activeListener = listener;

        return () => {
            if (activeListener === listener) {
                activeListener = null;
            }
        };
    }, []);

    const handleResolve = useCallback((confirmed: boolean) => {
        const active = currentRef.current;
        if (!active) {
            return;
        }
        currentRef.current = null;
        setQueue((prev) => prev.slice(1));
        active.resolve(confirmed);
    }, []);

    const options = current?.options;
    const variant = options?.variant || 'danger';

    return (
        <ReactModal
            className="Modal__Bootstrap modal-dialog modal-dialog-centered confirm-modal__dialog"
            closeTimeoutMS={0}
            isOpen={Boolean(options)}
            onRequestClose={() => handleResolve(false)}>
            {options && (
                <div className="modal-content confirm-modal">
                    <div className="modal-header">
                        <h4 className="modal-title">
                            {options.title || t('confirm_action', 'Confirm')}
                        </h4>
                        <button
                            type="button"
                            className="close"
                            onClick={() => handleResolve(false)}
                            aria-label="Close">
                            <span className="sr-only">Close</span>
                        </button>
                    </div>

                    <div className="modal-body">
                        <p className="confirm-modal__message mb-0">{options.message}</p>
                    </div>

                    <div className="modal-footer">
                        <button
                            type="button"
                            className="btn btn-secondary"
                            onClick={() => handleResolve(false)}
                            data-testid="confirm_modal_cancel">
                            {options.cancelText || t('cancel_btn')}
                        </button>
                        <button
                            type="button"
                            className={classNames('btn', {
                                'btn-danger': variant === 'danger',
                                'btn-success': variant === 'success',
                                'btn-primary': variant === 'primary',
                            })}
                            onClick={() => handleResolve(true)}
                            data-testid="confirm_modal_ok">
                            {options.confirmText || t('confirm_btn', 'Confirm')}
                        </button>
                    </div>
                </div>
            )}
        </ReactModal>
    );
};

export default ConfirmModalHost;
