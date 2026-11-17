import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';

import Card from '../ui/Card';
import Cell from '../ui/Cell';
import EmptyCardState from '../ui/EmptyCardState';
import LogsSearchLink from '../ui/LogsSearchLink';
import DomainCell from './DomainCell';
import MobileStatsList from './MobileStatsList';

import {
    DASHBOARD_STATS_URLS,
    DASHBOARD_TABLES_DEFAULT_PAGE_SIZE,
    STATUS_COLORS,
    TABLES_MIN_ROWS,
} from '../../helpers/constants';

import { formatNumber, getPercent } from '../../helpers/helpers';

const getQueriedPercentColor = (percent: any) => {
    if (percent > 10) {
        return STATUS_COLORS.red;
    }
    if (percent > 5) {
        return STATUS_COLORS.yellow;
    }
    return STATUS_COLORS.green;
};

const countCell = (dnsQueries: any) =>
    function cell(row: any) {
        const { value } = row;
        const percent = getPercent(dnsQueries, value);
        const percentColor = getQueriedPercentColor(percent);

        return <Cell value={value} percent={percent} color={percentColor} search={row.original.domain} />;
    };

interface QueriedDomainsProps {
    topQueriedDomains: { name: string; count: number }[];
    dnsQueries: number;
    refreshButton: React.ReactNode;
    subtitle: string;
    t: (...args: unknown[]) => string;
}

const QueriedDomains = ({ t, refreshButton, topQueriedDomains, subtitle, dnsQueries }: QueriedDomainsProps) => {
    const hasData = topQueriedDomains && topQueriedDomains.length > 0;

    const mobileItems = hasData
        ? topQueriedDomains.map(({ name: domain, count }) => {
              const percent = getPercent(dnsQueries, count);
              return {
                  key: domain,
                  label: domain,
                  isTrackerDomain: true,
                  percent,
                  barColor: '#94a3b8',
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
        <Card title={t('stats_query_domain')} subtitle={subtitle} bodyType="card-table" refresh={refreshButton}>
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <>
                    <div className="dashboard-desktop-only">
                        <ReactTable
                            data={topQueriedDomains.map(({ name: domain, count }: any) => ({
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
                                    Cell: countCell(dnsQueries),
                                },
                            ]}
                            showPagination={false}
                            noDataText={t('no_domains_found')}
                            minRows={TABLES_MIN_ROWS}
                            defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                            className="-highlight card-table-overflow--limited stats__table"
                        />
                    </div>

                    <MobileStatsList
                        items={mobileItems}
                        totalCount={topQueriedDomains.length}
                        showMoreTo={DASHBOARD_STATS_URLS.topQueriedDomains}
                    />
                </>
            )}
        </Card>
    );
};

export default withTranslation()(QueriedDomains);
