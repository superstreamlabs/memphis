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

function Users() {
    const [selectedMenuItem, selectMenuItem] = useState('editProfile');
    const [state, dispatch] = useContext(Context);

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'settings' });
    }, []);

    const getComponent = () => {
        switch (selectedMenuItem) {
            case 'editProfile':
                return <Profile />;
            case 'clusterConfiguration':
                return <ClusterConfiguration />;
            case 'notifications':
                return <Integrations />;
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
                    <div className={selectedMenuItem === 'editProfile' ? 'menu-item selected' : 'menu-item'} onClick={() => selectMenuItem('editProfile')}>
                        <img src={selectedMenuItem === 'editProfile' ? EditProfileColor : EditProfileGray} alt="editProfile" />
                        Edit Profile
                    </div>
                    <div className={selectedMenuItem === 'notifications' ? 'menu-item selected' : 'menu-item'} onClick={() => selectMenuItem('notifications')}>
                        <img src={selectedMenuItem === 'notifications' ? IntegrationColor : IntegrationGray} alt="notifications" />
                        Notifications
                    </div>
                    <div
                        className="menu-item disabled"
                        //  className={selectedMenuItem === 'clusterConfiguration' ? 'menu-item selected' : 'menu-item'}
                        // onClick={() => selectMenuItem('clusterConfiguration')}
                    >
                        <img src={selectedMenuItem === 'clusterConfiguration' ? ClusterConfColor : ClusterConfGray} alt="clusterConfiguration" />
                        Cluster configuration
                    </div>
                </div>
            </div>
            <div className="setting-items">{getComponent()}</div>
        </div>
    );
}
export default Users;
