import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';

import { TFunction } from 'i18next';
import Card from '../ui/Card';
import Cell from '../ui/Cell';
import EmptyCardState from '../ui/EmptyCardState';
import DomainCell from './DomainCell';
import MobileStatsList from './MobileStatsList';

import { formatNumber, getPercent } from '../../helpers/helpers';
import {
    DASHBOARD_STATS_URLS,
    DASHBOARD_TABLES_DEFAULT_PAGE_SIZE,
    STATUS_COLORS,
    TABLES_MIN_ROWS,
} from '../../helpers/constants';

const CountCell = (totalBlocked: any) =>
    function cell(row: any) {
        const { value } = row;
        const percent = getPercent(totalBlocked, value);

        return <Cell value={value} percent={percent} color={STATUS_COLORS.green} />;
    };

export const getTotalUpstreamRequests = (stats: { count: number }[]) => {
    let total = 0;
    stats.forEach(({ count }) => {
        total += count;
    });

    return total;
};

interface UpstreamResponsesProps {
    topUpstreamsResponses: { name: string; count: number }[];
    refreshButton: React.ReactNode;
    subtitle: string;
    t: TFunction;
}

const UpstreamResponses = ({ t, refreshButton, topUpstreamsResponses, subtitle }: UpstreamResponsesProps) => {
    const hasData = topUpstreamsResponses && topUpstreamsResponses.length > 0;
    const totalUpstreamRequests = hasData ? getTotalUpstreamRequests(topUpstreamsResponses) : 0;

    const mobileItems = hasData
        ? topUpstreamsResponses.map(({ name: domain, count }) => {
              const percent = getPercent(totalUpstreamRequests, count);
              return {
                  key: domain,
                  label: domain,
                  percent,
                  barColor: '#94a3b8',
                  valueNode: (
                      <>
                          <span>{formatNumber(count)}</span>{' '}
                          <span className="mobile-stat-item__percent">({percent}%)</span>
                      </>
                  ),
              };
          })
        : [];

    return (
        <Card title={t('top_upstreams')} subtitle={subtitle} bodyType="card-table" refresh={refreshButton}>
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <>
                    <div className="dashboard-desktop-only">
                        <ReactTable
                            data={topUpstreamsResponses.map(({ name: domain, count }: { name: string; count: number }) => ({
                                domain,
                                count,
                            }))}
                            columns={[
                                {
                                    Header: <Trans>upstream</Trans>,
                                    accessor: 'domain',
                                    minWidth: 280,
                                    Cell: DomainCell,
                                },
                                {
                                    Header: <Trans>requests_count</Trans>,
                                    accessor: 'count',
                                    minWidth: 210,
                                    maxWidth: 230,
                                    Cell: CountCell(totalUpstreamRequests),
                                },
                            ]}
                            showPagination={false}
                            noDataText={t('no_upstreams_data_found')}
                            minRows={TABLES_MIN_ROWS}
                            defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                            className="-highlight card-table-overflow--limited stats__table"
                        />
                    </div>

                    <MobileStatsList
                        items={mobileItems}
                        totalCount={topUpstreamsResponses.length}
                        showMoreTo={DASHBOARD_STATS_URLS.topUpstreams}
                    />
                </>
            )}
        </Card>
    );
};

export default withTranslation()(UpstreamResponses);
