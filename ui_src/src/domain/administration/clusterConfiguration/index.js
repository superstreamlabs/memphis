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

import React, { useEffect, useState } from 'react';

import { compareObjects, isCloud } from '../../../services/valueConvertor';
import BrokerHostname from '../../../assets/images/BrokerHostname.svg';
import UIHostname from '../../../assets/images/UIHostname.svg';
import DeadLetterInHours from '../../../assets/images/DeadLetterInHours.svg';
import LogsRetentionInDays from '../../../assets/images/LogsRetentionInDays.svg';
import RestHostname from '../../../assets/images/RestHostname.svg';
import TieredStorageInterval from '../../../assets/images/TieredStorageInterval.svg';

import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import Button from '../../../components/button';
import SliderRow from './components/sliderRow';
import InputRow from './components/inputRow';
import TieredInputRow from './components/tieredInputRow';
import { message } from 'antd';
import {
    LOCAL_STORAGE_BROKER_HOST,
    LOCAL_STORAGE_ENV,
    LOCAL_STORAGE_REST_GW_HOST,
    LOCAL_STORAGE_UI_HOST,
    LOCAL_STORAGE_TIERED_STORAGE_TIME,
    DEAD_LETTERED_MESSAGES_RETENTION_IN_HOURS,
    TIERED_STORAGE_UPLOAD_INTERVAL,
    LOGS_RETENTION_IN_DAYS
} from '../../../const/localStorageConsts';
import Loader from '../../../components/loader';
import { showMessages } from '../../../services/genericServices';

