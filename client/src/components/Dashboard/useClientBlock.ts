import { useCallback } from 'react';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';
import { useTranslation } from 'react-i18next';

import { getAccessList, toggleClientBlock } from '../../actions/access';
import { getStats } from '../../actions/stats';
import { confirmDialog } from '../ui/ConfirmModal';
import { RootState } from '../../initialState';

/**
 * Shared client block/unblock action hook used by Dashboard Top Clients
 * (desktop table, mobile card list) and the Show More Top Clients page.
 */
export const useClientBlock = () => {
    const dispatch = useDispatch();
    const { t } = useTranslation();

    const processingSet = useSelector<RootState, boolean>(
        (state) => Boolean(state.access.processingSet),
    );

    const allowedClients = useSelector<RootState, string>(
        (state) => state.access.allowed_clients || '',
        shallowEqual,
    );

    const isLastRuleInAllowlist = useCallback(
        (disallowed: boolean, disallowedRule?: string, ip?: string) => {
            if (disallowed || !allowedClients) {
                return false;
            }
            const trimmedAllowed = allowedClients.trim();
            return Boolean(trimmedAllowed) && trimmedAllowed === (disallowedRule || ip || '');
        },
        [allowedClients],
    );

    const toggleClientStatus = useCallback(
        async (ip: string, disallowed: boolean, disallowedRule?: string) => {
            const rule = disallowedRule || '';
            let confirmMessage: string;

            if (disallowed) {
                confirmMessage = t('client_confirm_unblock', { ip: rule || ip });
            } else {
                confirmMessage = `${t('adg_will_drop_dns_queries')} ${t('client_confirm_block', { ip })}`;
                if (allowedClients.length > 0) {
                    confirmMessage = confirmMessage.concat(`\n\n${t('filter_allowlist', { disallowed_rule: rule || ip })}`);
                }
            }

            const confirmed = await confirmDialog({
                title: disallowed ? t('unblock_client', 'Unblock client') : t('block_client', 'Block client'),
                message: confirmMessage,
                confirmText: disallowed ? t('unblock_client', 'Unblock client') : t('block_client', 'Block client'),
                variant: disallowed ? 'success' : 'danger',
            });

            if (confirmed) {
                await dispatch(toggleClientBlock(ip, disallowed, rule) as any);
                await Promise.all([
                    dispatch(getAccessList() as any),
                    dispatch(getStats() as any),
                ]);
            }
        },
        [allowedClients, dispatch, t],
    );

    return {
        processingSet,
        isLastRuleInAllowlist,
        toggleClientStatus,
    };
};

export default useClientBlock;
