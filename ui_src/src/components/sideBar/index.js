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

import React, { useCallback, useContext, useEffect, useState } from 'react';
import { SettingOutlined, ExceptionOutlined } from '@ant-design/icons';
import ExitToAppOutlined from '@material-ui/icons/ExitToAppOutlined';
import PersonOutlinedIcon from '@material-ui/icons/PersonOutlined';
import { useHistory, Link } from 'react-router-dom';
import { Divider, Popover } from 'antd';

import {
    LOCAL_STORAGE_ACCOUNT_NAME,
    LOCAL_STORAGE_AVATAR_ID,
    LOCAL_STORAGE_COMPANY_LOGO,
    LOCAL_STORAGE_FULL_NAME,
    LOCAL_STORAGE_USER_NAME,
    USER_IMAGE
} from '../../const/localStorageConsts';
import integrationIconColor from '../../assets/images/integrationIconColor.svg';
import functionsIconActive from '../../assets/images/functionsIconActive.svg';
import overviewIconActive from '../../assets/images/overviewIconActive.svg';
import stationsIconActive from '../../assets/images/stationsIconActive.svg';
import documentIconActive from '../../assets/images/documentIconActive.svg';
import { compareVersions, isCloud } from '../../services/valueConvertor';
import supportIconColor from '../../assets/images/supportIconColor.svg';
import schemaIconActive from '../../assets/images/schemaIconActive.svg';
import integrationIcon from '../../assets/images/integrationIcon.svg';
import usersIconActive from '../../assets/images/usersIconActive.svg';
import functionsIcon from '../../assets/images/functionsIcon.svg';
import documentIcon from '../../assets/images/documentIcon.svg';
import overviewIcon from '../../assets/images/overviewIcon.svg';
import stationsIcon from '../../assets/images/stationsIcon.svg';
import supportIcon from '../../assets/images/supportIcon.svg';
import { GithubRequest } from '../../services/githubRequests';
import logsActive from '../../assets/images/logsActive.svg';
import schemaIcon from '../../assets/images/schemaIcon.svg';
import { DOC_URL, LATEST_RELEASE_URL } from '../../config';
import usersIcon from '../../assets/images/usersIcon.svg';
import logsIcon from '../../assets/images/logsIcon.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import Logo from '../../assets/images/logo.svg';
import AuthService from '../../services/auth';
import { Context } from '../../hooks/store';
import pathDomains from '../../router';
import Spinner from '../spinner';
import Support from './support';
import UpgradePlans from '../upgradePlans';

const overlayStyles = {
    borderRadius: '8px',
    width: '230px',
    paddingTop: '5px',
    paddingBottom: '5px',
    marginBottom: '10px'
};
const overlayStylesSupport = {
    borderRadius: '8px',
    width: '380px',
    padding: '15px',
    marginBottom: '10px'
};

