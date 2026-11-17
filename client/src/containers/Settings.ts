import { connect } from 'react-redux';

import { initSettings, toggleSetting, changeLanguage, changeTheme, getProfile, getDnsStatus } from '../actions';
import { getTlsStatus } from '../actions/encryption';
import { getBlockedServices, updateBlockedServices } from '../actions/services';
import { getStatsConfig, setStatsConfig, resetStats } from '../actions/stats';
import { clearLogs, getLogsConfig, setLogsConfig } from '../actions/queryLogs';
import { getFilteringStatus, setFiltersConfig } from '../actions/filtering';

import Settings from '../components/Settings';

const mapStateToProps = (state: any) => {
    const { settings, services, stats, queryLogs, filtering, dashboard } = state;
    const props = {
        settings,
        services,
        stats,
        queryLogs,
        filtering,
        dashboard,
    };
    return props;
};

const mapDispatchToProps = {
    initSettings,
    toggleSetting,
    getBlockedServices,
    updateBlockedServices,
    getStatsConfig,
    setStatsConfig,
    resetStats,
    clearLogs,
    getLogsConfig,
    setLogsConfig,
    getFilteringStatus,
    setFiltersConfig,
    changeLanguage,
    changeTheme,
    getProfile,
    getDnsStatus,
    getTlsStatus,
};

export default connect(mapStateToProps, mapDispatchToProps)(Settings);
