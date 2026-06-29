import { createAction } from 'redux-actions';
import apiClient from '../api/Api';
import { addErrorToast, addSuccessToast } from './toasts';

export const getYoutubeConfigRequest = createAction('GET_YOUTUBE_CONFIG_REQUEST');
export const getYoutubeConfigFailure = createAction('GET_YOUTUBE_CONFIG_FAILURE');
export const getYoutubeConfigSuccess = createAction('GET_YOUTUBE_CONFIG_SUCCESS');

export const getYoutubeConfig = () => async (dispatch: any) => {
    dispatch(getYoutubeConfigRequest());
    try {
        const data = await apiClient.getYoutubeConfig();
        dispatch(getYoutubeConfigSuccess(data));
    } catch (error) {
        dispatch(getYoutubeConfigFailure());
        try {
            dispatch(addErrorToast({ error }));
        } catch (e) {
            console.error('youtube config error toast failed', e);
        }
    }
};

export const getYoutubeStatusRequest = createAction('GET_YOUTUBE_STATUS_REQUEST');
export const getYoutubeStatusFailure = createAction('GET_YOUTUBE_STATUS_FAILURE');
export const getYoutubeStatusSuccess = createAction('GET_YOUTUBE_STATUS_SUCCESS');

export const getYoutubeStatus = () => async (dispatch: any) => {
    dispatch(getYoutubeStatusRequest());
    try {
        const data = await apiClient.getYoutubeStatus();
        dispatch(getYoutubeStatusSuccess(data));
    } catch (error) {
        dispatch(getYoutubeStatusFailure());
    }
};

export const getYoutubeStatsRequest = createAction('GET_YOUTUBE_STATS_REQUEST');
export const getYoutubeStatsFailure = createAction('GET_YOUTUBE_STATS_FAILURE');
export const getYoutubeStatsSuccess = createAction('GET_YOUTUBE_STATS_SUCCESS');

export const getYoutubeStats = () => async (dispatch: any) => {
    dispatch(getYoutubeStatsRequest());
    try {
        const data = await apiClient.getYoutubeStats();
        dispatch(getYoutubeStatsSuccess(data));
    } catch (error) {
        dispatch(getYoutubeStatsFailure());
    }
};

export const setYoutubeConfigRequest = createAction('SET_YOUTUBE_CONFIG_REQUEST');
export const setYoutubeConfigFailure = createAction('SET_YOUTUBE_CONFIG_FAILURE');
export const setYoutubeConfigSuccess = createAction('SET_YOUTUBE_CONFIG_SUCCESS');

export const setYoutubeConfig = (config: any) => async (dispatch: any) => {
    dispatch(setYoutubeConfigRequest());
    try {
        await apiClient.setYoutubeConfig(config);
        dispatch(setYoutubeConfigSuccess());
        dispatch(getYoutubeConfig());
        dispatch(addSuccessToast('youtube_config_saved'));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(setYoutubeConfigFailure());
    }
};
