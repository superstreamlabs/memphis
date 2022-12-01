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

import React, { createContext, useReducer } from 'react';

import Reducer from './reducer';

const initialState = {
    userData: {
        user_id: '',
        already_logged_in: false,
        creation_date: '',
        user_type: '',
        avatar_id: 1
    },
    companyLogo: '',
    monitor_data: {},
    loading: false,
    error: null,
    route: '',
    isAuthentication: false,
    analytics_modal: true,
    socket: null,
    skipSignup: false,
    createSchema: false,
    domainList: [],
    filteredList: [],
    integrationsList: []
};

const Store = ({ children }) => {
    const [state, dispatch] = useReducer(Reducer, initialState);

    return <Context.Provider value={[state, dispatch]}>{children}</Context.Provider>;
};

export const Context = createContext(initialState);
export default Store;
