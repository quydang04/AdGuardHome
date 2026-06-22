import { connect } from 'react-redux';
import { getStats } from '../actions/stats';
import SystemOverview from '../components/SystemOverview';
import { RootState } from '../initialState';

const mapStateToProps = (state: RootState) => ({
    systemInfo: state.stats.systemInfo,
    processing: state.stats.processingStats,
});

const mapDispatchToProps = { getStats };

export default connect(mapStateToProps, mapDispatchToProps)(SystemOverview);
