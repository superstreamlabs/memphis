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

import { INTEGRATION_LIST } from '../../../../../const/integrationList';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Button from '../../../../../components/button';
import { Context } from '../../../../../hooks/store';
import Input from '../../../../../components/Input';
import SelectComponent from '../../../../../components/select';

import { URL } from '../../../../../config';

const urlSplit = URL.split('/', 3);

const regionOptions = [
    {
        name: 'US East (Ohio) [us-east-2]',
        value: 'us-east-2'
    },
    {
        name: 'US East (N. Virginia) [us-east-1]',
        value: 'us-east-1'
    },
    {
        name: 'US West (N. California) [us-west-1]',
        value: 'us-west-1'
    },
    {
        name: 'US West (Oregon) [us-west-2]',
        value: 'us-west-2'
    },
    {
        name: 'Africa (Cape Town) [af-south-1]',
        value: 'af-south-1'
    },
    {
        name: 'Asia Pacific (Hong Kong) [ap-east-1]',
        value: 'ap-east-1'
    },
    {
        name: 'Asia Pacific (Hyderabad) [ap-south-2]',
        value: 'ap-south-2'
    },
    {
        name: 'Asia Pacific (Jakarta) [ap-southeast-3]',
        value: 'ap-southeast-3'
    },
    {
        name: 'Asia Pacific (Mumbai) [ap-south-1]',
        value: 'ap-south-1'
    },
    {
        name: 'Asia Pacific (Osaka) [ap-northeast-3]',
        value: 'ap-northeast-3'
    },
    {
        name: 'Asia Pacific (Seoul) [ap-northeast-2]',
        value: 'ap-northeast-2'
    },
    {
        name: 'Asia Pacific (Singapore) [ap-southeast-1]',
        value: 'ap-southeast-1'
    },
    {
        name: 'Asia Pacific (Sydney) [ap-southeast-2]',
        value: 'ap-southeast-2'
    },
    {
        name: 'Asia Pacific (Tokyo) [ap-northeast-1]',
        value: 'ap-northeast-1'
    },
    {
        name: 'Canada (Central) [ca-central-1]',
        value: 'ca-central-1'
    },
    {
        name: 'Europe (Frankfurt) [eu-central-1]',
        value: 'eu-central-1'
    },
    {
        name: 'Europe (Ireland) [eu-west-1]',
        value: 'eu-west-1'
    },
    {
        name: 'Europe (London) [eu-west-2]',
        value: 'eu-west-2'
    },
    {
        name: 'Europe (Milan) [eu-south-1]',
        value: 'eu-south-1'
    },
    {
        name: 'Europe (Paris) [eu-west-3]',
        value: 'eu-west-3'
    },
    {
        name: 'Europe (Spain) [eu-south-2]',
        value: 'eu-south-2'
    },
    {
        name: 'Europe (Stockholm) [eu-north-1]',
        value: 'eu-north-1'
    },
    {
        name: 'Europe (Zurich) [eu-central-2]',
        value: 'eu-central-2'
    },
    {
        name: 'Middle East (Bahrain) [me-south-1]',
        value: 'me-south-1'
    },
    {
        name: 'Middle East (UAE) [me-central-1]',
        value: 'me-central-1'
    },
    {
        name: 'South America (São Paulo) [sa-east-1]',
        value: 'sa-east-1'
    }
];

const S3Integration = ({ close, value }) => {
    const isValue = value && Object.keys(value)?.length !== 0;
    const s3Configuration = INTEGRATION_LIST[1];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [formFields, setFormFields] = useState({
        name: 's3',
        ui_url: `${urlSplit[0]}//${urlSplit[2]}`,
        keys: {
            secret_access_key: value?.keys?.secret_access_key || '',
            access_key_id: value?.keys?.access_key_id || ''
        },
        region: value?.keys?.region || regionOptions[0].value,
        bucket_name: value?.keys?.bucket_name || ''
    });
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);

    const updateKeysState = (field, value) => {
        let updatedValue = { ...formFields.keys };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    };
    const updateState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    const handleSubmit = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            setLoadingSubmit(true);
            if (isValue) {
                if (values.secret_access_key === 'xoxb-****') {
                    updateIntegration(false);
                } else {
                    updateIntegration();
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
            updatedKeys['secret_access_key'] = '';
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
                            name="secret_access_key"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert auth token.'
                                }
                            ]}
                            initialValue={formFields?.keys?.secret_access_key}
                        >
                            <Input
                                placeholder="***************3FUIjt"
                                type="text"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="purple"
                                borderColorType="none"
                                height="40px"
                                fontSize="12px"
                                onBlur={(e) => updateKeysState('secret_access_key', e.target.value)}
                                onChange={(e) => updateKeysState('secret_access_key', e.target.value)}
                                value={formFields?.keys?.secret_access_key}
                            />
                        </Form.Item>
                    </div>
                    <div className="input-field">
                        <p>Access Key ID</p>
                        <Form.Item
                            name="access_key_id"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert access key id'
                                }
                            ]}
                            initialValue={formFields?.keys?.access_key_id}
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
                                onBlur={(e) => updateKeysState('access_key_id', e.target.value)}
                                onChange={(e) => updateKeysState('access_key_id', e.target.value)}
                                value={formFields.keys?.access_key_id}
                            />
                        </Form.Item>
                    </div>
                    <div className="select-field">
                        <p>Region</p>
                        <Form.Item name="region" initialValue={formFields?.keys?.region || regionOptions[0].name}>
                            <SelectComponent
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                popupClassName="select-options"
                                options={regionOptions}
                                value={formFields?.keys?.region || regionOptions[0].name}
                                onChange={(e) => updateState('region', e)}
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
                                onBlur={(e) => updateState('bucket_name', e.target.value)}
                                onChange={(e) => updateState('bucket_name', e.target.value)}
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
