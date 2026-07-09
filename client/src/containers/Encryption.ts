import { connect } from 'react-redux';
import {
    getTlsStatus,
    setTlsConfig,
    validateTlsConfig,
    getAcmeConfig,
    setAcmeConfig,
    issueAcmeCertificate,
} from '../actions/encryption';

import { Encryption } from '../components/Settings/Encryption';

const mapStateToProps = (state: any) => {
    const { encryption } = state;
    const props = {
        encryption,
    };
    return props;
};

const mapDispatchToProps = {
    getTlsStatus,
    setTlsConfig,
    validateTlsConfig,
    getAcmeConfig,
    setAcmeConfig,
    issueAcmeCertificate,
};

export default connect(mapStateToProps, mapDispatchToProps)(Encryption);
