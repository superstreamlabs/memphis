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

const Reducer = (state, action) => {
    let index;
    let updateState = state?.domainList;
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
            updateState[index].tags = action.payload.tags;
            return {
                ...state,
                domainList: updateState
            };
        case 'SET_UPDATE_SCHEMA':
            index = state?.domainList?.findIndex((schema) => schema.name === action.payload?.schemaName);
            updateState[index] = action.payload.schemaDetails;
            return {
                ...state,
                domainList: updateState
            };

        default:
            return state;
    }
};

export default Reducer;
