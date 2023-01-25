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

import React, { useState, useContext } from 'react';
import { Form } from 'antd';

import { INTEGRATION_LIST, REGIONS_OPTIONS } from '../../../../../const/integrationList';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Button from '../../../../../components/button';
import { Context } from '../../../../../hooks/store';
import Input from '../../../../../components/Input';
import SelectComponent from '../../../../../components/select';

const S3Integration = ({ close, value }) => {
    const isValue = value && Object.keys(value)?.length !== 0;
    const s3Configuration = INTEGRATION_LIST[1];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [formFields, setFormFields] = useState({
        name: 's3',
        keys: {
            secret_key: value?.keys?.secret_key || '',
            access_key: value?.keys?.access_key || '',
            region: value?.keys?.region || REGIONS_OPTIONS[0].value,
            bucket_name: value?.keys?.bucket_name || ''
        }
    });
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);

    const updateKeysState = (field, value) => {
        let updatedValue = { ...formFields.keys };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    };

    const handleSubmit = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            setLoadingSubmit(true);
            if (isValue) {
                if (creationForm.isFieldTouched('secret_key')) {
                    updateIntegration();
                } else {
                    updateIntegration(false);
                }
            } else {
                createIntegration();
            }
        }
    };

    const updateIntegration = async (withToken = true) => {
        let newFormFields = { ...formFields };
        if (!withToken) {
            let updatedKeys = { ...formFields.keys };
            updatedKeys['secret_key'] = '';
            newFormFields = { ...newFormFields, keys: updatedKeys };
        }
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATIONL, { ...newFormFields });
            dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
            setTimeout(() => {
                setLoadingSubmit(false);
            }, 1000);
            close(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const createIntegration = async () => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, { ...formFields });
            dispatch({ type: 'ADD_INTEGRATION', payload: data });
            setTimeout(() => {
                setLoadingSubmit(false);
            }, 1000);
            close(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };
    const disconnect = async () => {
        setLoadingDisconnect(true);
        try {
            await httpRequest('DELETE', ApiEndpoints.DISCONNECT_INTEGRATION, {
                name: formFields.name
            });
            dispatch({ type: 'REMOVE_INTEGRATION', payload: formFields.name });
            setTimeout(() => {
                setLoadingDisconnect(false);
            }, 1000);
            close({});
        } catch (err) {
            setLoadingDisconnect(false);
        }
    };

    return (
        <dynamic-integration is="3xd" className="integration-modal-container">
            {s3Configuration?.insideBanner}
            <div className="integrate-header">
                {s3Configuration.header}
                <div className={!isValue ? 'action-buttons flex-end' : 'action-buttons'}>
                    {isValue && (
                        <Button
                            width="100px"
                            height="35px"
                            placeholder="Disconnect"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="red"
                            border="none"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            isLoading={loadingDisconnect}
                            disabled={process.env.REACT_APP_SANDBOX_ENV}
                            onClick={() => disconnect()}
                        />
                    )}
                    <Button
                        width="140px"
                        height="35px"
                        placeholder="Integration guide"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        border="none"
                        fontSize="12px"
                        fontFamily="InterSemiBold"
                        onClick={() => window.open('https://docs.memphis.dev/memphis/dashboard-gui/integrations/storage/amazon-s3', '_blank')}
                    />
                </div>
            </div>
            {s3Configuration.integrateDesc}
            <Form name="form" form={creationForm} autoComplete="off" className="integration-form">
                <div className="api-details">
                    <p className="title">Integration details</p>
                    <div className="api-key">
                        <p>Secret access key</p>
                        <span className="desc">
                            When you use AWS programmatically, you provide your AWS access keys so that AWS can verify your identity in programmatic calls. Access keys
                            can be either temporary (short-term) credentials or long-term credentials, such as for an IAM user or the AWS account root user. <br />
                            <b>Memphis encrypts all stored information using Triple DES algorithm</b>
                        </span>
                        <Form.Item
                            name="secret_key"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert auth token.'
                                }
                            ]}
                            initialValue={formFields?.keys?.secret_key}
                        >
                            <Input
                                placeholder="****+crc"
                                type="text"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="purple"
                                borderColorType="none"
                                height="40px"
                                fontSize="12px"
                                onBlur={(e) => updateKeysState('secret_key', e.target.value)}
                                onChange={(e) => updateKeysState('secret_key', e.target.value)}
                                value={formFields?.keys?.secret_key}
                            />
                        </Form.Item>
                    </div>
                    <div className="input-field">
                        <p>Access Key ID</p>
                        <Form.Item
                            name="access_key"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert access key id'
                                }
                            ]}
                            initialValue={formFields?.keys?.access_key}
                        >
                            <Input
                                placeholder="AKIOOJB9EKLP69O4RTHR"
                                type="text"
                                fontSize="12px"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                height="40px"
                                onBlur={(e) => updateKeysState('access_key', e.target.value)}
                                onChange={(e) => updateKeysState('access_key', e.target.value)}
                                value={formFields.keys?.access_key}
                            />
                        </Form.Item>
                    </div>
                    <div className="select-field">
                        <p>Region</p>
                        <Form.Item name="region" initialValue={formFields?.keys?.region || REGIONS_OPTIONS[0].name}>
                            <SelectComponent
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                popupClassName="select-options"
                                options={REGIONS_OPTIONS}
                                value={formFields?.keys?.region || REGIONS_OPTIONS[0].name}
                                onChange={(e) => updateKeysState('region', e.match(/\[(.*?)\]/)[1])}
                            />
                        </Form.Item>
                    </div>
                    <div className="input-field">
                        <p>Bucket name</p>
                        <Form.Item
                            name="bucket_name"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert bucket name'
                                }
                            ]}
                            initialValue={formFields?.keys?.bucket_name}
                        >
                            <Input
                                placeholder="Insert your bucket name"
                                type="text"
                                fontSize="12px"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                height="40px"
                                onBlur={(e) => updateKeysState('bucket_name', e.target.value)}
                                onChange={(e) => updateKeysState('bucket_name', e.target.value)}
                                value={formFields.keys?.bucket_name}
                            />
                        </Form.Item>
                    </div>
                </div>
                <Form.Item className="button-container">
                    <div className="button-wrapper">
                        <Button
                            width="80%"
                            height="45px"
                            placeholder="Close"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="14px"
                            fontFamily="InterSemiBold"
                            onClick={() => close(value)}
                        />
                        <Button
                            width="80%"
                            height="45px"
                            placeholder={isValue ? 'Update' : 'Connect'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="14px"
                            fontFamily="InterSemiBold"
                            isLoading={loadingSubmit}
                            disabled={process.env.REACT_APP_SANDBOX_ENV || (isValue && !creationForm.isFieldsTouched())}
                            onClick={handleSubmit}
                        />
                    </div>
                </Form.Item>
            </Form>
        </dynamic-integration>
    );
};

export default S3Integration;
