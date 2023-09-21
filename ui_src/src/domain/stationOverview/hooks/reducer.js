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

const Reducer = (stationState, action) => {
    let updatedSocketState = stationState?.stationSocketData;
    let updatedMetdaDataState = stationState?.stationMetaData;
    switch (action.type) {
        case 'SET_STATION_META_DATA':
            return {
                ...stationState,
                stationMetaData: action.payload
            };
        case 'SET_STATION_PARTITION':
            return {
                ...stationState,
                stationPartition: action.payload
            };
        case 'SET_SOCKET_DATA':
            return {
                ...stationState,
                stationSocketData: action.payload
            };
        case 'SET_MESSAGES':
            updatedSocketState.messages = action.payload;
            return {
                ...stationState,
                stationSocketData: updatedSocketState
            };
        case 'SET_POISON_MESSAGES':
            updatedSocketState.poison_messages = action.payload;
            return {
                ...stationState,
                stationSocketData: updatedSocketState
            };
        case 'SET_FAILED_MESSAGES':
            updatedSocketState.schema_failed_messages = action.payload;
            return {
                ...stationState,
                stationSocketData: updatedSocketState
            };
        case 'SET_TAGS':
            updatedSocketState.tags = action.payload;
            return {
                ...stationState,
                stationSocketData: updatedSocketState
            };
        case 'SET_SCHEMA':
            updatedSocketState.schema = action.payload;
            return {
                ...stationState,
                stationSocketData: updatedSocketState
            };
        case 'SET_DLS':
            updatedMetdaDataState.dls_station = action.payload;
            return {
                ...stationState,
                stationMetaData: updatedMetdaDataState
            };
        case 'SET_SELECTED_ROW_ID':
            return {
                ...stationState,
                selectedRowId: action.payload
            };
        case 'SET_SELECTED_ROW_PARTITION':
            return {
                ...stationState,
                selectedRowPartition: action.payload
            };
        case 'SET_SCHEMA_TYPE':
            return {
                ...stationState,
                schemaType: action.payload
            };

        default:
            return stationState;
    }
};

export default Reducer;
