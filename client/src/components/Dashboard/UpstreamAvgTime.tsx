import React from 'react';

// @ts-ignore FIXME: update react-table
import ReactTable from 'react-table';
import round from 'lodash/round';
import { withTranslation, Trans } from 'react-i18next';

import { TFunction } from 'i18next';
import Card from '../ui/Card';
import EmptyCardState from '../ui/EmptyCardState';
import DomainCell from './DomainCell';
import { DASHBOARD_TABLES_DEFAULT_PAGE_SIZE, TABLES_MIN_ROWS } from '../../helpers/constants';
import { formatNumber } from '../../helpers/helpers';

interface TimeCellProps {
    value?: string | number;
}

const TimeCell = ({ value }: TimeCellProps) => {
    if (!value) {
        return '–';
    }

    const valueInMilliseconds = formatNumber(round(Number(value) * 1000));

    return (
        <div className="logs__row o-hidden">
            <span className="logs__text logs__text--full" title={valueInMilliseconds.toString()}>
                {valueInMilliseconds}&nbsp;ms
            </span>
        </div>
    );
};

interface UpstreamAvgTimeProps {
    topUpstreamsAvgTime: { name: string; count: number }[];
    refreshButton: React.ReactNode;
    subtitle: string;
    t: TFunction;
}

const UpstreamAvgTime = ({ t, refreshButton, topUpstreamsAvgTime, subtitle }: UpstreamAvgTimeProps) => {
    const hasData = topUpstreamsAvgTime && topUpstreamsAvgTime.length > 0;
    const overallAvg =
        hasData
            ? round(
                  (topUpstreamsAvgTime.reduce((sum, item) => sum + Number(item.count || 0), 0) /
                      topUpstreamsAvgTime.length) *
                      1000,
              )
            : 0;

    return (
        <Card
            title={t('average_upstream_response_time')}
            subtitle={subtitle}
            headerExtra={
                overallAvg > 0 ? (
                    <span className="card-header-badge--muted">{overallAvg} ms</span>
                ) : null
            }
            bodyType="card-table"
            refresh={refreshButton}>
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <ReactTable
                    data={topUpstreamsAvgTime.map(({ name: domain, count }: { name: string; count: number }) => ({
                        domain,
                        count,
                    }))}
                    columns={[
                        {
                            Header: <Trans>upstream</Trans>,
                            accessor: 'domain',
                            Cell: DomainCell,
                        },
                        {
                            Header: <Trans>response_time</Trans>,
                            accessor: 'count',
                            maxWidth: 190,
                            Cell: TimeCell,
                        },
                    ]}
                    showPagination={false}
                    noDataText={t('no_upstreams_data_found')}
                    minRows={TABLES_MIN_ROWS}
                    defaultPageSize={DASHBOARD_TABLES_DEFAULT_PAGE_SIZE}
                    className="-highlight card-table-overflow--limited stats__table"
                />
            )}
        </Card>
    );
};

export default withTranslation()(UpstreamAvgTime);
