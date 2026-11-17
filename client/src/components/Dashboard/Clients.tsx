import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import { Link } from 'react-router-dom';
import { Trans, useTranslation } from 'react-i18next';
import { shallowEqual, useSelector } from 'react-redux';
import classNames from 'classnames';

import Card from '../ui/Card';
import Cell from '../ui/Cell';
import EmptyCardState from '../ui/EmptyCardState';
import LogsSearchLink from '../ui/LogsSearchLink';
import { WifiClientIcon, getClientSubtitleText } from './MobileStatsList';
import { useClientBlock } from './useClientBlock';

import { formatNumber, getPercent, sortIp } from '../../helpers/helpers';
import {
    DASHBOARD_MOBILE_PREVIEW_LIMIT,
    DASHBOARD_STATS_URLS,
    DASHBOARD_TABLES_DEFAULT_PAGE_SIZE,
    STATUS_COLORS,
    TABLES_MIN_ROWS,
} from '../../helpers/constants';
import { renderFormattedClientCell } from '../../helpers/renderFormattedClientCell';
import { RootState } from '../../initialState';

const getClientsPercentColor = (percent: number) => {
    if (percent > 50) {
        return STATUS_COLORS.green;
    }
    if (percent > 10) {
        return STATUS_COLORS.yellow;
    }
    return STATUS_COLORS.red;
};

const CountCell = (row: any) => {
    const {
        value,
        original: { ip },
    } = row;

    const numDnsQueries = useSelector<RootState, number>((state) => state.stats.numDnsQueries, shallowEqual);

    const percent = getPercent(numDnsQueries, value);
    const percentColor = getClientsPercentColor(percent);

    return <Cell value={value} percent={percent} color={percentColor} search={ip} />;
};

interface ClientBlockActionProps {
    ip: string;
    disallowed: boolean;
    disallowedRule?: string;
    variant?: 'compact' | 'full';
}

export const ClientBlockActionButton = ({
    ip,
    disallowed,
    disallowedRule = '',
    variant = 'compact',
}: ClientBlockActionProps) => {
    const { t } = useTranslation();
    const { processingSet, isLastRuleInAllowlist, toggleClientStatus } = useClientBlock();

    const lastRuleInAllowlist = isLastRuleInAllowlist(disallowed, disallowedRule, ip);
    const disabled = processingSet || lastRuleInAllowlist;
    const label = disallowed ? t('unblock_client', 'Unblock client') : t('block_client', 'Block client');

    return (
        <button
            type="button"
            className={classNames(
                variant === 'full' ? 'mobile-client-block-btn' : 'desktop-client-block-btn',
                disallowed ? 'client-block-btn--unblock' : 'client-block-btn--block',
            )}
            onClick={(e) => {
                e.stopPropagation();
                toggleClientStatus(ip, disallowed, disallowedRule);
            }}
            disabled={disabled}
            title={lastRuleInAllowlist ? t('last_rule_in_allowlist', { disallowed_rule: disallowedRule || ip }) : label}>
            {label}
        </button>
    );
};

const ClientCell = (row: any) => {
    const { t } = useTranslation();
    const {
        value,
        original: { info },
    } = row;

    const disallowed = Boolean(info?.disallowed);
    const disallowedRule = info?.disallowed_rule || '';
    const isExplicitlyAllowed = !disallowed && Boolean(disallowedRule);

    return (
        <div className="logs__row logs__row--overflow logs__row--column d-flex align-items-center">
            {renderFormattedClientCell(value, info, true)}
            <div className="table__action d-flex align-items-center gap-2">
                {isExplicitlyAllowed && (
                    <span className="client-status-badge client-status-badge--allowed" title={disallowedRule}>
                        {t('client_status_allowed', 'Allowed')}
                    </span>
                )}
                {disallowed && (
                    <span className="client-status-badge client-status-badge--blocked" title={disallowedRule || value}>
                        {t('client_status_blocked', 'Blocked')}
                    </span>
                )}
                <ClientBlockActionButton
                    ip={value}
                    disallowed={disallowed}
                    disallowedRule={disallowedRule}
                    variant="compact"
                />
            </div>
        </div>
    );
};

interface ClientsProps {
    refreshButton: React.ReactNode;
    subtitle: string;
}

