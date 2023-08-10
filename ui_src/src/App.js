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

import './App.scss';

import { Switch, Route, withRouter } from 'react-router-dom';
import React, { useCallback, useContext, useEffect, useRef, useState } from 'react';
import { JSONCodec, StringCodec, connect } from 'nats.ws';
import { useMediaQuery } from 'react-responsive';
import { useHistory } from 'react-router-dom';
import { message, notification } from 'antd';
import { Redirect } from 'react-router-dom';

import {
    LOCAL_STORAGE_ACCOUNT_ID,
    LOCAL_STORAGE_INTERNAL_WS_PASS,
    LOCAL_STORAGE_CONNECTION_TOKEN,
    LOCAL_STORAGE_TOKEN,
    LOCAL_STORAGE_USER_PASS_BASED_AUTH,
    LOCAL_STORAGE_WS_PORT,
    USER_IMAGE
} from './const/localStorageConsts';
import { CLOUD_URL, ENVIRONMENT, HANDLE_REFRESH_INTERVAL, WS_PREFIX, WS_SERVER_URL_PRODUCTION } from './config';
import { handleRefreshTokenRequest, httpRequest } from './services/http';
import infoNotificationIcon from './assets/images/infoNotificationIcon.svg';
import redirectIcon from './assets/images/redirectIcon.svg';
import successIcon from './assets/images/successIcon.svg';
import close from './assets/images/closeNotification.svg';
import StationOverview from './domain/stationOverview';
import errorIcon from './assets/images/errorIcon.svg';
import MessageJourney from './domain/messageJourney';
import Administration from './domain/administration';
import { ApiEndpoints } from './const/apiEndpoints';
import { isCloud } from './services/valueConvertor';
import warnIcon from './assets/images/warnIcon.svg';
import AppWrapper from './components/appWrapper';
import StationsList from './domain/stationsList';
import SchemaManagment from './domain/schema';
import PrivateRoute from './PrivateRoute';
import AuthService from './services/auth';
import Overview from './domain/overview';
import Loader from './components/loader';
import { Context } from './hooks/store';
import Profile from './domain/profile';
import pathDomains from './router';
import Users from './domain/users';
import Functions from './domain/functions';

let SysLogs = undefined;
let Login = undefined;
let Signup = undefined;

if (!isCloud()) {
    SysLogs = require('./domain/sysLogs').default;
    Login = require('./domain/login').default;
    Signup = require('./domain/signup').default;
}

