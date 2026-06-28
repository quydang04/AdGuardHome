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
        dispatch(addErrorToast({ error }));
        dispatch(getYoutubeConfigFailure());
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
