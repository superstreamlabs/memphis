// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

import Integrations from './integrations';
import Profile from './profile';
import Alerts from './alerts';

function Users() {
    const [selectedMenuItem, selectMenuItem] = useState('editProfile');
    const [state, dispatch] = useContext(Context);

    useEffect(() => {}, []);

    const getComponent = () => {
        switch (selectedMenuItem) {
            case 'editProfile':
                return <Profile />;
            case 'clusterConfiguration':
                return <Alerts />;
            case 'intergrations':
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
                        <img src={selectedMenuItem === 'editProfile' ? EditProfileColor : EditProfileGray} />
                        Edit Profile
                    </div>
                    <div
                        className={selectedMenuItem === 'clusterConfiguration' ? 'menu-item selected' : 'menu-item'}
                        onClick={() => selectMenuItem('clusterConfiguration')}
                    >
                        <img src={selectedMenuItem === 'clusterConfiguration' ? ClusterConfColor : ClusterConfGray} />
                        Cluster configuration
                    </div>
                    <div className={selectedMenuItem === 'intergrations' ? 'menu-item selected' : 'menu-item'} onClick={() => selectMenuItem('intergrations')}>
                        <img src={selectedMenuItem === 'intergrations' ? IntegrationColor : IntegrationGray} />
                        Intergrations
                    </div>
                </div>
            </div>
            <div className="setting-items">{getComponent()}</div>
        </div>
    );
}
export default Users;
