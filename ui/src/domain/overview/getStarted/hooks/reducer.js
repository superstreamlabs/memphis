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

const Reducer = (getStartedState, action) => {
    switch (action.type) {
        case 'SET_NEXT_DISABLE':
            return {
                ...getStartedState,
                nextDisable: action.payload
            };
        case 'SET_BACK_DISABLE':
            return {
                ...getStartedState,
                backDisable: action.payload
            };
        case 'SET_FACTORY':
            return {
                ...getStartedState,
                factoryName: action.payload
            };
        case 'SET_STATION':
            return {
                ...getStartedState,
                stationName: action.payload
            };
        case 'SET_USER_NAME':
            return {
                ...getStartedState,
                username: action.payload
            };
        case 'SET_CURRENT_STEP':
            return {
                ...getStartedState,
                currentStep: action.payload
            };
        case 'SET_COMPLETED_STEPS':
            return {
                ...getStartedState,
                completedSteps: action.payload
            };
        case 'SET_CREATE_APP_USER_DISABLE':
            return {
                ...getStartedState,
                createAppUserDisable: action.payload
            };
        case 'SET_FORM_FIELDS_CREATE_STATION':
            let formFieldsChanges = getStartedState.formFieldsCreateStation;
            formFieldsChanges[action.payload.field] = action.payload.value;
            return {
                ...getStartedState,
                formFieldsCreateStation: formFieldsChanges
            };
        case 'SET_BROKER_CONNECTION_CREDS':
            return {
                ...getStartedState,
                connectionCreds: action.payload
            };
        case 'GET_STATION_DATA':
            return {
                ...getStartedState,
                stationData: action.payload
            };
        case 'IS_APP_USER_CREATED':
            return {
                ...getStartedState,
                isAppUserCreated: action.payload
            };
        case 'IS_LOADING':
            return {
                ...getStartedState,
                isLoading: action.payload
            };
        case 'SET_HIDDEN_BUTTON':
            return {
                ...getStartedState,
                isHiddenButton: action.payload
            };
        case 'SET_DESIRED_PODS':
            return {
                ...getStartedState,
                desiredPods: action.payload
            };
        default:
            return getStartedState;
    }
};

export default Reducer;
