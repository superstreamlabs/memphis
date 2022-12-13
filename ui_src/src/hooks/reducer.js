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

const Reducer = (state, action) => {
    let index;
    let updateDomainState = state?.domainList;
    let copyIntegration = state?.integrationsList;
    switch (action.type) {
        case 'SET_USER_DATA':
            return {
                ...state,
                userData: action.payload
            };
        case 'SET_COMPANY_LOGO':
            return {
                ...state,
                companyLogo: action.payload
            };
        case 'SET_LOADER':
            return {
                ...state,
                loading: action.payload
            };
        case 'SET_ROUTE':
            return {
                ...state,
                route: action.payload
            };
        case 'SET_AUTHENTICATION':
            return {
                ...state,
                isAuthentication: action.payload
            };
        case 'ANALYTICS_MODAL':
            return {
                ...state,
                analytics_modal: action.payload
            };
        case 'SET_MONITOR_DATA':
            return {
                ...state,
                monitor_data: action.payload
            };
        case 'SET_AVATAR_ID':
            let newUserData = state.userData;
            newUserData.avatar_id = action.payload;
            return {
                ...state,
                userData: newUserData
            };
        case 'SET_SOCKET_DETAILS':
            return {
                ...state,
                socket: action.payload
            };
        case 'SKIP_SIGNUP':
            return {
                ...state,
                skipSignup: action.payload
            };
        case 'SET_DOMAIN_LIST':
            return {
                ...state,
                domainList: action.payload
            };
        case 'SET_FILTERED_LIST':
            return {
                ...state,
                filteredList: action.payload
            };
        case 'SET_FILTERED_OPTION':
            return {
                ...state,
                FilterOption: action.payload
            };
        case 'SET_CREATE_SCHEMA':
            return {
                ...state,
                createSchema: action.payload
            };

        case 'SET_SCHEMA_TAGS':
            index = state?.domainList?.findIndex((schema) => schema.name === action.payload?.schemaName);
            updateDomainState[index].tags = action.payload.tags;
            return {
                ...state,
                domainList: updateDomainState
            };
        case 'SET_IS_USED':
            index = state?.domainList?.findIndex((schema) => schema.name === action.payload?.schemaName);
            updateDomainState[index].used = true;
            return {
                ...state,
                domainList: updateDomainState
            };
        case 'SET_UPDATE_SCHEMA':
            index = state?.domainList?.findIndex((schema) => schema.name === action.payload?.schemaName);
            updateDomainState[index] = action.payload.schemaDetails;
            return {
                ...state,
                domainList: updateDomainState
            };
        case 'SET_INTEGRATIONS':
            return {
                ...state,
                integrationsList: action.payload
            };
        case 'REMOVE_INTEGRATION':
            index = state?.integrationsList?.findIndex((integration) => integration.name === action.payload);
            copyIntegration.splice(index, 1);
            return {
                ...state,
                integrationsList: copyIntegration
            };
        case 'ADD_INTEGRATION':
            copyIntegration = [...copyIntegration, action.payload];
            return {
                ...state,
                integrationsList: copyIntegration
            };
        case 'UPDATE_INTEGRATION':
            index = state?.integrationsList?.findIndex((integration) => integration.name === action.payload.name);
            copyIntegration[index] = action.payload;
            return {
                ...state,
                integrationsList: copyIntegration
            };
        case 'SET_LOG_FILTER':
            return {
                ...state,
                logsFilter: action.payload
            };
        default:
            return state;
    }
};

export default Reducer;
