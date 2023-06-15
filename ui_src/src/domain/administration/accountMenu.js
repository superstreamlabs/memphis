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

import React, { useContext } from 'react';
import { Context } from '../../hooks/store';
import ClusterConfColor from '../../assets/images/setting/clusterConfColor.svg';
import ClusterConfGray from '../../assets/images/setting/clusterConfGray.svg';
import IntegrationColor from '../../assets/images/setting/integrationColor.svg';
import IntegrationGray from '../../assets/images/setting/integrationGray.svg';
import versionUpgradeColor from '../../assets/images/setting/versionUpgradeColor.svg';
import versionUpgradeGray from '../../assets/images/setting/versionUpgradeGray.svg';

function AccountMenu({ selectedMenuItem, setMenuItem }) {
    const [state, dispatch] = useContext(Context);

    return (
        <>
            <p className="header">Account</p>
            <p className="memphis-label">Modify environment configuration</p>
            <div className="side-menu">
                <div className={selectedMenuItem === 'integrations' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('integrations')}>
                    <img src={selectedMenuItem === 'integrations' ? IntegrationColor : IntegrationGray} alt="notifications" />
                    Integrations
                </div>
                <div className={selectedMenuItem === 'cluster_configuration' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('cluster_configuration')}>
                    <img src={selectedMenuItem === 'cluster_configuration' ? ClusterConfColor : ClusterConfGray} alt="clusterConfiguration" />
                    Cluster configuration
                </div>
                <div className={selectedMenuItem === 'version_upgrade' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('version_upgrade')}>
                    <img src={selectedMenuItem === 'version_upgrade' ? versionUpgradeColor : versionUpgradeGray} alt="versionUpgrade" />
                    Software Update
                    {!state.isLatest && <div className="update-available">Update available</div>}
                </div>
            </div>
        </>
    );
}
export default AccountMenu;
