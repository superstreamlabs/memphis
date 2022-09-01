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

import './App.scss';

import { Switch, Route, withRouter } from 'react-router-dom';
import React, { useContext, useEffect, useState } from 'react';
import { useMediaQuery } from 'react-responsive';
import io from 'socket.io-client';

import { LOCAL_STORAGE_TOKEN } from './const/localStorageConsts';
import { HANDLE_REFRESH_INTERVAL, SOCKET_URL } from './config';
import { handleRefreshTokenRequest } from './services/http';
import StationOverview from './domain/stationOverview';
import MessageJourney from './domain/messageJourney';
import FactoriesList from './domain/factoriesList';
import AppWrapper from './components/appWrapper';
import StationsList from './domain/stationsList';
import SandboxLogin from './domain/sandboxLogin';
import { useHistory } from 'react-router-dom';
import { Redirect } from 'react-router-dom';
import PrivateRoute from './PrivateRoute';
import Overview from './domain/overview';
import Settings from './domain/settings';
import { Context } from './hooks/store';
import SysLogs from './domain/sysLogs';
import pathDomains from './router';
import Users from './domain/users';
import Login from './domain/login';

const Desktop = ({ children }) => {
    const isDesktop = useMediaQuery({ minWidth: 850 });
    return isDesktop ? children : null;
};

const Mobile = ({ children }) => {
    const isMobile = useMediaQuery({ maxWidth: 849 });
    return isMobile ? children : null;
};

const App = withRouter(() => {
    const [authCheck, setAuthCheck] = useState(true);
    const [state, dispatch] = useContext(Context);

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
                    const socket = await io.connect(SOCKET_URL, {
                        path: '/api/socket.io',
                        query: {
                            authorization: localStorage.getItem(LOCAL_STORAGE_TOKEN)
                        },
                        reconnection: false
                    });
                    dispatch({ type: 'SET_SOCKET_DETAILS', payload: socket });
                }
                return true;
            }
        } else {
            history.push('/login');
        }
    };

    return (
        <div className="app-container">
            <div>
                {' '}
                {!authCheck && (
                    <Switch>
                        <Route exact path={pathDomains.login} component={process.env.REACT_APP_SANDBOX_ENV ? SandboxLogin : Login} />
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
                            path={pathDomains.settings}
                            component={
                                <AppWrapper
                                    content={
                                        <div>
                                            <Settings />
                                        </div>
                                    }
                                ></AppWrapper>
                            }
                        />
                        <PrivateRoute
                            exact
                            path={pathDomains.factoriesList}
                            component={
                                <AppWrapper
                                    content={
                                        <div>
                                            <FactoriesList />
                                        </div>
                                    }
                                ></AppWrapper>
                            }
                        />
                        <PrivateRoute
                            exact
                            path={`${pathDomains.factoriesList}/:id`}
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
                            path={`${pathDomains.factoriesList}/:id/:id`}
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
                            path={`${pathDomains.factoriesList}/:id/:id/:id`}
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