const Clients = ({ refreshButton, subtitle }: ClientsProps) => {
    const { t } = useTranslation();

    const topClients = useSelector<RootState, RootState['stats']['topClients']>(
        (state) => state.stats.topClients,
        shallowEqual,
    );

    const numDnsQueries = useSelector<RootState, number>((state) => state.stats.numDnsQueries, shallowEqual);

    const hasData = topClients && topClients.length > 0;
    const previewClients = hasData ? topClients.slice(0, DASHBOARD_MOBILE_PREVIEW_LIMIT) : [];

    return (
        <Card title={t('top_clients')} subtitle={subtitle} bodyType="card-table" refresh={refreshButton}>
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <>
                    <div className="dashboard-desktop-only">
                        <ReactTable
                            data={topClients.map(({ name: ip, count, info, blocked }: any) => ({
                                ip,
                                count,
                                info,
                                blocked,
                            }))}
                            columns={[
                                {
                                    Header: <Trans>client_table_header</Trans>,
                                    accessor: 'ip',
                                    minWidth: 280,
                                    sortMethod: sortIp,
                                    Cell: ClientCell,
                                },
                                {
                                    Header: <Trans>requests_count</Trans>,
                                    accessor: 'count',
                                    minWidth: 210,
                                    maxWidth: 230,
                                    Cell: CountCell,
                                },
                            ]}
                            showPagination={false}
                            noDataText={t('no_clients_found')}
                            minRows={TABLES_MIN_ROWS}
                            defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                            className="-highlight card-table-overflow--limited clients__table"
                            getTrProps={(_state: any, rowInfo: any) => {
                                if (!rowInfo) {
                                    return {};
                                }

                                const disallowed = Boolean(rowInfo.original?.info?.disallowed);
                                return disallowed ? { className: 'logs__row--red' } : {};
                            }}
                        />
                    </div>

                    <div className="dashboard-mobile-only">
                        <div className="mobile-clients-list">
                            {previewClients.map(({ name: ip, count, info }: any) => {
                                const percent = getPercent(numDnsQueries, count);
                                const disallowed = Boolean(info?.disallowed);
                                const disallowedRule = info?.disallowed_rule || '';
                                const subtitleText = getClientSubtitleText(info);

                                return (
                                    <div
                                        key={ip}
                                        className={classNames('mobile-client-card', {
                                            'mobile-client-card--blocked': disallowed,
                                        })}>
                                        <div className="mobile-client-card__title-row">
                                            <WifiClientIcon />
                                            <Link
                                                to={`logs?search="${encodeURIComponent(ip)}"`}
                                                className="mobile-client-card__name">
                                                {ip}
                                            </Link>
                                            {!disallowed && disallowedRule && (
                                                <span className="client-status-badge client-status-badge--allowed ml-auto">
                                                    {t('client_status_allowed', 'Allowed')}
                                                </span>
                                            )}
                                            {disallowed && (
                                                <span className="client-status-badge client-status-badge--blocked ml-auto">
                                                    {t('client_status_blocked', 'Blocked')}
                                                </span>
                                            )}
                                        </div>

                                        <div className="mobile-client-card__subtitle">{subtitleText}</div>

                                        <div className="mobile-client-card__stats">
                                            <span className="mobile-client-card__count">
                                                <LogsSearchLink search={ip}>{formatNumber(count)}</LogsSearchLink>{' '}
                                                <span className="mobile-client-card__percent">({percent}%)</span>
                                            </span>
                                            <div className="mobile-client-card__bar-track">
                                                <div
                                                    className="mobile-client-card__bar-fill"
                                                    style={{
                                                        width: `${Math.min(Math.max(percent, 0), 100)}%`,
                                                        backgroundColor: disallowed ? '#f97316' : '#94a3b8',
                                                    }}
                                                />
                                            </div>
                                        </div>

                                        <ClientBlockActionButton
                                            ip={ip}
                                            disallowed={disallowed}
                                            disallowedRule={disallowedRule}
                                            variant="full"
                                        />
                                    </div>
                                );
                            })}
                        </div>

                        {topClients.length > DASHBOARD_MOBILE_PREVIEW_LIMIT && (
                            <Link to={DASHBOARD_STATS_URLS.topClients} className="dashboard-card-show-more">
                                {t('show_more', 'Show more')}
                            </Link>
                        )}
                    </div>
                </>
            )}
        </Card>
    );
};

export default Clients;
