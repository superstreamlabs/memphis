// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useContext, useState, useEffect } from 'react';
import { Divider } from 'antd';

import CustomTabs from '../../components/Tabs';
import { Context } from '../../hooks/store';
import ClusterConfColor from '../../assets/images/setting/clusterConfColor.svg';
import ClusterConfGray from '../../assets/images/setting/clusterConfGray.svg';
import EditProfileColor from '../../assets/images/setting/editProfileColor.svg';
import EditProfileGray from '../../assets/images/setting/editProfileGray.svg';
import IntegrationColor from '../../assets/images/setting/integrationColor.svg';
import IntegrationGray from '../../assets/images/setting/integrationGray.svg';
import NotificationGray from '../../assets/images/setting/notificationGray.svg';
import Integrations from './integrations';
import Profile from './profile';
import Alerts from './clusterConfiguration';
import ClusterConfiguration from './clusterConfiguration';
import { useHistory } from 'react-router-dom';
import pathDomains from '../../router';

function Preferences({ step }) {
    const [selectedMenuItem, selectMenuItem] = useState(step || 'profile');
    const [state, dispatch] = useContext(Context);
    const history = useHistory();

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'preferences' });
    }, []);

    const getComponent = () => {
        switch (selectedMenuItem) {
            case 'profile':
                if (window.location.href.split('/profile').length > 1) {
                    return <Profile />;
                } else {
                    history.push(`${pathDomains.preferences}/profile`);
                    break;
                }
            case 'cluster_configuration':
                if (window.location.href.split('/cluster_configuration').length > 1) {
                    return <ClusterConfiguration />;
                } else {
                    history.push(`${pathDomains.preferences}/cluster_configuration`);
                    break;
                }
            case 'integrations':
                if (window.location.href.split('/integrations').length > 1) {
                    return <Integrations />;
                } else {
                    history.push(`${pathDomains.preferences}/integrations`);
                    break;
                }
            default:
                return;
        }
    };
    return (
        <div className="setting-container">
            <div className="menu-container">
                <p className="header">My account</p>
                <p className="sub-header">Update and manage your account</p>
                <div className="side-menu">
                    <div className={selectedMenuItem === 'profile' ? 'menu-item selected' : 'menu-item'} onClick={() => selectMenuItem('profile')}>
                        <img src={selectedMenuItem === 'profile' ? EditProfileColor : EditProfileGray} alt="editProfile" />
                        Edit Profile
                    </div>
                    <div className={selectedMenuItem === 'integrations' ? 'menu-item selected' : 'menu-item'} onClick={() => selectMenuItem('integrations')}>
                        <img src={selectedMenuItem === 'integrations' ? IntegrationColor : IntegrationGray} alt="notifications" />
                        Integrations
                    </div>
                    <div
                        className="menu-item disabled"
                        //  className={selectedMenuItem === 'clusterConfiguration' ? 'menu-item selected' : 'menu-item'}
                        // onClick={() => selectMenuItem('cluster_configuration')}
                    >
                        <img src={selectedMenuItem === 'cluster_configuration' ? ClusterConfColor : ClusterConfGray} alt="clusterConfiguration" />
                        Cluster configuration
                    </div>
                </div>
            </div>
            <div className="setting-items">{getComponent()}</div>
        </div>
    );
}
export default Preferences;
