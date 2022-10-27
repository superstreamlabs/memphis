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

import React, { useCallback, useContext, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { Link } from 'react-router-dom';
import { Menu } from 'antd';

import { LOCAL_STORAGE_AVATAR_ID, LOCAL_STORAGE_COMPANY_LOGO, LOCAL_STORAGE_FULL_NAME, LOCAL_STORAGE_USER_NAME } from '../../const/localStorageConsts';
import overviewIconActive from '../../assets/images/overviewIconActive.svg';
import stationsIconActive from '../../assets/images/stationsIconActive.svg';
import schemaIconActive from '../../assets/images/schemaIconActive.svg';
import usersIconActive from '../../assets/images/usersIconActive.svg';
import overviewIcon from '../../assets/images/overviewIcon.svg';
import stationsIcon from '../../assets/images/stationsIcon.svg';
import supportIcon from '../../assets/images/supportIcon.svg';
import accountIcon from '../../assets/images/accountIcon.svg';
import logoutIcon from '../../assets/images/logoutIcon.svg';
import logsActive from '../../assets/images/logsActive.svg';
import schemaIcon from '../../assets/images/schemaIcon.svg';
import usersIcon from '../../assets/images/usersIcon.svg';
import betaLogo from '../../assets/images/betaLogo.svg';
import logsIcon from '../../assets/images/logsIcon.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import Logo from '../../assets/images/logo.svg';
import AuthService from '../../services/auth';
import { Context } from '../../hooks/store';
import pathDomains from '../../router';
import { DOC_URL } from '../../config';

const { SubMenu } = Menu;

function SideBar() {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const [avatarUrl, SetAvatarUrl] = useState(require('../../assets/images/bots/avatar1.svg'));
    const [systemVersion, setSystemVersion] = useState('');

    const getCompanyLogo = useCallback(async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_COMPANY_LOGO);
            if (data) {
                localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, data.image);
                dispatch({ type: 'SET_COMPANY_LOGO', payload: data.image });
            }
        } catch (error) {}
    }, []);

    const getSystemVersion = useCallback(async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_INFO);
            if (data) {
                setSystemVersion(data.version);
            }
        } catch (error) {}
    }, []);

    useEffect(() => {
        getCompanyLogo().catch(console.error);
        getSystemVersion().catch(console.error);
        setAvatarImage(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID) || state?.userData?.avatar_id);
    }, []);

    useEffect(() => {
        setAvatarImage(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID) || state?.userData?.avatar_id);
    }, [state]);

    const setAvatarImage = (avatarId) => {
        SetAvatarUrl(require(`../../assets/images/bots/avatar${avatarId}.svg`));
    };

    const handleClick = async (e) => {
        switch (e.key) {
            case '1':
                history.push(pathDomains.settings);
                break;
            case '3':
                await AuthService.logout();
                break;
            default:
                break;
        }
    };

    return (
        <div className="sidebar-container">
            <div className="upper-icons">
                <Link to={pathDomains.overview}>
                    <img src={betaLogo} width="62" className="logoimg" alt="logo" />
                </Link>
                <div className="item-wrapper">
                    <Link to={pathDomains.overview}>
                        <div className="icon">
                            {state.route === 'overview' ? (
                                <img src={overviewIconActive} alt="overviewIconActive" width="20" height="20"></img>
                            ) : (
                                <img src={overviewIcon} alt="overviewIcon" width="20" height="20"></img>
                            )}
                        </div>
                        <p className={state.route === 'overview' ? 'checked' : 'name'}>Overview</p>
                    </Link>
                </div>
                <div className="item-wrapper">
                    <div id="e2e-tests-station-sidebar">
                        <Link to={pathDomains.stations}>
                            <div className="icon">
                                {state.route === 'stations' ? (
                                    <img src={stationsIconActive} alt="stationsIconActive" width="20" height="20"></img>
                                ) : (
                                    <img src={stationsIcon} alt="stationsIcon" width="20" height="20"></img>
                                )}
                            </div>
                            <p className={state.route === 'stations' ? 'checked' : 'name'}>Stations</p>
                        </Link>
                    </div>
                </div>
                <div className="item-wrapper">
                    <div id="e2e-tests-users-sidebar">
                        <Link to={pathDomains.schemas}>
                            <div className="icon">
                                {state.route === 'schemas' ? (
                                    <img src={schemaIconActive} alt="schemaIconActive" width="20" height="20"></img>
                                ) : (
                                    <img src={schemaIcon} alt="schemaIcon" width="20" height="20"></img>
                                )}
                            </div>
                            <p className={state.route === 'schemas' ? 'checked' : 'name'}>Schemas</p>
                        </Link>
                    </div>
                </div>
                <div className="item-wrapper">
                    <div id="e2e-tests-users-sidebar">
                        <Link to={pathDomains.users}>
                            <div className="icon">
                                {state.route === 'users' ? (
                                    <img src={usersIconActive} alt="usersIconActive" width="20" height="20"></img>
                                ) : (
                                    <img src={usersIcon} alt="usersIcon" width="20" height="20"></img>
                                )}
                            </div>
                            <p className={state.route === 'users' ? 'checked' : 'name'}>Users</p>
                        </Link>
                    </div>
                </div>
                <div className="item-wrapper">
                    <div id="e2e-tests-users-sidebar">
                        <Link to={pathDomains.sysLogs}>
                            <div className="icon">
                                {state.route === 'logs' ? (
                                    <img src={logsActive} alt="usersIconActive" width="20" height="20"></img>
                                ) : (
                                    <img src={logsIcon} alt="usersIcon" width="20" height="20"></img>
                                )}
                            </div>
                            <p className={state.route === 'logs' ? 'checked' : 'name'}>Logs</p>
                        </Link>
                    </div>
                </div>
            </div>
            <div id="e2e-tests-settings-btn" className="bottom-icons">
                <Menu onClick={handleClick} className="app-menu" mode="vertical" triggerSubMenuAction="click">
                    <SubMenu
                        key="subMenu"
                        icon={
                            <div className="sub-icon-wrapper">
                                <img
                                    className="sandboxUserImg"
                                    src={localStorage.getItem('profile_pic') || avatarUrl} // profile_pic is available only in sandbox env
                                    referrerPolicy="no-referrer"
                                    width={localStorage.getItem('profile_pic') ? 35 : 25}
                                    height={localStorage.getItem('profile_pic') ? 35 : 25}
                                    alt="avatar"
                                ></img>
                            </div>
                        }
                    >
                        <Menu.ItemGroup
                            title={
                                <div className="header-menu">
                                    <div className="company-logo">
                                        <img className="logoimg" src={state?.companyLogo || Logo} width="24" alt="companyLogo" />
                                    </div>
                                    <p>
                                        {localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== 'undefined' && localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== ''
                                            ? localStorage.getItem(LOCAL_STORAGE_FULL_NAME)
                                            : localStorage.getItem(LOCAL_STORAGE_USER_NAME)}
                                    </p>
                                </div>
                            }
                        >
                            <Menu.Item key={1} className="customclass">
                                <div className="item-wrapp" id="e2e-tests-menu-preferences">
                                    <img src={accountIcon} width="15" height="15" alt="accountIcon" />
                                    <p className="item-title">Preferences</p>
                                </div>
                            </Menu.Item>
                            <Menu.Item key={2}>
                                <Link to={{ pathname: DOC_URL }} target="_blank">
                                    <div className="item-wrapp" id="e2e-tests-menu-support">
                                        <img src={supportIcon} width="15" height="15" alt="supportIcon" />
                                        <p className="item-title">Support</p>
                                    </div>
                                </Link>
                            </Menu.Item>
                            <Menu.Item key={3}>
                                <div className="item-wrapp" id="e2e-tests-menu-logout">
                                    <img src={logoutIcon} width="15" height="15" alt="logoutIcon" />
                                    <p className="item-title">Log out</p>
                                </div>
                            </Menu.Item>
                        </Menu.ItemGroup>
                    </SubMenu>
                </Menu>
                <version is="x3d">
                    <p>v{systemVersion}</p>
                </version>
            </div>
        </div>
    );
}

export default SideBar;
