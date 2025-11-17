import { useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation } from 'react-router-dom';

import { getStats, getStatsConfig } from '../../actions/stats';
import { getFilteringStatus } from '../../actions/filtering';
import { DASHBOARD_REFRESH_INTERVAL_MS, MENU_URLS } from '../../helpers/constants';
import { RootState } from '../../initialState';

const GlobalStatsWatcher = () => {
    const dispatch = useDispatch();
    const location = useLocation();

    const isCoreReady = useSelector<RootState, boolean>(
        (state) => state.dashboard.isCoreRunning && !state.dashboard.processing,
    );
    const statsProcessing = useSelector<RootState, boolean>(
        (state) => state.stats.processingStats || state.stats.processingGetConfig,
    );
    const filteringProcessing = useSelector<RootState, boolean>((state) => state.filtering.processingFilters);

    const processingRef = useRef(false);

    useEffect(() => {
        processingRef.current = statsProcessing || filteringProcessing;
    }, [statsProcessing, filteringProcessing]);

    const isDashboardRoute = location.pathname === MENU_URLS.root || location.pathname === '';

    useEffect(() => {
        if (!isCoreReady || isDashboardRoute) {
            return undefined;
        }

        const fetchStats = () => {
            dispatch(getStats());
            dispatch(getFilteringStatus());
        };

        // Keep stats configuration in sync in background as well.
        dispatch(getStatsConfig());
        fetchStats();

        const intervalId = window.setInterval(() => {
            if (document.visibilityState !== 'visible') {
                return;
            }

            if (!processingRef.current) {
                fetchStats();
            }
        }, DASHBOARD_REFRESH_INTERVAL_MS);

        return () => {
            window.clearInterval(intervalId);
        };
    }, [dispatch, isCoreReady, isDashboardRoute]);

    return null;
};

export default GlobalStatsWatcher;
