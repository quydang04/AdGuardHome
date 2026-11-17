import { connect } from 'react-redux';

import { toggleProtection, getClients } from '../actions';
import { getStats, getStatsConfig, setStatsConfig, setStatsViewInterval } from '../actions/stats';
import { getAccessList } from '../actions/access';
import { getFilteringStatus } from '../actions/filtering';

import Dashboard from '../components/Dashboard';
import { RootState } from '../initialState';

const mapStateToProps = (state: RootState) => {
    const { dashboard, stats, access } = state;
    const props = { dashboard, stats, access };
    return props;
};

type DispatchProps = {
    toggleProtection: (...args: unknown[]) => unknown;
    getClients: (...args: unknown[]) => unknown;
    getStats: (...args: unknown[]) => unknown;
    getStatsConfig: (...args: unknown[]) => unknown;
    setStatsConfig: (...args: unknown[]) => unknown;
    setStatsViewInterval: (interval: number | null) => unknown;
    getAccessList: () => (dispatch: any) => void;
    getFilteringStatus: (...args: unknown[]) => unknown;
};

const mapDispatchToProps: DispatchProps = {
    toggleProtection,
    getClients,
    getStats,
    getStatsConfig,
    setStatsConfig,
    setStatsViewInterval,
    getAccessList,
    getFilteringStatus,
};

export default connect(mapStateToProps, mapDispatchToProps)(Dashboard);
