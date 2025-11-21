import React from 'react';
import ReactDOM from 'react-dom';
import { Provider } from 'react-redux';

import 'mdui/dist/css/mdui.css';
import 'mdui/dist/js/mdui.js';
import '../components/App/index.css';
import '../components/ui/ReactTable.css';
import '../theme/mdui-bridge.css';
import configureStore from '../configureStore';
import reducers from '../reducers/login';
import '../i18n';

import { Login } from './Login';
import { LoginState } from '../initialState';

const store = configureStore<LoginState>(reducers, {});

ReactDOM.render(
    <Provider store={store}>
        <Login />
    </Provider>,
    document.getElementById('root'),
);
