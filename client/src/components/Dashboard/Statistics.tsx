import React from 'react';
import { withTranslation, Trans } from 'react-i18next';

import { StatsCard, STATS_CARD_VARIANTS } from './StatsCard';
import { getPercent } from '../../helpers/helpers';
import { RESPONSE_FILTER } from '../../helpers/constants';

interface StatisticsProps {
    dnsQueries: number[];
    blockedFiltering: number[];
    replacedSafebrowsing: number[];
    replacedParental: number[];
    numDnsQueries: number;
    numBlockedFiltering: number;
    numReplacedSafebrowsing: number;
    numReplacedParental: number;
    refreshButton: React.ReactNode;
}

const Statistics = ({
    dnsQueries,
    blockedFiltering,
    replacedSafebrowsing,
    replacedParental,
    numDnsQueries,
    numBlockedFiltering,
    numReplacedSafebrowsing,
    numReplacedParental,
}: StatisticsProps) => (
    <div className="row stats-cards-row">
        <div className="col-6 col-lg-3">
            <StatsCard
                total={numDnsQueries}
                lineData={dnsQueries}
                title={<Trans>dns_query</Trans>}
                variant={STATS_CARD_VARIANTS.QUERIES}
                to="logs"
            />
        </div>

        <div className="col-6 col-lg-3">
            <StatsCard
                total={numBlockedFiltering}
                lineData={blockedFiltering}
                percent={getPercent(numDnsQueries, numBlockedFiltering)}
                title={<Trans>ads_blocked</Trans>}
                variant={STATS_CARD_VARIANTS.ADS}
                to={`logs?response_status=${RESPONSE_FILTER.BLOCKED.QUERY}`}
            />
        </div>

        <div className="col-6 col-lg-3">
            <StatsCard
                total={numReplacedSafebrowsing}
                lineData={replacedSafebrowsing}
                percent={getPercent(numDnsQueries, numReplacedSafebrowsing)}
                title={<Trans>threats_blocked</Trans>}
                variant={STATS_CARD_VARIANTS.THREATS}
                to={`logs?response_status=${RESPONSE_FILTER.BLOCKED_THREATS.QUERY}`}
            />
        </div>

        <div className="col-6 col-lg-3">
            <StatsCard
                total={numReplacedParental}
                lineData={replacedParental}
                percent={getPercent(numDnsQueries, numReplacedParental)}
                title={<Trans>adult_blocked</Trans>}
                variant={STATS_CARD_VARIANTS.ADULT}
                to={`logs?response_status=${RESPONSE_FILTER.BLOCKED_ADULT_WEBSITES.QUERY}`}
            />
        </div>
    </div>
);

export default withTranslation()(Statistics);
