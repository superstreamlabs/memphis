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
import ClusterConfiguration from './clusterConfiguration';
import { useHistory } from 'react-router-dom';
import pathDomains from '../../router';
import VersionUpgrade from './versionUpgrade';

function Administration({ step }) {
    const [selectedMenuItem, selectMenuItem] = useState(step || 'integrations');
    const [state, dispatch] = useContext(Context);
    const history = useHistory();

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'administration' });
    }, []);

    const getComponent = () => {
        switch (selectedMenuItem) {
            case 'cluster_configuration':
                if (window.location.href.split('/cluster_configuration').length > 1) {
                    return <ClusterConfiguration />;
                } else {
                    history.replace(`${pathDomains.administration}/cluster_configuration`);
                    break;
                }
            case 'integrations':
                if (window.location.href.split('/integrations').length > 1) {
                    return <Integrations />;
                } else {
                    history.replace(`${pathDomains.administration}/integrations`);
                    break;
                }
            case 'version_upgrade':
                if (window.location.href.split('/version_upgrade').length > 1) {
                    return <VersionUpgrade />;
                } else {
                    history.replace(`${pathDomains.administration}/version_upgrade`);
                    break;
                }
            case 'requests':
                if (window.location.href.split('/requests').length > 1) {
                    return <Integrations />;
                } else {
                    history.replace(`${pathDomains.administration}/requests`);
                    break;
                }
            case 'payments':
                if (window.location.href.split('/payments').length > 1) {
                    return <Payments />;
                } else {
                    history.replace(`${pathDomains.administration}/payments`);
                    break;
                }
            default:
                return;
        }
    };
    return (
        <div className="setting-container">
            <div className="menu-container">
                <AccountMenu selectedMenuItem={selectedMenuItem} setMenuItem={(item) => selectMenuItem(item)} />
                <BillingMenu selectedMenuItem={selectedMenuItem} setMenuItem={(item) => selectMenuItem(item)} />
            </div>

            <div className="setting-items">{getComponent()}</div>
        </div>
    );
}
export default Administration;
