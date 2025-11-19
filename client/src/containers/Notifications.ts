import { connect } from 'react-redux';

import Notifications from '../components/Notifications';
import { getTelegramConfig, setTelegramConfig, sendTelegramTest } from '../actions/notifications';
import { RootState } from '../initialState';

const mapStateToProps = (state: RootState) => ({
    notifications: state.notifications,
});

const mapDispatchToProps = {
    getTelegramConfig,
    setTelegramConfig,
    sendTelegramTest,
};

export default connect(mapStateToProps, mapDispatchToProps)(Notifications);
