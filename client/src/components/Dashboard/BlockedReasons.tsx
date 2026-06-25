import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import { withTranslation, Trans } from 'react-i18next';

import { TFunction } from 'i18next';
import { shallowEqual, useSelector } from 'react-redux';
import Card from '../ui/Card';

import Cell from '../ui/Cell';

import { getFilterName, getPercent, Filter } from '../../helpers/helpers';
import { DASHBOARD_TABLES_DEFAULT_PAGE_SIZE, STATUS_COLORS, TABLES_MIN_ROWS } from '../../helpers/constants';
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

    return (
        <Card title={t('blocked_reasons')} subtitle={subtitle} bodyType="card-table" refresh={refreshButton}>
            <ReactTable
                data={data}
                columns={[
                    {
                        Header: <Trans>filter_list</Trans>,
                        accessor: 'domain',
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
                        maxWidth: 190,
                        Cell: CountCell(numBlockedFiltering),
                    },
                ]}
                showPagination={false}
                noDataText={t('no_blocked_filter_lists')}
                minRows={TABLES_MIN_ROWS}
                defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                className="-highlight card-table-overflow--limited stats__table"
            />
        </Card>
    );
};

export default withTranslation()(BlockedReasons);
