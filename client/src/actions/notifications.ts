import { createAction } from 'redux-actions';

import apiClient from '../api/Api';
import { addErrorToast, addSuccessToast } from './toasts';

export const getTelegramConfigRequest = createAction('GET_TELEGRAM_CONFIG_REQUEST');
export const getTelegramConfigFailure = createAction('GET_TELEGRAM_CONFIG_FAILURE');
export const getTelegramConfigSuccess = createAction('GET_TELEGRAM_CONFIG_SUCCESS');

export const setTelegramConfigRequest = createAction('SET_TELEGRAM_CONFIG_REQUEST');
export const setTelegramConfigFailure = createAction('SET_TELEGRAM_CONFIG_FAILURE');
export const setTelegramConfigSuccess = createAction('SET_TELEGRAM_CONFIG_SUCCESS');

export const sendTelegramTestRequest = createAction('SEND_TELEGRAM_TEST_REQUEST');
export const sendTelegramTestFailure = createAction('SEND_TELEGRAM_TEST_FAILURE');
export const sendTelegramTestSuccess = createAction('SEND_TELEGRAM_TEST_SUCCESS');

export const getTelegramConfig = () => async (dispatch: any) => {
    dispatch(getTelegramConfigRequest());
    try {
        const data = await apiClient.getTelegramConfig();
        dispatch(getTelegramConfigSuccess(data));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(getTelegramConfigFailure());
    }
};

export const setTelegramConfig = (config: any) => async (dispatch: any) => {
    dispatch(setTelegramConfigRequest());
    try {
        await apiClient.setTelegramConfig(config);
        dispatch(setTelegramConfigSuccess(config));
        dispatch(addSuccessToast('telegram_config_saved'));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(setTelegramConfigFailure());
    }
};

export const sendTelegramTest = (message: string) => async (dispatch: any) => {
    dispatch(sendTelegramTestRequest());
    try {
        await apiClient.sendTelegramTest({ message });
        dispatch(sendTelegramTestSuccess());
        dispatch(addSuccessToast('telegram_test_success'));
    } catch (error) {
        dispatch(addErrorToast({ error }));
        dispatch(sendTelegramTestFailure());
    }
};
