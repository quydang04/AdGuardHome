import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';

import { TFunction } from 'i18next';
import Card from '../ui/Card';
import Cell from '../ui/Cell';
import EmptyCardState from '../ui/EmptyCardState';
import LogsSearchLink from '../ui/LogsSearchLink';
import DomainCell from './DomainCell';
import MobileStatsList from './MobileStatsList';

import { formatNumber, getPercent } from '../../helpers/helpers';
import {
    DASHBOARD_STATS_URLS,
    DASHBOARD_TABLES_DEFAULT_PAGE_SIZE,
    TABLES_MIN_ROWS,
} from '../../helpers/constants';

const CountCell = (totalBlocked: any) =>
    function cell(row: any) {
        const { value } = row;
        const percent = getPercent(totalBlocked, value);

        return <Cell value={value} percent={percent} color="#f97316" search={row.original.domain} />;
    };

interface BlockedDomainsProps {
    topBlockedDomains: { name: string; count: number }[];
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

    const mobileItems = hasData
        ? topBlockedDomains.map(({ name: domain, count }) => {
              const percent = getPercent(totalBlocked, count);
              return {
                  key: domain,
                  label: domain,
                  isTrackerDomain: true,
                  percent,
                  barColor: '#f97316',
                  valueColor: '#f97316',
                  valueNode: (
                      <>
                          <LogsSearchLink search={domain}>{formatNumber(count)}</LogsSearchLink>{' '}
                          <span className="mobile-stat-item__percent">({percent}%)</span>
                      </>
                  ),
              };
          })
        : [];

    return (
        <Card
            title={t('top_blocked_domains')}
            subtitle={subtitle}
            headerExtra={
                hasData ? (
                    <span className="card-header-badge--orange">
                        {formatNumber(totalBlocked)}
                        <span className="d-none d-md-inline"> {t('blocked_total', 'blocked total')}</span>
                    </span>
                ) : null
            }
            bodyType="card-table"
            refresh={refreshButton}>
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <>
                    <div className="dashboard-desktop-only">
                        <ReactTable
                            data={topBlockedDomains.map(({ name: domain, count }: any) => ({
                                domain,
                                count,
                            }))}
                            columns={[
                                {
                                    Header: <Trans>domain</Trans>,
                                    accessor: 'domain',
                                    minWidth: 280,
                                    Cell: DomainCell,
                                },
                                {
                                    Header: <Trans>requests_count</Trans>,
                                    accessor: 'count',
                                    minWidth: 210,
                                    maxWidth: 230,
                                    Cell: CountCell(totalBlocked),
                                },
                            ]}
                            showPagination={false}
                            noDataText={t('no_domains_found')}
                            minRows={TABLES_MIN_ROWS}
                            defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                            className="-highlight card-table-overflow--limited stats__table blocked-domains__table"
                        />
                    </div>

                    <MobileStatsList
                        items={mobileItems}
                        totalCount={topBlockedDomains.length}
                        showMoreTo={DASHBOARD_STATS_URLS.topBlockedDomains}
                    />
                </>
            )}
        </Card>
    );
};

export default withTranslation()(BlockedDomains);
