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

import { ReactComponent as VersionUpgradeColorIcon } from '../../assets/images/setting/versionUpgradeColor.svg';
import { ReactComponent as VersionUpgradeGrayIcon } from '../../assets/images/setting/versionUpgradeGray.svg';
import { ReactComponent as ClusterConfColorIcon } from '../../assets/images/setting/clusterConfColor.svg';
import { ReactComponent as IntegrationColorIcon } from '../../assets/images/setting/integrationColor.svg';
import { ReactComponent as ClusterConfGrayIcon } from '../../assets/images/setting/clusterConfGray.svg';
import { ReactComponent as IntegrationGrayIcon } from '../../assets/images/setting/integrationGray.svg';
import PersonOutlinedIcon from '@material-ui/icons/PersonOutlined';

import { isCloud } from '../../services/valueConvertor';
import { Context } from '../../hooks/store';

function AccountMenu({ selectedMenuItem, setMenuItem }) {
    const [state, dispatch] = useContext(Context);

    return (
        <>
            <p className="header">Administration</p>
            <p className="sub-header">Modify environment configuration</p>
            <div className="side-menu administration">
                {!isCloud() && (
                    <>
                        <div className={selectedMenuItem === 'version_upgrade' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('version_upgrade')}>
                            {selectedMenuItem === 'version_upgrade' ? <VersionUpgradeColorIcon alt="versionUpgrade" /> : <VersionUpgradeGrayIcon alt="versionUpgrade" />}
                            System information
                            {!state.isLatest && <div className="update-available">New version!</div>}
                        </div>
                    </>
                )}
                <div className={selectedMenuItem === 'profile' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('profile')}>
                    <PersonOutlinedIcon alt="versionUpgrade" />
                    Profile
                </div>
                <div className={selectedMenuItem === 'integrations' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('integrations')}>
                    {selectedMenuItem === 'integrations' ? <IntegrationColorIcon alt="notifications" /> : <IntegrationGrayIcon alt="notifications" />}
                    Integrations
                </div>
                <div className={selectedMenuItem === 'cluster_configuration' ? 'menu-item selected' : 'menu-item'} onClick={() => setMenuItem('cluster_configuration')}>
                    {selectedMenuItem === 'cluster_configuration' ? (
                        <ClusterConfColorIcon alt="clusterConfiguration" />
                    ) : (
                        <ClusterConfGrayIcon alt="clusterConfiguration" />
                    )}
                    Environment configuration
                </div>
            </div>
        </>
    );
}
export default AccountMenu;
