import { handleActions } from 'redux-actions';

import {
    getTelegramConfigRequest,
    getTelegramConfigFailure,
    getTelegramConfigSuccess,
    setTelegramConfigRequest,
    setTelegramConfigFailure,
    setTelegramConfigSuccess,
    sendTelegramTestRequest,
    sendTelegramTestFailure,
    sendTelegramTestSuccess,
} from '../actions/notifications';
import { NotificationsState, initialState } from '../initialState';

const defaultState: NotificationsState = initialState.notifications;

export default handleActions<NotificationsState, any>(
    {
        [getTelegramConfigRequest.toString()]: (state) => ({
            ...state,
            processingGet: true,
        }),
        [getTelegramConfigSuccess.toString()]: (state, { payload }) => ({
            ...state,
            processingGet: false,
            telegram: {
                ...state.telegram,
                ...payload,
            },
        }),
        [getTelegramConfigFailure.toString()]: (state) => ({
            ...state,
            processingGet: false,
        }),
        [setTelegramConfigRequest.toString()]: (state) => ({
            ...state,
            processingSave: true,
        }),
        [setTelegramConfigSuccess.toString()]: (state, { payload }) => ({
            ...state,
            processingSave: false,
            telegram: {
                ...state.telegram,
                ...payload,
            },
        }),
        [setTelegramConfigFailure.toString()]: (state) => ({
            ...state,
            processingSave: false,
        }),
        [sendTelegramTestRequest.toString()]: (state) => ({
            ...state,
            processingTest: true,
        }),
        [sendTelegramTestSuccess.toString()]: (state) => ({
            ...state,
            processingTest: false,
        }),
        [sendTelegramTestFailure.toString()]: (state) => ({
            ...state,
            processingTest: false,
        }),
    },
    defaultState,
);
