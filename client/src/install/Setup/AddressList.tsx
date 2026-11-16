import React from 'react';

import { getIpList, getDnsAddress, getWebAddress } from '../../helpers/helpers';
import { ALL_INTERFACES_IP } from '../../helpers/constants';
import { InstallInterface } from '../../initialState';

interface renderItemProps {
    ip: string;
    port: number;
    isDns: boolean;
}

const renderItem = ({ ip, port, isDns }: renderItemProps) => {
    const webAddress = getWebAddress(ip, port);
    const dnsAddress = getDnsAddress(ip, port);

    return (
        <li key={ip} className="setup__address-item">
            <span className="setup__address-dot" />
            {isDns ? (
                <strong className="setup__address-text">{dnsAddress}</strong>
            ) : (
                <a href={webAddress} target="_blank" rel="noopener noreferrer" className="setup__address-link">
                    {webAddress}
                </a>
            )}
        </li>
    );
};

interface AddressListProps {
    interfaces: InstallInterface[];
    address: string;
    port: number;
    isDns?: boolean;
}

const AddressList = ({ address, interfaces, port, isDns }: AddressListProps) => (
    <ul className="setup__address-list">
        {address === ALL_INTERFACES_IP
            ? getIpList(interfaces).map((ip: any) =>
                  renderItem({
                      ip,
                      port,
                      isDns,
                  }),
              )
            : renderItem({
                  ip: address,
                  port,
                  isDns,
              })}
    </ul>
);

export default AddressList;
