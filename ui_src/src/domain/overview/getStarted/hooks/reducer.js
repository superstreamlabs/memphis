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

const initialState = {
    currentStep: 1,
    completedSteps: 0,
    formFieldsCreateStation: {
        name: '',
        retention_type: 'message_age_sec',
        retention_value: 3600,
        storage_type: 'file',
        replicas: 1,
        days: 1,
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
        case 'SET_USER':
            return {
                ...getStartedState,
                username: action.payload.username,
                password: action.payload.password
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
