import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import 'mdui/dist/css/mdui.css';
import 'mdui/dist/js/mdui.js';
import '../components/App/index.css';
import '../components/ui/ReactTable.css';
import '../theme/mdui-bridge.css';
import configureStore from '../configureStore';
import reducers from '../reducers/install';
import '../i18n';

import { Setup } from './Setup';
import { InstallState } from '../initialState';

const store = configureStore<InstallState>(reducers, {});

ReactDOM.render(
    <Provider store={store}>
        <Setup />
    </Provider>,
    document.getElementById('root'),
);
