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
import React, { useState, useEffect, useContext } from 'react';
import { StationStoreContext } from '../../domain/stationOverview';
import Switcher from '../switcher';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';

const DlsConfig = () => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [dlsTypes, setDlsTypes] = useState({
        poison: stationState?.stationSocketData?.dls_configuration_poison,
        schemaverse: stationState?.stationSocketData?.dls_configuration_schemaverse
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
                    <p className="sub-header-dls">Messages that exceeded the maximum delivery attempts</p>
                </div>
                <Switcher onChange={() => updateDlsConfigurations(true, false)} checked={dlsTypes?.poison} loading={dlsLoading?.poison} />
            </div>
            <div className="toggle-dls-config">
                <div>
                    <p className="header-dls">Schema violation</p>
                    <p className="sub-header-dls">Messages that failed schema validation</p>
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
