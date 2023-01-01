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

const Reducer = (stationState, action) => {
    let updatedSocketState = stationState?.stationSocketData;
    switch (action.type) {
        case 'SET_STATION_META_DATA':
            return {
                ...stationState,
                stationMetaData: action.payload
            };
        case 'SET_SOCKET_DATA':
            return {
                ...stationState,
                stationSocketData: action.payload
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
        case 'SET_SELECTED_ROW_ID':
            return {
                ...stationState,
                selectedRowId: action.payload
            };
        case 'SET_DLS_TYPE':
            return {
                ...stationState,
                dlsType: action.payload
            };
        default:
            return stationState;
    }
};

export default Reducer;
