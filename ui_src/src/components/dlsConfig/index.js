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

import TitleComponent from '../titleComponent';
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

    useEffect(() => {
        updateDlsConfigurations();
    }, [dlsTypes]);

    useEffect(() => {
        setDlsTypes({
            poison: stationState?.stationSocketData?.dls_configuration?.poison,
            schemaverse: stationState?.stationSocketData?.dls_configuration?.schemaverse
        });
    }, [stationState?.stationSocketData?.dls_configuration]);

    const updateDlsConfigurations = async () => {
        const conf = {
            station_name: stationState?.stationMetaData?.name,
            poison: dlsTypes.poison,
            schemaverse: dlsTypes.schemaverse
        };
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

    const handlePoisonChange = () => {
        setDlsLoading({
            poison: true,
            schemaverse: false
        });
        setDlsTypes({ ...dlsTypes, poison: !dlsTypes.poison });
    };
    const handleSchemaChange = () => {
        setDlsLoading({
            poison: false,
            schemaverse: true
        });
        setDlsTypes({ ...dlsTypes, schemaverse: !dlsTypes.schemaverse });
    };

    return (
        <div className="dls-config-container">
            <div className="toggle-dls-config">
                <TitleComponent headerTitle="Poison" typeTitle="sub-header" headerDescription="Contrary to popular belief, Lorem Ipsum is not " />
                <Switcher onChange={handlePoisonChange} checked={dlsTypes?.poison} loading={dlsLoading.poison} />
            </div>
            <div className="toggle-dls-config">
                <TitleComponent headerTitle="Schemaverse" typeTitle="sub-header" headerDescription="Contrary to popular belief, Lorem Ipsum is not " />
                <Switcher
                    disabled={!stationState?.stationMetaData.is_native}
                    onChange={handleSchemaChange}
                    checked={dlsTypes?.schemaverse}
                    tooltip={!stationState?.stationMetaData.is_native && 'Not supported without using the native Memphis SDK’s'}
                    loading={dlsLoading.schemaverse}
                />
            </div>
        </div>
    );
};
export default DlsConfig;
