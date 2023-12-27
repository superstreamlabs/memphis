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

const Reducer = (state, action) => {
    let index;
    let updateSchemaListState = state?.schemaList;
    let copyIntegration = state?.integrationsList;
    let newUserData = state.userData;
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
        case 'IS_LATEST':
            return {
                ...state,
                isLatest: action.payload
            };
        case 'CURRENT_VERSION':
            return {
                ...state,
                currentVersion: action.payload
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
            newUserData.avatar_id = action.payload;
            return {
                ...state,
                userData: newUserData
            };
        case 'SET_ENTITLEMENTS':
            newUserData.entitlements = action.payload;
            return {
                ...state,
                userData: newUserData
            };
        case 'SET_PLAN_TYPE':
            return {
                ...state,
                isFreePlan: action.payload
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
        case 'SET_STATION_LIST':
            return {
                ...state,
                stationList: action.payload
            };
        case 'SET_STATION_FILTERED_LIST':
            return {
                ...state,
                stationFilteredList: action.payload
            };
        case 'SET_SCHEMA_LIST':
            return {
                ...state,
                schemaList: action.payload
            };
        case 'SET_SCHEMA_FILTERED_LIST':
            return {
                ...state,
                schemaFilteredList: action.payload
            };
        case 'SET_FILTERED_OPTION':
            return {
                ...state,
                FilterOption: action.payload
            };

        case 'SET_SCHEMA_TAGS':
            index = state?.schemaList?.findIndex((schema) => schema.name === action.payload?.schemaName);
            updateSchemaListState[index].tags = action.payload.tags;
            return {
                ...state,
                schemaList: updateSchemaListState
            };
        case 'SET_IS_USED':
            index = state?.schemaList?.findIndex((schema) => schema.name === action.payload?.schemaName);
            updateSchemaListState[index].used = true;
            return {
                ...state,
                schemaList: updateSchemaListState
            };
        case 'SET_UPDATE_SCHEMA':
            index = state?.schemaList?.findIndex((schema) => schema.name === action.payload?.schemaName);
            updateSchemaListState[index] = action.payload.schemaDetails;
            return {
                ...state,
                schemaList: updateSchemaListState
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
        case 'SET_ALLOWED_ACTIONS':
            return {
                ...state,
                allowedActions: action.payload
            };
        case 'SET_DARK_MODE':
            return {
                ...state,
                darkMode: action.payload
            };
        default:
            return state;
    }
};

export default Reducer;
