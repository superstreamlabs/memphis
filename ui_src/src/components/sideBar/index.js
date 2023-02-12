// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useCallback, useContext, useEffect, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { Link } from 'react-router-dom';
import { Menu } from 'antd';

import {
    LOCAL_STORAGE_AVATAR_ID,
    LOCAL_STORAGE_COMPANY_LOGO,
    LOCAL_STORAGE_FULL_NAME,
    LOCAL_STORAGE_USER_NAME,
    LOCAL_STORAGE_SKIP_GET_STARTED
} from '../../const/localStorageConsts';
import integrationNavIcon from '../../assets/images/integrationNavIcon.svg';
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
import TooltipComponent from '../tooltip/tooltip';
import Modal from '../modal';
import { capitalizeFirst } from '../../services/valueConvertor';
const { SubMenu } = Menu;

function SideBar() {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const [avatarUrl, SetAvatarUrl] = useState(require('../../assets/images/bots/avatar1.svg'));
    const [systemVersion, setSystemVersion] = useState('');
    const [open, modalFlip] = useState(false);
    const [goToRoute, setGoToRoute] = useState(null);

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

    useEffect(() => {
        if (goToRoute && state.route !== goToRoute) {
            if (state?.route === 'overview' && localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED) !== 'true') modalFlip(true);
            else typeof goToRoute === 'object' && goToRoute !== null ? handleClick(goToRoute) : handleChangeRoute(goToRoute);
        }
    }, [goToRoute]);

    const handleChangeRoute = () => {
        history.push(goToRoute);
    };

    const handleClick = async (e) => {
        switch (e.key) {
            case '1':
                history.push(`${pathDomains.preferences}/profile`);
                break;
            case '3':
                await AuthService.logout();
                break;
            default:
                break;
        }
    };

    const skipGetStarted = async () => {
        try {
            await httpRequest('POST', ApiEndpoints.SKIP_GET_STARTED, { username: capitalizeFirst(localStorage.getItem(LOCAL_STORAGE_USER_NAME)) });
            localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, true);
            typeof goToRoute === 'object' && goToRoute !== null ? handleClick(goToRoute) : handleChangeRoute(goToRoute);
            modalFlip(false);
        } catch (error) {}
    };

    return (
        <div className="sidebar-container">
            <div className="upper-icons">
                <Modal
                    header="Skip get-started?"
                    height="100px"
                    width="400px"
                    rBtnText="Skip"
                    lBtnText="Don't skip"
                    lBtnClick={() => {
                        typeof goToRoute === 'object' && goToRoute !== null ? handleClick(goToRoute) : handleChangeRoute(goToRoute);
                        modalFlip(false);
                    }}
                    rBtnClick={() => {
                        skipGetStarted();
                    }}
                    clickOutside={() => {
                        setGoToRoute(state.route);
                        modalFlip(false);
                    }}
                    open={open}
                >
                    <p>After choosing "skip" you won't see get-started flow anymore.</p>
                </Modal>
                <img src={betaLogo} width="62" className="logoimg" alt="logo" onClick={() => setGoToRoute(pathDomains.overview)} />
                <div className="item-wrapper" onClick={() => setGoToRoute(pathDomains.overview)}>
                    <div className="icon">
                        {state.route === 'overview' ? (
                            <img src={overviewIconActive} alt="overviewIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={overviewIcon} alt="overviewIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'overview' ? 'checked' : 'name'}>Overview</p>
                </div>
                <div className="item-wrapper" onClick={() => setGoToRoute(pathDomains.stations)}>
                    <div className="icon">
                        {state.route === 'stations' ? (
                            <img src={stationsIconActive} alt="stationsIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={stationsIcon} alt="stationsIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'stations' ? 'checked' : 'name'}>Stations</p>
                </div>
                <div className="item-wrapper" onClick={() => setGoToRoute(pathDomains.schemaverse)}>
                    <div className="icon">
                        {state.route === 'schemaverse' ? (
                            <img src={schemaIconActive} alt="schemaIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={schemaIcon} alt="schemaIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'schemaverse' ? 'checked' : 'name'}>Schemaverse</p>
                </div>
                <div className="item-wrapper" onClick={() => setGoToRoute(pathDomains.users)}>
                    <div className="icon">
                        {state.route === 'users' ? (
                            <img src={usersIconActive} alt="usersIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={usersIcon} alt="usersIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'users' ? 'checked' : 'name'}>Users</p>
                </div>
                <div className="item-wrapper" onClick={() => setGoToRoute(pathDomains.sysLogs)}>
                    <div className="icon">
                        {state.route === 'logs' ? (
                            <img src={logsActive} alt="usersIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={logsIcon} alt="usersIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'logs' ? 'checked' : 'name'}>Logs</p>
                </div>
            </div>
            <div className="bottom-icons">
                <TooltipComponent text="Integrations" placement="right">
                    <div className="integration-icon-wrapper" onClick={() => setGoToRoute(`${pathDomains.preferences}/integrations`)}>
                        <img src={integrationNavIcon} />
                    </div>
                </TooltipComponent>
                <Menu onClick={(e) => setGoToRoute(e)} className="app-menu" mode="vertical" triggerSubMenuAction="click">
                    <SubMenu
                        key="subMenu"
                        icon={
                            <div className={state.route === 'preferences' ? 'sub-icon-wrapper menu-preference-selected' : 'sub-icon-wrapper'}>
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
                            id="setting-menu"
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
                                <div className="item-wrapp">
                                    <img src={accountIcon} width="15" height="15" alt="accountIcon" />
                                    <p className="item-title">Preferences</p>
                                </div>
                            </Menu.Item>
                            <Menu.Item key={2}>
                                <Link to={{ pathname: DOC_URL }} target="_blank">
                                    <div className="item-wrapp">
                                        <img src={supportIcon} width="15" height="15" alt="supportIcon" />
                                        <p className="item-title">Support</p>
                                    </div>
                                </Link>
                            </Menu.Item>
                            <Menu.Item key={3}>
                                <div className="item-wrapp">
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