function SideBar() {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const [avatarUrl, SetAvatarUrl] = useState(require('../../assets/images/bots/avatar1.svg'));
    const [popoverOpenSetting, setPopoverOpenSetting] = useState(false);
    const [popoverOpenSupport, setPopoverOpenSupport] = useState(false);
    const [hoveredItem, setHoveredItem] = useState('');
    const [logoutLoader, setLogoutLoader] = useState(false);
    const getCompanyLogo = useCallback(async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_COMPANY_LOGO);
            if (data) {
                localStorage.setItem(LOCAL_STORAGE_COMPANY_LOGO, data.image);
                dispatch({ type: 'SET_COMPANY_LOGO', payload: data.image });
            }
        } catch (error) {}
    }, []);
    const isRoot = state?.userData?.user_type === 'root';

    const getSystemVersion = useCallback(async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_INFO);
            if (data) {
                const latest = await GithubRequest(LATEST_RELEASE_URL);
                let is_latest = compareVersions(data.version, latest[0].name.replace('v', '').replace('-beta', '').replace('-latest', '').replace('-stable', ''));
                let system_version = data.version;
                dispatch({ type: 'IS_LATEST', payload: is_latest });
                dispatch({ type: 'CURRENT_VERSION', payload: system_version });
            }
        } catch (error) {}
    }, []);

    useEffect(() => {
        getCompanyLogo().catch(console.error);
        {
            !isCloud() && getSystemVersion().catch(console.error);
        }
        setAvatarImage(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID) || state?.userData?.avatar_id);
    }, []);

    useEffect(() => {
        setAvatarImage(localStorage.getItem(LOCAL_STORAGE_AVATAR_ID) || state?.userData?.avatar_id);
    }, [state]);

    const setAvatarImage = (avatarId) => {
        SetAvatarUrl(require(`../../assets/images/bots/avatar${avatarId}.svg`));
    };

    const handleLogout = async () => {
        setLogoutLoader(true);
        if (isCloud()) {
            try {
                const data = await httpRequest('POST', ApiEndpoints.SIGN_OUT);
                if (data) {
                    setTimeout(() => {
                        AuthService.logout();
                        setLogoutLoader(false);
                    }, 1000);
                }
            } catch (error) {
                setLogoutLoader(false);
            }
        } else {
            AuthService.logout();
            setTimeout(() => {
                setLogoutLoader(false);
            }, 1000);
        }
    };

    const contentSetting = (
        <div className="menu-content">
            <div className="item-wrap-header">
                <span className="img-section">
                    <img
                        className={'avatar-image'}
                        src={localStorage.getItem(USER_IMAGE) && localStorage.getItem(USER_IMAGE) !== 'undefined' ? localStorage.getItem(USER_IMAGE) : avatarUrl}
                        referrerPolicy="no-referrer"
                        width="30"
                        alt="avatar"
                    ></img>
                    <span className="company-logo">
                        <img src={state?.companyLogo || Logo} width="15" height="15" alt="companyLogo" />
                    </span>
                </span>
                <div className="account-details">
                    <p className="username">
                        {localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== 'undefined' && localStorage.getItem(LOCAL_STORAGE_FULL_NAME) !== ''
                            ? localStorage.getItem(LOCAL_STORAGE_FULL_NAME)
                            : localStorage.getItem(LOCAL_STORAGE_USER_NAME)}
                    </p>
                    <span className="company-name">{state?.userData?.account_name || localStorage.getItem(LOCAL_STORAGE_ACCOUNT_NAME)}</span>
                </div>
            </div>
            <Divider />
            <div
                className="item-wrap"
                onClick={() => {
                    history.replace(pathDomains.profile);
                    setPopoverOpenSetting(false);
                }}
            >
                <div className="item">
                    <span className="icons">
                        <PersonOutlinedIcon className="icons-sidebar" />
                    </span>
                    <p className="item-title">Profile</p>
                </div>
            </div>
            <div
                className="item-wrap"
                onClick={() => {
                    history.replace(`${pathDomains.administration}/integrations`);
                    setPopoverOpenSetting(false);
                }}
            >
                <div className="item">
                    <span className="icons">
                        <SettingOutlined className="icons-sidebar" />
                    </span>
                    <p className="item-title">Administration</p>
                </div>
            </div>
            {isCloud() && (
                <div
                    className="item-wrap"
                    onClick={() => {
                        history.replace(`${pathDomains.administration}/usage`);
                        setPopoverOpenSetting(false);
                    }}
                >
                    <div className="item">
                        <span className="icons">
                            <ExceptionOutlined className="icons-sidebar" />
                        </span>
                        <p className="item-title">Billing</p>
                    </div>
                </div>
            )}
            <div className="item-wrap" onClick={() => handleLogout()}>
                <div className="item">
                    <span className="icons">{logoutLoader ? <Spinner /> : <ExitToAppOutlined className="icons-sidebar" />}</span>
                    <p className="item-title">Log out</p>
                </div>
            </div>
        </div>
    );
    return (
        <div className="sidebar-container">
            <div className="upper-icons">
                <img
                    src={isCloud() ? state?.companyLogo || Logo : Logo}
                    width="45"
                    height="45"
                    className="logoimg"
                    alt="logo"
                    onClick={() => history.replace(pathDomains.overview)}
                />

                <div
                    className="item-wrapper"
                    onMouseEnter={() => setHoveredItem('overview')}
                    onMouseLeave={() => setHoveredItem('')}
                    onClick={() => history.replace(pathDomains.overview)}
                >
                    <div className="icon">
                        {state.route === 'overview' ? (
                            <img src={overviewIconActive} alt="overviewIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={hoveredItem === 'overview' ? overviewIconActive : overviewIcon} alt="overviewIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'overview' ? 'checked' : 'name'}>Overview</p>
                </div>
                <div
                    className="item-wrapper"
                    onMouseEnter={() => setHoveredItem('stations')}
                    onMouseLeave={() => setHoveredItem('')}
                    onClick={() => history.replace(pathDomains.stations)}
                >
                    <div className="icon">
                        {state.route === 'stations' ? (
                            <img src={stationsIconActive} alt="stationsIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={hoveredItem === 'stations' ? stationsIconActive : stationsIcon} alt="stationsIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'stations' ? 'checked' : 'name'}>Stations</p>
                </div>

                <div
                    className="item-wrapper"
                    onMouseEnter={() => setHoveredItem('schemaverse')}
                    onMouseLeave={() => setHoveredItem('')}
                    onClick={() => history.replace(`${pathDomains.schemaverse}/list`)}
                >
                    <div className="icon">
                        {state.route === 'schemaverse' ? (
                            <img src={schemaIconActive} alt="schemaIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={hoveredItem === 'schemaverse' ? schemaIconActive : schemaIcon} alt="schemaIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'schemaverse' ? 'checked' : 'name'}>Schemaverse</p>
                </div>
                {/* <div
                    className="item-wrapper"
                    onMouseEnter={() => setHoveredItem('functions')}
                    onMouseLeave={() => setHoveredItem('')}
                    onClick={() => history.replace(pathDomains.functions)}
                >
                    {state.route === 'functions' ? (
                        <img src={functionsIconActive} alt="functionsIcon" width="20" height="20"></img>
                    ) : (
                        <img src={hoveredItem === 'functions' ? functionsIconActive : functionsIcon} alt="functionsIcon" width="20" height="20"></img>
                    )}
                    <p className={state.route === 'functions' ? 'checked' : 'name'}>Functions</p>
                </div> */}
                <div
                    className="item-wrapper"
                    onMouseEnter={() => setHoveredItem('users')}
                    onMouseLeave={() => setHoveredItem('')}
                    onClick={() => history.replace(pathDomains.users)}
                >
                    <div className="icon">
                        {state.route === 'users' ? (
                            <img src={usersIconActive} alt="usersIconActive" width="20" height="20"></img>
                        ) : (
                            <img src={hoveredItem === 'users' ? usersIconActive : usersIcon} alt="usersIcon" width="20" height="20"></img>
                        )}
                    </div>
                    <p className={state.route === 'users' ? 'checked' : 'name'}>Users</p>
                </div>
                {!isCloud() && (
                    <div
                        className="item-wrapper"
                        onMouseEnter={() => setHoveredItem('logs')}
                        onMouseLeave={() => setHoveredItem('')}
                        onClick={() => history.replace(pathDomains.sysLogs)}
                    >
                        <div className="icon">
                            {state.route === 'logs' ? (
                                <img src={logsActive} alt="usersIconActive" width="20" height="20"></img>
                            ) : (
                                <img src={hoveredItem === 'logs' ? logsActive : logsIcon} alt="usersIcon" width="20" height="20"></img>
                            )}
                        </div>
                        <p className={state.route === 'logs' ? 'checked' : 'name'}>Logs</p>
                    </div>
                )}
                {isCloud() && (
                    <div className="item-wrapper">
                        <div className="icon not-available">
                            <img src={functionsIcon} alt="usersIcon" width="20" height="20"></img>
                        </div>
                        <p className="not-available">Functions</p>
                        <p className="coming-soon">Soon</p>
                    </div>
                )}
            </div>
            <div className="bottom-icons">
                <div
                    className="integration-icon-wrapper"
                    onMouseEnter={() => setHoveredItem('integrations')}
                    onMouseLeave={() => setHoveredItem('')}
                    onClick={() => history.replace(`${pathDomains.administration}/integrations`)}
                >
                    <img src={hoveredItem === 'integrations' ? integrationIconColor : integrationIcon} />
                    <label className="icon-name">Integrations</label>
                </div>
                {isCloud() && (
                    <Popover
                        overlayInnerStyle={overlayStylesSupport}
                        placement="right"
                        content={<Support closeModal={(e) => setPopoverOpenSupport(e)} />}
                        trigger="click"
                        onOpenChange={() => setPopoverOpenSupport(!popoverOpenSupport)}
                        open={popoverOpenSupport}
                    >
                        <div
                            className="integration-icon-wrapper"
                            onMouseEnter={() => setHoveredItem('support')}
                            onMouseLeave={() => setHoveredItem('')}
                            onClick={() => setPopoverOpenSupport(true)}
                        >
                            <img src={hoveredItem === 'support' ? supportIconColor : supportIcon} />
                            <label className="icon-name">Support</label>
                        </div>
                    </Popover>
                )}
                {!isCloud() && (
                    <Link to={{ pathname: DOC_URL }} target="_blank">
                        <div className="integration-icon-wrapper" onMouseEnter={() => setHoveredItem('documentation')} onMouseLeave={() => setHoveredItem('')}>
                            <img src={hoveredItem === 'documentation' ? documentIconActive : documentIcon} />
                            <label className="icon-name">Docs</label>
                        </div>
                    </Link>
                )}

                <Popover
                    overlayInnerStyle={overlayStyles}
                    placement="right"
                    content={contentSetting}
                    trigger="click"
                    onOpenChange={() => setPopoverOpenSetting(!popoverOpenSetting)}
                    open={popoverOpenSetting}
                >
                    <div className="sub-icon-wrapper" onClick={() => setPopoverOpenSetting(true)}>
                        <img
                            className={`sandboxUserImg ${(state.route === 'profile' || state.route === 'administration') && 'sandboxUserImgSelected'}`}
                            src={localStorage.getItem(USER_IMAGE) && localStorage.getItem(USER_IMAGE) !== 'undefined' ? localStorage.getItem(USER_IMAGE) : avatarUrl}
                            referrerPolicy="no-referrer"
                            width={localStorage.getItem(USER_IMAGE) && localStorage.getItem(USER_IMAGE) !== 'undefined' ? 35 : 25}
                            height={localStorage.getItem(USER_IMAGE) && localStorage.getItem(USER_IMAGE) !== 'undefined' ? 35 : 25}
                            alt="avatar"
                        ></img>
                    </div>
                </Popover>
                {!isCloud() && (
                    <version
                        is="x3d"
                        style={{ cursor: !state.isLatest ? 'pointer' : 'default' }}
                        onClick={() => (!state.isLatest ? history.replace(`${pathDomains.administration}/version_upgrade`) : null)}
                    >
                        {!state.isLatest && <div className="update-note" />}
                        <p>v{state.currentVersion}</p>
                    </version>
                )}
                {isCloud() && (
                    <UpgradePlans
                        content={
                            <div className="upgrade-button-wrapper">
                                <p className="upgrade-plan">Upgrade</p>
                            </div>
                        }
                        isExternal={false}
                    />
                )}
            </div>
        </div>
    );
}

export default SideBar;
