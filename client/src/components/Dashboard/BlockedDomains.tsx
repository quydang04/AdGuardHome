import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';

import { TFunction } from 'i18next';
import Card from '../ui/Card';
import Cell from '../ui/Cell';
import EmptyCardState from '../ui/EmptyCardState';
import DomainCell from './DomainCell';

import { formatNumber, getPercent } from '../../helpers/helpers';
import { DASHBOARD_TABLES_DEFAULT_PAGE_SIZE, STATUS_COLORS, TABLES_MIN_ROWS } from '../../helpers/constants';

const CountCell = (totalBlocked: any) =>
    function cell(row: any) {
        const { value } = row;
        const percent = getPercent(totalBlocked, value);

        return <Cell value={value} percent={percent} color="#f97316" search={row.original.domain} />;
    };

interface BlockedDomainsProps {
    topBlockedDomains: unknown[];
    blockedFiltering: number;
    replacedSafebrowsing: number;
    replacedSafesearch: number;
    replacedParental: number;
    refreshButton: React.ReactNode;
    subtitle: string;
    t: TFunction;
}

const BlockedDomains = ({
    t,
    refreshButton,
    topBlockedDomains,
    subtitle,
    blockedFiltering,
    replacedSafebrowsing,
    replacedParental,
    replacedSafesearch,
}: BlockedDomainsProps) => {
    const totalBlocked = blockedFiltering + replacedSafebrowsing + replacedParental + replacedSafesearch;
    const hasData = topBlockedDomains && topBlockedDomains.length > 0;

    return (
        <Card
            title={t('top_blocked_domains')}
            subtitle={subtitle}
            headerExtra={
                hasData ? (
                    <span className="card-header-badge--orange">
                        {formatNumber(totalBlocked)} {t('blocked_total', 'blocked total')}
                    </span>
                ) : null
            }
            bodyType="card-table"
            refresh={refreshButton}>
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <ReactTable
                    data={topBlockedDomains.map(({ name: domain, count }: any) => ({
                        domain,
                        count,
                    }))}
                    columns={[
                        {
                            Header: <Trans>domain</Trans>,
                            accessor: 'domain',
                            Cell: DomainCell,
                        },
                        {
                            Header: <Trans>requests_count</Trans>,
                            accessor: 'count',
                            maxWidth: 190,
                            Cell: CountCell(totalBlocked),
                        },
                    ]}
                    showPagination={false}
                    noDataText={t('no_domains_found')}
                    minRows={TABLES_MIN_ROWS}
                    defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                    className="-highlight card-table-overflow--limited stats__table blocked-domains__table"
                />
            )}
        </Card>
    );
};

export default withTranslation()(BlockedDomains);
