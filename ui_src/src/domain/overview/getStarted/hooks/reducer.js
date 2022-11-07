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

const initialState = {
    currentStep: 1,
    completedSteps: 0,
    formFieldsCreateStation: {
        name: '',
        retention_type: 'message_age_sec',
        retention_value: 604800,
        storage_type: 'file',
        replicas: 1,
        days: 7,
        hours: 0,
        minutes: 0,
        seconds: 0,
        retentionSizeValue: '1000',
        retentionMessagesValue: '10'
    },
    nextDisable: false,
    isLoading: false,
    isHiddenButton: false,
    actualPods: null
};

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
        case 'INITIAL_STATE':
            return {
                getStartedState: initialState
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
        case 'SET_ACTUAL_PODS':
            return {
                ...getStartedState,
                actualPods: action.payload
            };
        default:
            return getStartedState;
    }
};

export default Reducer;
