import { handleActions } from 'redux-actions';

import * as actions from '../actions/youtube';

const youtube = handleActions(
    {
        [actions.getYoutubeConfigRequest.toString()]: (state: any) => ({
            ...state,
            processingGet: true,
        }),
        [actions.getYoutubeConfigFailure.toString()]: (state: any) => ({
            ...state,
            processingGet: false,
        }),
        [actions.getYoutubeConfigSuccess.toString()]: (state, { payload }: any) => ({
            ...state,
            ...payload,
            processingGet: false,
        }),

        [actions.getYoutubeStatusRequest.toString()]: (state: any) => ({
            ...state,
            processingStatus: true,
        }),
        [actions.getYoutubeStatusFailure.toString()]: (state: any) => ({
            ...state,
            processingStatus: false,
        }),
        [actions.getYoutubeStatusSuccess.toString()]: (state, { payload }: any) => ({
            ...state,
            status: payload,
            processingStatus: false,
        }),

        [actions.setYoutubeConfigRequest.toString()]: (state: any) => ({
            ...state,
            processingSet: true,
        }),
        [actions.setYoutubeConfigFailure.toString()]: (state: any) => ({
            ...state,
            processingSet: false,
        }),
        [actions.setYoutubeConfigSuccess.toString()]: (state: any) => ({
            ...state,
            processingSet: false,
        }),
    },
    {
        processingGet: true,
        processingSet: false,
        processingStatus: false,
        enabled: false,
        route_server: '',
        block_ads: true,
        block_tracking: true,
        custom_domains: [],
        ad_domains: [],
        tracking_domains: [],
        rewrite_domains: [],
        status: null,
    },
);

export default youtube;
