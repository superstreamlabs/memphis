// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useState } from 'react';

import { compareObjects } from '../../../services/valueConvertor';
import ConfImg1 from '../../../assets/images/confImg1.svg';
import ConfImg2 from '../../../assets/images/confImg2.svg';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import Button from '../../../components/button';
import SliderRow from './components/sliderRow';
import { message } from 'antd';

function ClusterConfiguration() {
    const [formFields, setFormFields] = useState({});
    const [oldValues, setOldValues] = useState({});
    const [isChanged, setIsChanged] = useState(false);

    useEffect(() => {
        getConfigurationValue();
    }, []);

    const getConfigurationValue = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_CONFIGURATION);
            setOldValues(data);
            setFormFields(data);
        } catch (err) {
            return;
        }
    };

    const updateConfiguration = async () => {
        try {
            const data = await httpRequest('PUT', ApiEndpoints.EDIT_CLUSTER_CONFIGURATION, { ...formFields });
            setIsChanged(false);
            setOldValues(data);
            message.success({
                key: 'memphisSuccessMessage',
                content: 'Successfully updated',
                duration: 5,
                style: { cursor: 'pointer' },
                onClick: () => message.destroy('memphisSuccessMessage')
            });
        } catch (err) {
            return;
        }
    };

    const handleChange = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setIsChanged(!compareObjects(updatedValue, oldValues));
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };
    const discardChanges = () => {
        setIsChanged(false);
        setFormFields((formFields) => ({ ...formFields, ...oldValues }));
    };

    return (
        <div className="configuration-container">
            <div className="header">
                <p className="main-header">Cluster configuration</p>
                <p className="sub-header">In this section, you can tune 'Memphis' internal configuration to suit your requirements</p>
            </div>
            <div className="configuration-body">
                <SliderRow
                    title="DEAD LETTER MESSAGES RETENTION IN HOURS"
                    desc="Amount of hours to retain dead letter messages in a DLS"
                    value={formFields?.pm_retention}
                    img={ConfImg2}
                    min={1}
                    max={30}
                    unit={'h'}
                    onChanges={(e) => handleChange('pm_retention', e)}
                />
                <SliderRow
                    title="LOGS RETENTION IN DAYS"
                    desc="Amount of days to retain system logs"
                    img={ConfImg1}
                    value={formFields?.logs_retention}
                    min={1}
                    max={100}
                    unit={'d'}
                    onChanges={(e) => handleChange('logs_retention', e)}
                />
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
        </div>
    );
}

export default ClusterConfiguration;
