import { useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useLocation } from 'react-router-dom';

import {
    getLiveStats,
    getLiveStatsFailure,
    getLiveStatsRequest,
    getLiveStatsSuccess,
    getStatsConfig,
} from '../../actions/stats';
import { getFilteringStatus } from '../../actions/filtering';
import { DASHBOARD_REFRESH_INTERVAL_MS, MENU_URLS } from '../../helpers/constants';
import { RootState } from '../../initialState';
import apiClient from '../../api/Api';

const GlobalStatsWatcher = () => {
    const dispatch = useDispatch();
    const location = useLocation();

    const isCoreReady = useSelector<RootState, boolean>(
        (state) => state.dashboard.isCoreRunning && !state.dashboard.processing,
    );
    const statsProcessing = useSelector<RootState, boolean>(
        (state) =>
            state.stats.processingStats || state.stats.processingGetConfig || state.stats.processingLiveStats,
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

        const supportsEventSource = typeof window.EventSource !== 'undefined';

        // Keep stats configuration in sync in background as well.
        dispatch(getStatsConfig());

        if (!supportsEventSource) {
            const fetchStats = () => {
                dispatch(getLiveStats());
                dispatch(getFilteringStatus());
            };

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
        }

        const startFilteringPolling = () => {
            const intervalId = window.setInterval(() => {
                if (document.visibilityState !== 'visible') {
                    return;
                }

                if (!processingRef.current) {
                    dispatch(getFilteringStatus());
                }
            }, DASHBOARD_REFRESH_INTERVAL_MS);

            return () => {
                window.clearInterval(intervalId);
            };
        };

        let eventSource: EventSource | null = null;
        let reconnectTimer: number | null = null;
        let closed = false;

        const connect = () => {
            dispatch(getLiveStatsRequest());

            const streamUrl = apiClient.getLiveStatsStreamUrl();
            const source = new EventSource(streamUrl);

            source.onopen = () => {
                dispatch(getFilteringStatus());
            };

            source.onmessage = (event) => {
                try {
                    const payload = JSON.parse(event.data);
                    dispatch(getLiveStatsSuccess(payload));
                } catch (error) {
                    // eslint-disable-next-line no-console
                    console.error('Failed to parse live stats payload', error);
                    dispatch(getLiveStatsFailure());
                }
            };

            source.onerror = () => {
                dispatch(getLiveStatsFailure());
                source.close();

                if (closed) {
                    return;
                }

                reconnectTimer = window.setTimeout(() => {
                    connect();
                }, DASHBOARD_REFRESH_INTERVAL_MS);
            };

            eventSource = source;
        };

        dispatch(getFilteringStatus());
        connect();

        const stopFilteringPolling = startFilteringPolling();

        return () => {
            closed = true;
            if (reconnectTimer) {
                window.clearTimeout(reconnectTimer);
            }
            if (eventSource) {
                eventSource.close();
            }
            stopFilteringPolling();
        };
    }, [dispatch, isCoreReady, isDashboardRoute]);

    return null;
};

export default GlobalStatsWatcher;
