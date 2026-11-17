import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';

import { TFunction } from 'i18next';
import { shallowEqual, useSelector } from 'react-redux';
import Card from '../ui/Card';
import Cell from '../ui/Cell';
import EmptyCardState from '../ui/EmptyCardState';
import MobileStatsList from './MobileStatsList';

import { formatNumber, getFilterName, getPercent, Filter } from '../../helpers/helpers';
import {
    DASHBOARD_STATS_URLS,
    DASHBOARD_TABLES_DEFAULT_PAGE_SIZE,
    STATUS_COLORS,
    TABLES_MIN_ROWS,
} from '../../helpers/constants';
import { RootState } from '../../initialState';

const CountCell = (totalBlocked: number) =>
    function cell(row: any) {
        const { value } = row;
        const percent = getPercent(totalBlocked, value);

        return <Cell value={value} percent={percent} color={STATUS_COLORS.red} />;
    };

interface BlockedReasonsProps {
    topBlockedFilterLists: { name: string; count: number }[];
    numBlockedFiltering: number;
    refreshButton: React.ReactNode;
    subtitle: string;
    t: TFunction;
}

const BlockedReasons = ({
    t,
    refreshButton,
    topBlockedFilterLists,
    numBlockedFiltering,
    subtitle,
}: BlockedReasonsProps) => {
    const filters = useSelector<RootState, Filter[]>(
        (state) => state.filtering?.filters || [],
        shallowEqual,
    );
    const whitelistFilters = useSelector<RootState, Filter[]>(
        (state) => state.filtering?.whitelistFilters || [],
        shallowEqual,
    );

    const data = topBlockedFilterLists.map(({ name: filterId, count }: { name: string; count: number }) => {
        const id = parseInt(filterId, 10);
        const filterName = isNaN(id)
            ? filterId
            : getFilterName(filters, whitelistFilters, id);

        return {
            domain: filterName,
            count,
        };
    });

    const hasData = data && data.length > 0;

    const mobileItems = hasData
        ? data.map(({ domain, count }) => {
              const percent = getPercent(numBlockedFiltering, count);
              return {
                  key: domain,
                  label: domain,
                  percent,
                  barColor: STATUS_COLORS.red,
                  valueColor: STATUS_COLORS.red,
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
        <Card title={t('blocked_reasons')} subtitle={subtitle} bodyType="card-table" refresh={refreshButton}>
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <>
                    <div className="dashboard-desktop-only">
                        <ReactTable
                            data={data}
                            columns={[
                                {
                                    Header: <Trans>filter_list</Trans>,
                                    accessor: 'domain',
                                    minWidth: 280,
                                    Cell: ({ value }: any) => (
                                        <div className="logs__row o-hidden">
                                            <span className="logs__text" title={value}>
                                                {value}
                                            </span>
                                        </div>
                                    ),
                                },
                                {
                                    Header: <Trans>requests_count</Trans>,
                                    accessor: 'count',
                                    minWidth: 210,
                                    maxWidth: 230,
                                    Cell: CountCell(numBlockedFiltering),
                                },
                            ]}
                            showPagination={false}
                            noDataText={t('no_blocked_filter_lists')}
                            minRows={TABLES_MIN_ROWS}
                            defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                            className="-highlight card-table-overflow--limited stats__table"
                        />
                    </div>

                    <MobileStatsList
                        items={mobileItems}
                        totalCount={data.length}
                        showMoreTo={DASHBOARD_STATS_URLS.blockedReasons}
                    />
                </>
            )}
        </Card>
    );
};

export default withTranslation()(BlockedReasons);
