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

import './App.scss';

import { Switch, Route, withRouter } from 'react-router-dom';
import React, { useContext, useEffect, useState } from 'react';
import { useMediaQuery } from 'react-responsive';
import { connect } from 'nats.ws';
import { message } from 'antd';

import { LOCAL_STORAGE_TOKEN } from './const/localStorageConsts';
import { HANDLE_REFRESH_INTERVAL, SOCKET_URL } from './config';
import { handleRefreshTokenRequest } from './services/http';
import StationOverview from './domain/stationOverview';
import MessageJourney from './domain/messageJourney';
import AppWrapper from './components/appWrapper';
import StationsList from './domain/stationsList';
import SandboxLogin from './domain/sandboxLogin';
import { useHistory } from 'react-router-dom';
import SchemaManagment from './domain/schema';
import { Redirect } from 'react-router-dom';
import PrivateRoute from './PrivateRoute';
import Overview from './domain/overview';
import { Context } from './hooks/store';
import SysLogs from './domain/sysLogs';
import Signup from './domain/signup';
import pathDomains from './router';
import Users from './domain/users';
import Login from './domain/login';
import Preferences from './domain/preferences';

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
            const handleRefreshStatus = await handleRefreshTokenRequest();
            if (handleRefreshStatus) {
                if (firstTime) {
                    try {
                        const conn = await connect({
                            servers: [SOCKET_URL],
                            token: 'memphis'
                        });
                        dispatch({ type: 'SET_SOCKET_DETAILS', payload: conn });
                    } catch (error) {}
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
                        {process.env.REACT_APP_SANDBOX_ENV && <Route exact path={pathDomains.login} component={SandboxLogin} />}
                        {!process.env.REACT_APP_SANDBOX_ENV && <Route exact path={pathDomains.signup} component={Signup} />}
                        {!process.env.REACT_APP_SANDBOX_ENV && <Route exact path={pathDomains.login} component={Login} />}
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
                        <PrivateRoute
                            exact
                            path={`${pathDomains.preferences}/profile`}
                            component={<AppWrapper content={<Preferences step={'profile'} />}></AppWrapper>}
                        />
                        <PrivateRoute
                            exact
                            path={`${pathDomains.preferences}/integrations`}
                            component={<AppWrapper content={<Preferences step={'integrations'} />}></AppWrapper>}
                        />
                        <PrivateRoute
                            exact
                            path={`${pathDomains.preferences}/cluster_configuration`}
                            component={<AppWrapper content={<Preferences step={'cluster_configuration'} />}></AppWrapper>}
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
                            path={pathDomains.schemaverse}
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
                            path={`${pathDomains.schemaverse}/:name`}
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