function ClusterConfiguration() {
    const [formFields, setFormFields] = useState({});
    const [oldValues, setOldValues] = useState({});
    const [isChanged, setIsChanged] = useState(false);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        getConfigurationValue();
    }, []);

    const updateLocalStorage = (data) => {
        localStorage.setItem(DEAD_LETTERED_MESSAGES_RETENTION_IN_HOURS, data.dls_retention);
        localStorage.setItem(LOGS_RETENTION_IN_DAYS, data.logs_retention);
        localStorage.setItem(TIERED_STORAGE_UPLOAD_INTERVAL, data.tiered_storage_time_sec);
        if (!isCloud()) {
            localStorage.setItem(LOCAL_STORAGE_BROKER_HOST, data.broker_host);
            localStorage.setItem(LOCAL_STORAGE_REST_GW_HOST, data.rest_gw_host);
            localStorage.setItem(LOCAL_STORAGE_UI_HOST, data.ui_host);
            localStorage.setItem(LOCAL_STORAGE_TIERED_STORAGE_TIME, data.tiered_storage_time_sec);
        }
    };
    const getConfigurationValue = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_CONFIGURATION);
            updateLocalStorage(data);
            setOldValues(data);
            setFormFields(data);
            setIsLoading(false);
        } catch (err) {
            setIsLoading(false);
            return;
        }
    };

    const updateConfiguration = async () => {
        try {
            const data = await httpRequest('PUT', ApiEndpoints.EDIT_CLUSTER_CONFIGURATION, { ...formFields });
            updateLocalStorage(data);
            setIsChanged(false);
            setOldValues(data);
            showMessages('success', 'Successfully updated');
        } catch (err) {
            return;
        }
    };

    const handleChange = (field, value, err) => {
        if (err !== '' && err !== undefined) {
            setIsChanged(false);
        } else {
            let updatedValue = { ...formFields };
            updatedValue[field] = value;
            setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
            setIsChanged(!compareObjects(updatedValue, oldValues));
        }
    };
    const discardChanges = () => {
        setIsChanged(false);
        setFormFields((formFields) => ({ ...formFields, ...oldValues }));
    };

    return (
        <div className="configuration-container">
            <div className="header">
                <p className="main-header">Environment configuration</p>
                <p className="memphis-label">Customize the internal configuration to match your requirements</p>
            </div>
            {isLoading && <Loader className="loader-container" />}
            {!isLoading && (
                <>
                    <div className="configuration-body">
                        <SliderRow
                            title="DEAD LETTERED MESSAGES RETENTION IN HOURS"
                            desc="Amount of hours to retain dead lettered messages in a DLS"
                            value={formFields?.dls_retention}
                            img={DeadLetterInHours}
                            min={1}
                            max={30}
                            unit={'h'}
                            onChanges={(e) => handleChange('dls_retention', e)}
                        />
                        <SliderRow
                            title="DISCONNECTED PRODUCERS AND CONSUMERS RETENTION"
                            desc="Amount of hours to retain inactive producers and consumers"
                            value={formFields?.gc_producer_consumer_retention_hours}
                            img={DeadLetterInHours}
                            min={1}
                            max={48}
                            unit={'h'}
                            onChanges={(e) => handleChange('gc_producer_consumer_retention_hours', e)}
                        />
                        {!isCloud() && (
                            <>
                                <SliderRow
                                    title="MAX MESSAGE SIZE"
                                    desc="Maximum message size (payload + headers) in megabytes"
                                    value={formFields?.max_msg_size_mb}
                                    img={DeadLetterInHours}
                                    min={1}
                                    max={12}
                                    unit={'mb'}
                                    onChanges={(e) => handleChange('max_msg_size_mb', e)}
                                />
                                <SliderRow
                                    title="LOGS RETENTION IN DAYS"
                                    desc="Amount of days to retain system logs"
                                    img={LogsRetentionInDays}
                                    value={formFields?.logs_retention}
                                    min={1}
                                    max={100}
                                    unit={'d'}
                                    onChanges={(e) => handleChange('logs_retention', e)}
                                />
                                <TieredInputRow
                                    title="TIERED STORAGE UPLOAD INTERVAL"
                                    desc="(if configured) The interval which the broker will migrate a batch of messages to the second storage tier"
                                    img={TieredStorageInterval}
                                    value={formFields?.tiered_storage_time_sec}
                                    onChanges={(e, err) => handleChange('tiered_storage_time_sec', e, err)}
                                />
                            </>
                        )}

                        {localStorage.getItem(LOCAL_STORAGE_ENV) !== 'docker' && !isCloud() && (
                            <>
                                <InputRow
                                    title="BROKER HOSTNAME"
                                    desc={`*For display purpose only*\nWhich URL should be seen as the "broker hostname"`}
                                    img={BrokerHostname}
                                    value={formFields?.broker_host}
                                    onChanges={(e) => handleChange('broker_host', e.target.value)}
                                    placeholder={localStorage.getItem(LOCAL_STORAGE_BROKER_HOST) === undefined ? localStorage.getItem(LOCAL_STORAGE_BROKER_HOST) : ''}
                                />
                                <InputRow
                                    title="UI HOSTNAME"
                                    desc={`*For display purpose only*\nWhich URL should be seen as the "UI hostname"`}
                                    img={UIHostname}
                                    value={formFields?.ui_host}
                                    onChanges={(e) => handleChange('ui_host', e.target.value)}
                                    placeholder={localStorage.getItem(LOCAL_STORAGE_UI_HOST) === undefined ? localStorage.getItem(LOCAL_STORAGE_UI_HOST) : ''}
                                />
                                <InputRow
                                    title="REST GATEWAY HOSTNAME"
                                    desc={`*For display purpose only*\nWhich URL should be seen as the "REST Gateway hostname"`}
                                    img={RestHostname}
                                    value={formFields?.rest_gw_host}
                                    onChanges={(e) => handleChange('rest_gw_host', e.target.value)}
                                    placeholder={localStorage.getItem(LOCAL_STORAGE_REST_GW_HOST) === undefined ? localStorage.getItem(LOCAL_STORAGE_REST_GW_HOST) : ''}
                                />
                            </>
                        )}
                    </div>
                    <div className="configuration-footer">
                        <div className="btn-container">
                            <Button
                                className="modal-btn"
                                width="100px"
                                height="34px"
                                placeholder="Discard"
                                colorType="gray-dark"
                                radiusType="circle"
                                backgroundColorType="none"
                                border="gray"
                                boxShadowsType="gray"
                                fontSize="12px"
                                fontWeight="600"
                                aria-haspopup="true"
                                disabled={!isChanged}
                                onClick={() => discardChanges()}
                            />
                            <Button
                                className="modal-btn"
                                width="100px"
                                height="34px"
                                placeholder="Apply"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontWeight="600"
                                aria-haspopup="true"
                                disabled={!isChanged}
                                onClick={() => updateConfiguration()}
                            />
                        </div>
                    </div>
                </>
            )}
        </div>
    );
}

export default ClusterConfiguration;
