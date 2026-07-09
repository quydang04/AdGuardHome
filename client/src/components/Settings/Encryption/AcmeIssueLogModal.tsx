import React, { useEffect, useRef, useState } from 'react';
import ReactModal from 'react-modal';
import { useDispatch } from 'react-redux';
import { useTranslation } from 'react-i18next';

import apiClient from '../../../api/Api';
import {
    issueAcmeCertificateRequest,
    issueAcmeCertificateSuccess,
    issueAcmeCertificateFailure,
} from '../../../actions/encryption';
import { addErrorToast, addSuccessToast } from '../../../actions/toasts';

import '../../ui/Modal.css';
import './AcmeIssueLogModal.css';

ReactModal.setAppElement('#root');

type AcmeLogLine = {
    time: string;
    level: 'info' | 'error' | 'success';
    message: string;
};

type AcmeDoneEvent = {
    success: boolean;
    error?: string;
    status?: Record<string, unknown>;
    certificate_chain?: string;
    private_key?: string;
};

type Props = {
    isOpen: boolean;
    onClose: () => void;
};

export const AcmeIssueLogModal = ({ isOpen, onClose }: Props) => {
    const dispatch = useDispatch();
    const { t } = useTranslation();

    const [lines, setLines] = useState<AcmeLogLine[]>([]);
    const [finished, setFinished] = useState(false);
    const [success, setSuccess] = useState<boolean | null>(null);

    const logEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!isOpen) {
            return undefined;
        }

        setLines([]);
        setFinished(false);
        setSuccess(null);

        let cancelled = false;
        let isDone = false;
        let es: EventSource | null = null;

        const finish = (ok: boolean) => {
            isDone = true;
            if (cancelled) return;
            setFinished(true);
            setSuccess(ok);
        };

        (async () => {
            dispatch(issueAcmeCertificateRequest());

            try {
                await apiClient.startAcmeIssue();
            } catch (error) {
                if (cancelled) return;
                dispatch(addErrorToast({ error }));
                dispatch(issueAcmeCertificateFailure());
                finish(false);
                return;
            }

            if (cancelled) return;

            es = new EventSource(apiClient.getAcmeIssueStreamUrl());

            es.addEventListener('line', (event: MessageEvent) => {
                try {
                    const data: AcmeLogLine = JSON.parse(event.data);
                    setLines((prev) => [...prev, data]);
                } catch {
                    // Ignore malformed lines.
                }
            });

            es.addEventListener('done', (event: MessageEvent) => {
                let data: AcmeDoneEvent | null = null;
                try {
                    data = JSON.parse(event.data);
                } catch {
                    data = null;
                }

                if (data?.success) {
                    dispatch(
                        issueAcmeCertificateSuccess({
                            status: data.status,
                            certificate_chain: data.certificate_chain ? atob(data.certificate_chain) : '',
                            private_key: data.private_key ? atob(data.private_key) : '',
                        }),
                    );
                    dispatch(addSuccessToast('acme_issue_success'));
                    finish(true);
                } else {
                    dispatch(issueAcmeCertificateFailure());
                    dispatch(addErrorToast({ error: data?.error || t('acme_issue_log_unknown_error') }));
                    finish(false);
                }

                es?.close();
            });

            es.onerror = () => {
                es?.close();
                if (!cancelled && !isDone) {
                    dispatch(issueAcmeCertificateFailure());
                    dispatch(addErrorToast({ error: t('acme_issue_log_connection_lost') }));
                    finish(false);
                }
            };
        })();

        return () => {
            cancelled = true;
            es?.close();
        };
    }, [isOpen]);

    useEffect(() => {
        logEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
    }, [lines]);

    const handleClose = () => {
        if (finished) {
            onClose();
        }
    };

    return (
        <ReactModal
            className="Modal__Bootstrap modal-dialog modal-dialog-centered modal-dialog--acme-log"
            closeTimeoutMS={0}
            isOpen={isOpen}
            onRequestClose={handleClose}
            shouldCloseOnOverlayClick={finished}>
            <div className="modal-content">
                <div className="modal-header">
                    <h4 className="modal-title">{t('acme_issue_log_title')}</h4>
                    {finished && (
                        <button type="button" className="close" onClick={onClose}>
                            <span className="sr-only">Close</span>
                        </button>
                    )}
                </div>

                <div className="modal-body">
                    <div className="acme-log">
                        {lines.length === 0 && !finished && (
                            <div className="acme-log__line acme-log__line--info">
                                {t('acme_issue_log_starting')}
                            </div>
                        )}
                        {lines.map((line, idx) => (
                            <div
                                // eslint-disable-next-line react/no-array-index-key
                                key={idx}
                                className={`acme-log__line acme-log__line--${line.level}`}>
                                <span className="acme-log__time">
                                    {new Date(line.time).toLocaleTimeString()}
                                </span>{' '}
                                {line.message}
                            </div>
                        ))}
                        {finished && (
                            <div className={`acme-log__line acme-log__line--${success ? 'success' : 'error'}`}>
                                {success ? t('acme_issue_log_success') : t('acme_issue_log_failed')}
                            </div>
                        )}
                        <div ref={logEndRef} />
                    </div>
                </div>

                <div className="modal-footer">
                    <button
                        type="button"
                        className="btn btn-secondary"
                        onClick={onClose}
                        disabled={!finished}>
                        {finished ? t('acme_issue_log_close') : t('acme_issue_log_in_progress')}
                    </button>
                </div>
            </div>
        </ReactModal>
    );
};
