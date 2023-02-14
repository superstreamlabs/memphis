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
import { Popover } from 'antd';
import { SettingOutlined, QuestionCircleOutlined, LogoutOutlined } from '@ant-design/icons';

import { LOCAL_STORAGE_AVATAR_ID, LOCAL_STORAGE_COMPANY_LOGO, LOCAL_STORAGE_FULL_NAME, LOCAL_STORAGE_USER_NAME } from '../../const/localStorageConsts';
import integrationNavIcon from '../../assets/images/integrationNavIcon.svg';
import overviewIconActive from '../../assets/images/overviewIconActive.svg';
import stationsIconActive from '../../assets/images/stationsIconActive.svg';
import schemaIconActive from '../../assets/images/schemaIconActive.svg';
import usersIconActive from '../../assets/images/usersIconActive.svg';
import overviewIcon from '../../assets/images/overviewIcon.svg';
import stationsIcon from '../../assets/images/stationsIcon.svg';
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

const overlayStyles = {
    borderRadius: '4px',
    width: '230px',
    padding: '5px'
};

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
                history.push(`${pathDomains.preferences}/profile`);
                break;
            case '3':
                await AuthService.logout();
                break;
            default:
                break;
        }
    };
    const content = (
        <div>
            <div className="item-wrapp" onClick={() => history.push(pathDomains.profile)}>
                <span className="icons">
                    <img className="logoimg" src={state?.companyLogo || Logo} width="24" alt="companyLogo" />
                </span>
                <p>
                    {localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== 'undefined' && localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== ''
                        ? localStorage.getItem(LOCAL_STORAGE_FULL_NAME)
                        : localStorage.getItem(LOCAL_STORAGE_USER_NAME)}
                </p>
            </div>
            <div className="item-wrapp" onClick={() => history.push(`${pathDomains.preferences}/integrations`)}>
                <span className="icons">
                    <SettingOutlined className="icons-sidebar" />
                </span>
                <p className="item-title">Preferences</p>
            </div>

            <Link to={{ pathname: DOC_URL }} target="_blank">
                <div className="item-wrapp">
                    <span className="icons">
                        <QuestionCircleOutlined className="icons-sidebar" />
                    </span>
                    <p className="item-title">Support</p>
                </div>
            </Link>
            <div className="item-wrapp">
                <span className="icons">
                    <LogoutOutlined className="icons-sidebar" />
                </span>
                <p className="item-title">Log out</p>
            </div>
        </div>
    );
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
                <div className="item-wrapper">
                    <Link to={pathDomains.schemaverse}>
                        <div className="icon">
                            {state.route === 'schemaverse' ? (
                                <img src={schemaIconActive} alt="schemaIconActive" width="20" height="20"></img>
                            ) : (
                                <img src={schemaIcon} alt="schemaIcon" width="20" height="20"></img>
                            )}
                        </div>
                        <p className={state.route === 'schemaverse' ? 'checked' : 'name'}>Schemaverse</p>
                    </Link>
                </div>
                <div className="item-wrapper">
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
                <div className="item-wrapper">
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
            <div className="bottom-icons">
                <Link to={`${pathDomains.preferences}/integrations`}>
                    <div className="integration-icon-wrapper">
                        <TooltipComponent text="Integrations" placement="right">
                            <img src={integrationNavIcon} />
                        </TooltipComponent>
                    </div>
                </Link>
                <Popover overlayInnerStyle={overlayStyles} placement="rightBottom" content={content} trigger="click">
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
                </Popover>
                <version is="x3d">
                    <p>v{systemVersion}</p>
                </version>
            </div>
        </div>
    );
}

export default SideBar;
