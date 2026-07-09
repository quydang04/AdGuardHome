import { handleActions } from 'redux-actions';

import * as actions from '../actions/encryption';
import { ENCRYPTION_SOURCE } from '../helpers/constants';

const encryption = handleActions(
    {
        [actions.getTlsStatusRequest.toString()]: (state: any) => ({
            ...state,
            processing: true,
        }),
        [actions.getTlsStatusFailure.toString()]: (state: any) => ({
            ...state,
            processing: false,
        }),
        [actions.getTlsStatusSuccess.toString()]: (state: any, { payload }: any) => {
            const newState = {
                ...state,
                ...payload,
                /* TODO: handle property delete on api refactor */
                server_name: payload.server_name || '',
                processing: false,
            };
            return newState;
        },

        [actions.setTlsConfigRequest.toString()]: (state: any) => ({
            ...state,
            processingConfig: true,
        }),
        [actions.setTlsConfigFailure.toString()]: (state: any) => ({
            ...state,
            processingConfig: false,
        }),
        [actions.setTlsConfigSuccess.toString()]: (state: any, { payload }: any) => {
            const newState = {
                ...state,
                ...payload,
                server_name: payload.server_name || '',
                processingConfig: false,
            };
            return newState;
        },

        [actions.validateTlsConfigRequest.toString()]: (state: any) => ({
            ...state,
            processingValidate: true,
        }),
        [actions.validateTlsConfigFailure.toString()]: (state: any) => ({
            ...state,
            processingValidate: false,
        }),
        [actions.validateTlsConfigSuccess.toString()]: (state: any, { payload }: any) => {
            const {
                issuer = '',
                key_type = '',
                not_after = '',
                not_before = '',
                subject = '',
                warning_validation = '',
                dns_names = '',
                ...values
            } = payload;

            const newState = {
                ...state,
                ...values,
                issuer,
                key_type,
                not_after,
                not_before,
                subject,
                warning_validation,
                dns_names,
                server_name: payload.server_name || '',
                processingValidate: false,
            };
            return newState;
        },

        [actions.getAcmeConfigRequest.toString()]: (state: any) => ({
            ...state,
            processingAcme: true,
        }),
        [actions.getAcmeConfigFailure.toString()]: (state: any) => ({
            ...state,
            processingAcme: false,
        }),
        [actions.getAcmeConfigSuccess.toString()]: (state: any, { payload }: any) => ({
            ...state,
            acme: payload,
            processingAcme: false,
        }),

        [actions.setAcmeConfigRequest.toString()]: (state: any) => ({
            ...state,
            processingAcmeConfig: true,
        }),
        [actions.setAcmeConfigFailure.toString()]: (state: any) => ({
            ...state,
            processingAcmeConfig: false,
        }),
        [actions.setAcmeConfigSuccess.toString()]: (state: any, { payload }: any) => ({
            ...state,
            acme: payload,
            processingAcmeConfig: false,
        }),

        [actions.issueAcmeCertificateRequest.toString()]: (state: any) => ({
            ...state,
            processingAcmeIssue: true,
        }),
        [actions.issueAcmeCertificateFailure.toString()]: (state: any) => ({
            ...state,
            processingAcmeIssue: false,
        }),
        [actions.issueAcmeCertificateSuccess.toString()]: (state: any, { payload }: any) => ({
            ...state,
            ...payload.status,
            certificate_chain: payload.certificate_chain,
            private_key: payload.private_key,
            certificate_source: ENCRYPTION_SOURCE.CONTENT,
            key_source: ENCRYPTION_SOURCE.CONTENT,
            certificate_path: '',
            private_key_path: '',
            processingAcmeIssue: false,
        }),
    },
    {
        processing: true,
        processingConfig: false,
        processingValidate: false,
        processingAcme: false,
        processingAcmeConfig: false,
        processingAcmeIssue: false,
        enabled: false,
        serve_plain_dns: false,
        dns_names: null,
        force_https: false,
        issuer: '',
        key_type: '',
        not_after: '',
        not_before: '',
        port_dns_over_tls: '',
        port_https: '',
        subject: '',
        valid_chain: false,
        valid_key: false,
        valid_cert: false,
        valid_pair: false,
        status_cert: '',
        status_key: '',
        certificate_chain: '',
        private_key: '',
        server_name: '',
        warning_validation: '',
        certificate_path: '',
        private_key_path: '',
        acme: {
            enabled: false,
            email: '',
            domains: [],
            challenge: 'http-01',
            cloudflare_api_token: '',
            dns_resolvers: [],
            auto_renew: true,
            renew_before_days: 14,
            last_issued_at: null,
            last_error: '',
        },
    },
);

export default encryption;