const App = withRouter((props) => {
    const [state, dispatch] = useContext(Context);
    const isMobile = useMediaQuery({ maxWidth: 849 });
    const [authCheck, setAuthCheck] = useState(true);
    const history = useHistory();
    const urlParams = new URLSearchParams(window.location.search);
    const firebase_id_token = urlParams.get('firebase_id_token');
    const firebase_organization_id = urlParams.get('firebase_organization_id');
    const [cloudLogedIn, setCloudLogedIn] = useState(isCloud() ? false : true);
    const [persistedNotifications, setPersistedNotifications] = useState(() => {
        const storedNotifications = JSON.parse(localStorage.getItem('persistedNotifications'));
        return storedNotifications || [];
    });
    const [displayedNotifications, setDisplayedNotifications] = useState([]);
    const [systemMessage, setSystemMessage] = useState([]);

    const stateRef = useRef([]);
    stateRef.current = [cloudLogedIn, persistedNotifications];

    const handleLoginWithToken = async () => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.LOGIN, { firebase_id_token, firebase_organization_id }, {}, {}, false);
            if (data) {
                localStorage.setItem(USER_IMAGE, data.user_image);
                AuthService.saveToLocalStorage(data);
                try {
                    const ws_port = data.ws_port;
                    const SOCKET_URL = ENVIRONMENT === 'production' ? `${WS_PREFIX}://${WS_SERVER_URL_PRODUCTION}:${ws_port}` : `${WS_PREFIX}://localhost:${ws_port}`;
                    let conn;
                    if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                        const account_id = localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID);
                        const internal_ws_pass = localStorage.getItem(LOCAL_STORAGE_INTERNAL_WS_PASS);
                        conn = await connect({
                            servers: [SOCKET_URL],
                            user: '$memphis_user$' + account_id,
                            pass: internal_ws_pass,
                            timeout: '5000'
                        });
                    } else {
                        const connection_token = localStorage.getItem(LOCAL_STORAGE_CONNECTION_TOKEN);
                        conn = await connect({
                            servers: [SOCKET_URL],
                            token: '::' + connection_token,
                            timeout: '5000'
                        });
                    }
                    dispatch({ type: 'SET_SOCKET_DETAILS', payload: conn });
                } catch (error) {
                    return;
                }
                dispatch({ type: 'SET_USER_DATA', payload: data });
            }
            history.push('/overview');
            setCloudLogedIn(true);
        } catch (error) {}
    };

    useEffect(() => {
        if (isCloud() && firebase_id_token) {
            const fetchData = async () => {
                await handleLoginWithToken();
            };
            fetchData();
        } else setCloudLogedIn(true);
    }, []);

    useEffect(() => {
        if (isMobile) {
            message.warn({
                key: 'memphisWarningMessage',
                duration: 0,
                content: 'Hi, please pay attention. We do not support these dimensions.',
                style: { cursor: 'not-allowed' }
            });
        }
        return () => {
            message.destroy('memphisWarningMessage');
        };
    }, [isMobile]);

    const handleRefresh = useCallback(async (firstTime) => {
        if (window.location.pathname === pathDomains.login || (firebase_id_token !== null && !stateRef.current[0])) {
            return;
        } else if (localStorage.getItem(LOCAL_STORAGE_TOKEN)) {
            const ws_port = localStorage.getItem(LOCAL_STORAGE_WS_PORT);
            const SOCKET_URL = ENVIRONMENT === 'production' ? `${WS_PREFIX}://${WS_SERVER_URL_PRODUCTION}:${ws_port}` : `${WS_PREFIX}://localhost:${ws_port}`;
            const handleRefreshStatus = await handleRefreshTokenRequest();
            if (handleRefreshStatus) {
                if (firstTime) {
                    try {
                        let conn;
                        if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                            const account_id = localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID);
                            const internal_ws_pass = localStorage.getItem(LOCAL_STORAGE_INTERNAL_WS_PASS);
                            conn = await connect({
                                servers: [SOCKET_URL],
                                user: '$memphis_user$' + account_id,
                                pass: internal_ws_pass,
                                timeout: '5000'
                            });
                        } else {
                            const connection_token = localStorage.getItem(LOCAL_STORAGE_CONNECTION_TOKEN);
                            conn = await connect({
                                servers: [SOCKET_URL],
                                token: '::' + connection_token,
                                timeout: '5000'
                            });
                        }
                        dispatch({ type: 'SET_SOCKET_DETAILS', payload: conn });
                    } catch (error) {
                        return;
                    }
                }
                return true;
            }
        } else {
            isCloud() ? window.location.replace(CLOUD_URL) : history.push(pathDomains.signup);
        }
    }, []);

    useEffect(() => {
        const fetchData = async () => {
            await Promise.all([handleRefresh(true), setAuthCheck(false)]);
        };

        fetchData();

        const interval = setInterval(() => {
            handleRefresh(false);
        }, HANDLE_REFRESH_INTERVAL);

        return () => {
            clearInterval(interval);
            state.socket?.close();
        };
    }, [handleRefresh, setAuthCheck]);

    useEffect(() => {
        const sc = StringCodec();
        const jc = JSONCodec();
        let sub;
        const subscribeToNotifications = async () => {
            try {
                const rawBrokerName = await state.socket?.request(`$memphis_ws_subs.get_system_messages`, sc.encode('SUB'));
                if (rawBrokerName) {
                    const brokerName = JSON.parse(sc.decode(rawBrokerName?._rdata))['name'];
                    sub = state.socket?.subscribe(`$memphis_ws_pubs.get_system_messages.${brokerName}`);
                    listenForUpdates();
                }
            } catch (err) {
                console.error('Error subscribing to overview data:', err);
            }
        };

        const listenForUpdates = async () => {
            try {
                if (sub) {
                    for await (const msg of sub) {
                        let data = jc.decode(msg.data);
                        const uniqueNewNotifications = data.filter((newNotification) => {
                            return !stateRef.current[1].some((existingNotification) => existingNotification.id === newNotification.id);
                        });
                        const systemMeesage = data.filter((sys) => {
                            return sys.message_type === 'system';
                        });
                        setSystemMessage(systemMeesage);
                        setPersistedNotifications((prevPersistedNotifications) => [...prevPersistedNotifications, ...uniqueNewNotifications]);
                        localStorage.setItem('persistedNotifications', JSON.stringify([...stateRef.current[1], ...uniqueNewNotifications]));
                    }
                }
            } catch (err) {
                console.error('Error receiving overview data updates:', err);
            }
        };

        subscribeToNotifications();

        return () => {
            if (sub) {
                try {
                    sub.unsubscribe();
                } catch (err) {
                    console.error('Error unsubscribing from overview data:', err);
                }
            }
        };
    }, [state.socket]);

    const notificationHandler = (id, type, message, duration) => {
        const defaultAntdField = {
            className: 'notification-wrapper',
            closeIcon: <img src={close} alt="close" />,
            message: 'System Message',
            onClose: () => {
                const updatedNotifications = stateRef.current[1].map((n) => (n.id === id ? { ...n, read: true } : n));
                setPersistedNotifications(updatedNotifications);
                localStorage.setItem('persistedNotifications', JSON.stringify(updatedNotifications));
            }
        };
        switch (type) {
            case 'info':
                notification.info({
                    ...defaultAntdField,
                    icon: <img src={infoNotificationIcon} alt="info" />,
                    description: message,
                    duration: duration
                });
                break;
            case 'warning':
                notification.warning({
                    ...defaultAntdField,

                    icon: <img src={warnIcon} alt="warn" />,
                    description: message,
                    duration: duration
                });
                break;
            case 'error':
                notification.error({
                    ...defaultAntdField,
                    icon: <img src={errorIcon} alt="error" />,
                    description: message,
                    duration: duration
                });
                break;
            case 'success':
                notification.success({
                    ...defaultAntdField,
                    icon: <img src={successIcon} alt="success" />,
                    description: message,
                    duration: duration
                });
                break;
            default:
                break;
        }
    };

    useEffect(() => {
        stateRef.current[1].forEach((notification) => {
            if (!displayedNotifications.includes(notification.id) && !notification.read) {
                notificationHandler(notification.id, notification.message_type, notification.message_payload, 0);
                setDisplayedNotifications((prevDisplayedNotifications) => [...prevDisplayedNotifications, notification.id]);
            }
        });
    }, [stateRef.current[1]]);

    const displaySystemMessage = () => {
        return (
            <div className={`system-notification ${systemMessage?.length > 0 ? 'show-notification' : 'hide-notification'}`}>
                <div className="notification-wrapper">
                    {systemMessage[0]?.badge && (
                        <div className="notification-badge">
                            <span>{systemMessage[0]?.badge}</span>
                        </div>
                    )}
                    <p>{systemMessage[0]?.message_payload}</p>
                    {systemMessage[0]?.link_url && (
                        <a className="a-link" href={systemMessage[0]?.link_url} target="_blank" rel="noreferrer">
                            {systemMessage[0]?.link_content}
                            <img src={redirectIcon} alt="redirectIcon" />
                        </a>
                    )}
                </div>
            </div>
        );
    };

    return (
        <div className="app-container">
            {!cloudLogedIn && <Loader />}
            {systemMessage?.length > 0 && displaySystemMessage()}
            <div>
                {' '}
                {!authCheck &&
                    cloudLogedIn &&
                    (!isCloud() ? (
                        <Switch>
                            <Route exact path={pathDomains.signup} component={Signup} />
                            <Route exact path={pathDomains.login} component={Login} />
                            <PrivateRoute
                                exact
                                path={pathDomains.overview}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Overview />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={pathDomains.stations}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <StationsList />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.stations}/:id`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <StationOverview />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.stations}/:id/:id`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <MessageJourney />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            {/* <PrivateRoute
                                exact
                                path={`${pathDomains.functions}`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Functions />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.functions}/:name`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Functions />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            /> */}
                            <PrivateRoute
                                exact
                                path={`${pathDomains.schemaverse}/create`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <SchemaManagment />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.schemaverse}/list`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <SchemaManagment />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.schemaverse}/list/:name`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <SchemaManagment />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={pathDomains.users}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Users />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.sysLogs}`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <SysLogs />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute exact path={pathDomains.profile} component={<AppWrapper content={<Profile />}></AppWrapper>} />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.administration}/integrations`}
                                component={<AppWrapper content={<Administration step={'integrations'} />}></AppWrapper>}
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.administration}/cluster_configuration`}
                                component={<AppWrapper content={<Administration step={'cluster_configuration'} />}></AppWrapper>}
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.administration}/version_upgrade`}
                                component={<AppWrapper content={<Administration step={'version_upgrade'} />}></AppWrapper>}
                            />
                            <PrivateRoute
                                path="/"
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Overview />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <Route>
                                <Redirect to={pathDomains.overview} />
                            </Route>
                        </Switch>
                    ) : (
                        <Switch>
                            <PrivateRoute
                                exact
                                path={pathDomains.overview}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Overview />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={pathDomains.stations}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <StationsList />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.stations}/:id`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <StationOverview />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.stations}/:id/:id`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <MessageJourney />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            {/* <PrivateRoute
                                exact
                                path={`${pathDomains.functions}`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Functions />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.functions}/:name`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Functions />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            /> */}
                            <PrivateRoute
                                exact
                                path={`${pathDomains.schemaverse}/create`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <SchemaManagment />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.schemaverse}/list`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <SchemaManagment />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.schemaverse}/list/:name`}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <SchemaManagment />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <PrivateRoute
                                exact
                                path={pathDomains.users}
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Users />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />

                            <PrivateRoute exact path={pathDomains.profile} component={<AppWrapper content={<Profile />}></AppWrapper>} />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.administration}/integrations`}
                                component={<AppWrapper content={<Administration step={'integrations'} />}></AppWrapper>}
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.administration}/cluster_configuration`}
                                component={<AppWrapper content={<Administration step={'cluster_configuration'} />}></AppWrapper>}
                            />
                            <PrivateRoute
                                exact
                                path={`${pathDomains.administration}/usage`}
                                component={<AppWrapper content={<Administration step={'usage'} />}></AppWrapper>}
                            />
                            {/* <PrivateRoute
                            exact
                            path={`${pathDomains.administration}/payments`}
                            component={<AppWrapper content={<Administration step={'payments'} />}></AppWrapper>}
                        /> */}
                            <PrivateRoute
                                path="/"
                                component={
                                    <AppWrapper
                                        content={
                                            <div>
                                                <Overview />
                                            </div>
                                        }
                                    ></AppWrapper>
                                }
                            />
                            <Route>
                                <Redirect to={pathDomains.overview} />
                            </Route>
                        </Switch>
                    ))}
            </div>
        </div>
    );
});

export default App;
