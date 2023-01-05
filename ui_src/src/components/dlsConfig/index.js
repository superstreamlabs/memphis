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
import './style.scss';
import React, { useState, useEffect, useContext } from 'react';
import { StationStoreContext } from '../../domain/stationOverview';
import Switcher from '../switcher';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';

const DlsConfig = () => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [dlsTypes, setDlsTypes] = useState({
        poison: stationState?.stationSocketData?.dls_configuration?.poison,
        schemaverse: stationState?.stationSocketData?.dls_configuration?.schemaverse
    });

    const [dlsLoading, setDlsLoading] = useState({
        poison: false,
        schemaverse: false
    });

    const updateDlsConfigurations = async (poison = false, schema = false) => {
        setDlsLoading({
            poison: poison,
            schemaverse: schema
        });
        const conf = {
            station_name: stationState?.stationMetaData?.name,
            poison: poison ? !dlsTypes?.poison : dlsTypes?.poison,
            schemaverse: schema ? !dlsTypes?.schemaverse : dlsTypes?.schemaverse
        };
        poison && setDlsTypes({ ...dlsTypes, poison: !dlsTypes?.poison });
        schema && setDlsTypes({ ...dlsTypes, schemaverse: !dlsTypes?.schemaverse });

        try {
            await httpRequest('PUT', ApiEndpoints.UPDATE_DLS_CONFIGURATION, conf);
            setDlsLoading({
                poison: false,
                schemaverse: false
            });
        } catch (error) {
            setDlsLoading({
                poison: false,
                schemaverse: false
            });
        }
    };

    return (
        <div className="dls-config-container">
            <div className="toggle-dls-config">
                <div>
                    <p className="header-dls">Unacknowledged</p>
                    <p className="sub-header-dls">Unacknowledged messages that passed "maxMsgDeliveries"</p>
                </div>
                <Switcher onChange={() => updateDlsConfigurations(true, false)} checked={dlsTypes?.poison} loading={dlsLoading?.poison} />
            </div>
            <div className="toggle-dls-config">
                <div>
                    <p className="header-dls">Schema violation</p>
                    <p className="sub-header-dls">Messages that did not pass schema validation</p>
                </div>
                <Switcher
                    disabled={!stationState?.stationMetaData?.is_native}
                    onChange={() => updateDlsConfigurations(false, true)}
                    checked={dlsTypes?.schemaverse}
                    tooltip={!stationState?.stationMetaData?.is_native && 'Supported only by using Memphis SDKs'}
                    loading={dlsLoading?.schemaverse}
                />
            </div>
        </div>
    );
};
export default DlsConfig;
