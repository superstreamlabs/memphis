// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useContext, useState, useEffect } from 'react';

import { Context } from '../../hooks/store';
import Integrations from './integrations';
import AccountMenu from './accountMenu';
import BillingMenu from './billing/billingMenu';
import Payments from './billing/payments';
import Requests from './billing/requests';
import ClusterConfiguration from './clusterConfiguration';
import SoftwareUpates from './softwareUpdates';
import { useHistory } from 'react-router-dom';
import pathDomains from '../../router';
import VersionUpgrade from './versionUpgrade';
import { isCloud } from '../../services/valueConvertor';
import { useLocation } from 'react-router-dom/cjs/react-router-dom.min';

function Administration({ step }) {
    const [selectedMenuItem, setSelectedMenuItem] = useState('integrations');
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const location = useLocation();

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'administration' });
    }, []);

    useEffect(() => {
        const pathSegments = location.pathname.split('/');
        const selected = pathSegments[pathSegments.length - 1];
        setSelectedMenuItem(selected);
    }, [location]);

    const handleMenuItemChange = (menuItem) => {
        setSelectedMenuItem(menuItem);
        history.push(`${pathDomains.administration}/${menuItem}`);
    };

    const renderSelectedComponent = () => {
        switch (selectedMenuItem) {
            case 'integrations':
                return <Integrations />;
            case 'cluster_configuration':
                return <ClusterConfiguration />;
            case 'version_upgrade':
                if (!isCloud()) {
                    // return <VersionUpgrade />;
                    return <SoftwareUpates />;
                }
                break;
            case 'usage':
                return <Requests />;
            case 'payments':
                return <Payments />;
            default:
                return null; // Handle invalid selections
        }
    };

    return (
        <div className="setting-container">
            <div className="menu-container">
                <AccountMenu selectedMenuItem={selectedMenuItem} setMenuItem={handleMenuItemChange} />
                {isCloud() && <BillingMenu selectedMenuItem={selectedMenuItem} setMenuItem={handleMenuItemChange} />}
            </div>
            {selectedMenuItem === 'version_upgrade' ? <>{renderSelectedComponent()}</> : <div className="setting-items">{renderSelectedComponent()}</div>}
        </div>
    );
}
export default Administration;
