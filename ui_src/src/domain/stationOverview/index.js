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

import './style.scss';

import React, { useEffect, useContext, useState, createContext, useReducer } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import { extractValueFromURL, parsingDate } from 'services/valueConvertor';
import StationOverviewHeader from './stationOverviewHeader';
import StationObservabilty from './stationObservabilty';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import Loader from 'components/loader';
import { Context } from 'hooks/store';
import pathDomains from 'router';
import Reducer from './hooks/reducer';

const initializeState = {
    stationMetaData: { is_native: true },
    stationSocketData: {},
    stationPartition: null,
    stationFunctions: {}
};
let sub;
const StationOverview = () => {
    const [stationState, stationDispatch] = useReducer(Reducer);
    const url = window.location.href;
    const stationName = extractValueFromURL(url, 'name');
    const history = useHistory();
    const [state, dispatch] = useContext(Context);
    const [isLoading, setisLoading] = useState(false);
    const [isReloading, setisReloading] = useState(false);
    const location = useLocation();
    const [referredFunction, setReferredFunction] = useState(null);

    useEffect(() => {
        location?.selectedFunction && setReferredFunction(location?.selectedFunction);
    }, [location?.selectedFunction]);

    const sortData = (data) => {
        data.audit_logs?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.messages?.sort((a, b) => b.message_seq - a.message_seq);
        data.active_producers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.active_consumers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.destroyed_consumers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.destroyed_producers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.killed_consumers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.killed_producers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        return data;
    };

    const getStaionMetaData = async () => {
        try {
            let data = await httpRequest('GET', `${ApiEndpoints.GET_STATION}?station_name=${stationName}`);
            data.created_at = await parsingDate(data.created_at);
            stationDispatch({ type: 'SET_STATION_META_DATA', payload: data });
            stationDispatch({ type: 'SET_STATION_PARTITION', payload: data?.partitions_list ? 1 : -1 });
        } catch (error) {
            if (error.status === 404) {
                history.push(pathDomains.stations);
            }
        }
    };

    const getStationDetails = async () => {
        setisReloading(true);
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_STATION_DATA}?station_name=${stationName}&partition_number=${stationState?.stationPartition || 1}`);
            await sortData(data);
            stationDispatch({ type: 'SET_SOCKET_DATA', payload: data });
            stationDispatch({ type: 'SET_SCHEMA_TYPE', payload: data.schema.schema_type });
            setisLoading(false);
            setisReloading(false);
        } catch (error) {
            setisLoading(false);
            setisReloading(false);
            if (error.status === 404) {
                history.push(pathDomains.stations);
            }
        }
    };

    useEffect(() => {
        stationState?.stationPartition && getStationDetails();
    }, [stationState?.stationPartition]);

    useEffect(() => {
        setisLoading(true);
        dispatch({ type: 'SET_ROUTE', payload: 'stations' });
        getStaionMetaData();
    }, [stationName]);

    return (
        <StationStoreContext.Provider value={[stationState, stationDispatch]}>
            <React.Fragment>
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {!isLoading && (
                    <div className="station-overview-container">
                        <div className="overview-header">
                            <StationOverviewHeader refresh={() => getStationDetails()} />
                        </div>
                        <div className="station-observability">
                            <StationObservabilty referredFunction={referredFunction} loading={isReloading} />
                        </div>
                    </div>
                )}
            </React.Fragment>
        </StationStoreContext.Provider>
    );
};

export const StationStoreContext = createContext({ initializeState });
export default StationOverview;
