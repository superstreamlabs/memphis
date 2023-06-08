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
import React, { useContext, useEffect, useState } from 'react';
import { useMediaQuery } from 'react-responsive';
import { connect } from 'nats.ws';
import { message } from 'antd';

import { LOCAL_STORAGE_ACCOUNT_ID, LOCAL_STORAGE_CONNECTION_TOKEN, LOCAL_STORAGE_TOKEN, LOCAL_STORAGE_USER_PASS_BASED_AUTH, LOCAL_STORAGE_WS_PORT } from './const/localStorageConsts';
import { ENVIRONMENT, HANDLE_REFRESH_INTERVAL, WS_PREFIX, WS_SERVER_URL_PRODUCTION } from './config';
import { handleRefreshTokenRequest } from './services/http';
import StationOverview from './domain/stationOverview';
import MessageJourney from './domain/messageJourney';
import { is_cloud } from './services/valueConvertor';
import Administration from './domain/administration';
import AppWrapper from './components/appWrapper';
import StationsList from './domain/stationsList';
import SchemaManagment from './domain/schema';
import { useHistory } from 'react-router-dom';
import { Redirect } from 'react-router-dom';
import PrivateRoute from './PrivateRoute';
import Overview from './domain/overview';
import { Context } from './hooks/store';
import Profile from './domain/profile';
import Signup from './domain/signup';
import pathDomains from './router';
import Users from './domain/users';
import Login from './domain/login';

const SysLogs = undefined;
if (!is_cloud()) {
    SysLogs = require('./domain/sysLogs').default;
}

const App = withRouter(() => {
    const [state, dispatch] = useContext(Context);
    const isMobile = useMediaQuery({ maxWidth: 849 });
    const [authCheck, setAuthCheck] = useState(true);

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

    const history = useHistory();

    useEffect(async () => {
        await handleRefresh(true);
        setAuthCheck(false);

        const interval = setInterval(() => {
            handleRefresh(false);
        }, HANDLE_REFRESH_INTERVAL);

        return () => {
            clearInterval(interval);
            state.socket?.close();
        };
    }, []);

    const handleRefresh = async (firstTime) => {
        if (window.location.pathname === pathDomains.login) {
            return;
        } else if (localStorage.getItem(LOCAL_STORAGE_TOKEN)) {
            const ws_port = localStorage.getItem(LOCAL_STORAGE_WS_PORT);
            const SOCKET_URL = ENVIRONMENT === 'production' ? `${WS_PREFIX}://${WS_SERVER_URL_PRODUCTION}:${ws_port}` : `${WS_PREFIX}://localhost:${ws_port}`;
            const handleRefreshStatus = await handleRefreshTokenRequest();
            if (handleRefreshStatus) {
                if (firstTime) {
                    try {
                        let conn;
                        const connection_token = localStorage.getItem(LOCAL_STORAGE_CONNECTION_TOKEN);
                        if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                            const account_id = localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)
                            conn = await connect({
                                servers: [SOCKET_URL],
                                user: '$memphis_user$' + account_id,
                                pass: connection_token,
                                timeout: '5000'
                            });
                        } else {
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
            history.push(pathDomains.signup);
        }
    };

    return (
        <div className="app-container">
            <div>
                {' '}
                {!authCheck && (
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

                        {!is_cloud() && (
                            <>
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
                            </>
                        )}
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
                )}
            </div>
        </div>
    );
});

export default App;
