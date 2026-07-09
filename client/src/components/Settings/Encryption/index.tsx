import React, { useCallback, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { debounce } from 'lodash';
import { DEBOUNCE_TIMEOUT, ENCRYPTION_SOURCE } from '../../../helpers/constants';

import { EncryptionFormValues, Form } from './Form';
import { AcmeForm, AcmeFormValues, CHALLENGE_HTTP01 } from './AcmeForm';
import Card from '../../ui/Card';
import PageTitle from '../../ui/PageTitle';
import Loading from '../../ui/Loading';
import { EncryptionData } from '../../../initialState';

type Props = {
    encryption: EncryptionData;
    setTlsConfig: (values: Partial<EncryptionData>) => void;
    validateTlsConfig: (values: Partial<EncryptionData>) => void;
    getAcmeConfig: () => void;
    setAcmeConfig: (values: any) => Promise<void>;
    issueAcmeCertificate: () => void;
};

export const Encryption = ({
    encryption,
    setTlsConfig,
    validateTlsConfig,
    getAcmeConfig,
    setAcmeConfig,
    issueAcmeCertificate,
}: Props) => {
    const { t } = useTranslation();

    useEffect(() => {
        getAcmeConfig();
    }, []);

    const initialValues = useMemo((): EncryptionFormValues => {
        const {
            enabled,
            serve_plain_dns,
            server_name,
            force_https,
            port_https,
            port_dns_over_tls,
            port_dns_over_quic,
            certificate_chain,
            private_key,
            certificate_path,
            private_key_path,
            private_key_saved,
        } = encryption;
        const certificate_source = certificate_chain ? ENCRYPTION_SOURCE.CONTENT : ENCRYPTION_SOURCE.PATH;
        const key_source = private_key || private_key_saved ? ENCRYPTION_SOURCE.CONTENT : ENCRYPTION_SOURCE.PATH;

        return {
            enabled,
            serve_plain_dns,
            server_name,
            force_https,
            port_https,
            port_dns_over_tls,
            port_dns_over_quic,
            certificate_chain,
            private_key,
            certificate_path,
            private_key_path,
            private_key_saved,
            certificate_source,
            key_source,
        };
    }, [encryption]);

    const getSubmitValues = useCallback((values: any) => {
        const { certificate_source, key_source, private_key_saved, ...config } = values;

        if (certificate_source === ENCRYPTION_SOURCE.PATH) {
            config.certificate_chain = '';
        } else {
            config.certificate_path = '';
        }

        if (key_source === ENCRYPTION_SOURCE.PATH) {
            config.private_key = '';
        } else {
            config.private_key_path = '';

            if (private_key_saved) {
                config.private_key = '';
                config.private_key_saved = private_key_saved;
            }
        }

        return config;
    }, []);

    const handleFormSubmit = useCallback(
        (values: any) => {
            const submitValues = getSubmitValues(values);
            setTlsConfig(submitValues);
        },
        [getSubmitValues, setTlsConfig],
    );

    const validateConfig = useCallback((values) => {
        const submitValues = getSubmitValues(values);

        if (submitValues.enabled) {
            validateTlsConfig(submitValues);
        }
    }, []);

    const debouncedConfigValidation = useMemo(() => debounce(validateConfig, DEBOUNCE_TIMEOUT), [validateConfig]);

    const acmeInitialValues = useMemo((): AcmeFormValues => {
        const { acme } = encryption;

        return {
            enabled: acme?.enabled ?? false,
            email: acme?.email ?? '',
            domains: (acme?.domains ?? []).join('\n'),
            challenge: acme?.challenge ?? CHALLENGE_HTTP01,
            cloudflareApiToken: acme?.cloudflare_api_token ?? '',
            autoRenew: acme?.auto_renew ?? true,
            renewBeforeDays: acme?.renew_before_days ?? 14,
        };
    }, [encryption.acme]);

    const getAcmeSubmitValues = useCallback(
        (values: AcmeFormValues) => ({
            enabled: values.enabled,
            email: values.email.trim(),
            domains: values.domains
                .split('\n')
                .map((domain) => domain.trim())
                .filter(Boolean),
            challenge: values.challenge,
            cloudflare_api_token: values.cloudflareApiToken.trim(),
            auto_renew: values.autoRenew,
            renew_before_days: Number(values.renewBeforeDays),
        }),
        [],
    );

    const handleAcmeSave = useCallback(
        (values: AcmeFormValues) => setAcmeConfig(getAcmeSubmitValues(values)),
        [getAcmeSubmitValues, setAcmeConfig],
    );

    const handleAcmeIssue = useCallback(
        async (values: AcmeFormValues) => {
            await setAcmeConfig(getAcmeSubmitValues(values));
            issueAcmeCertificate();
        },
        [getAcmeSubmitValues, setAcmeConfig, issueAcmeCertificate],
    );

    return (
        <div className="encryption">
            <PageTitle title={t('encryption_settings')} />

            {encryption.processing ? (
                <Loading />
            ) : (
                <>
                    <Card
                        title={t('encryption_title')}
                        subtitle={t('encryption_desc')}
                        bodyType="card-body box-body--settings">
                        <Form
                            initialValues={initialValues}
                            onSubmit={handleFormSubmit}
                            debouncedConfigValidation={debouncedConfigValidation}
                            setTlsConfig={setTlsConfig}
                            validateTlsConfig={validateTlsConfig}
                            encryption={encryption}
                        />
                    </Card>

                    <Card
                        title={t('acme_title')}
                        subtitle={t('acme_desc')}
                        bodyType="card-body box-body--settings">
                        <AcmeForm
                            initialValues={acmeInitialValues}
                            processingConfig={encryption.processingAcmeConfig}
                            processingIssue={encryption.processingAcmeIssue}
                            lastIssuedAt={encryption.acme?.last_issued_at ?? null}
                            lastError={encryption.acme?.last_error ?? ''}
                            onSave={handleAcmeSave}
                            onIssue={handleAcmeIssue}
                        />
                    </Card>
                </>
            )}
        </div>
    );
};
