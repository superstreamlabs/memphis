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

import './style.scss';

import React, { useEffect, useContext, useState, createContext, useReducer, useCallback } from 'react';

import StationOverviewHeader from './stationOverviewHeader';
import StationObservabilty from './stationObservabilty';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import { Context } from '../../hooks/store';
import pathDomains from '../../router';

import Reducer from './hooks/reducer';
import Loader from '../../components/loader';
import { useHistory } from 'react-router-dom';
import { parsingDate } from '../../services/valueConvertor';

const StationOverview = () => {
    const [stationState, stationDispatch] = useReducer(Reducer);
    const url = window.location.href;
    const stationName = url.split('factories/')[1].split('/')[1];
    const history = useHistory();
    const [state, dispatch] = useContext(Context);
    const [isLoading, setisLoading] = useState(false);

    const getStaionMetaData = async () => {
        try {
            let data = await httpRequest('GET', `${ApiEndpoints.GET_STATION}?station_name=${stationName}`);
            data.creation_date = await parsingDate(data.creation_date);
            stationDispatch({ type: 'SET_STATION_META_DATA', payload: data });
        } catch (error) {
            if (error.status === 404) {
                history.push(`${pathDomains.factoriesList}/${url.split('factories/')[1].split('/')[0]}`);
            }
        }
    };

    const sortData = (data) => {
        data.audit_logs?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.messages?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.active_producers?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.active_consumers?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.destroyed_consumers?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.destroyed_producers?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.killed_consumers?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        data.killed_producers?.sort((a, b) => new Date(b.creation_date) - new Date(a.creation_date));
        return data;
    };

    const getStationDetails = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_STATION_DATA}?station_name=${stationName}`);
            await sortData(data);
            stationDispatch({ type: 'SET_SOCKET_DATA', payload: data });
            setisLoading(false);
        } catch (error) {
            setisLoading(false);
            if (error.status === 404) {
                history.push(`${pathDomains.factoriesList}/${url.split('factories/')[1].split('/')[0]}`);
            }
        }
    };

    useEffect(() => {
        setisLoading(true);
        dispatch({ type: 'SET_ROUTE', payload: 'factories' });
        getStaionMetaData();
        getStationDetails();
    }, []);

    useEffect(() => {
        state.socket?.on(`station_overview_data_${stationName}`, (data) => {
            sortData(data);
            stationDispatch({ type: 'SET_SOCKET_DATA', payload: data });
        });

        setTimeout(() => {
            state.socket?.emit('register_station_overview_data', stationName);
        }, 1000);
        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

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
                            <StationOverviewHeader />
                        </div>
                        <div className="station-observability">
                            <StationObservabilty />
                        </div>
                    </div>
                )}
            </React.Fragment>
        </StationStoreContext.Provider>
    );
};
export const StationStoreContext = createContext({});
export default StationOverview;
