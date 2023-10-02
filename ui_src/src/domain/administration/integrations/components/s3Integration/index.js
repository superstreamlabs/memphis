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

import React, { useState, useContext, useEffect } from 'react';
import { Form, message } from 'antd';

import { INTEGRATION_LIST, getTabList } from '../../../../../const/integrationList';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Button from '../../../../../components/button';
import { Context } from '../../../../../hooks/store';
import CustomTabs from '../../../../../components/Tabs';
import Input from '../../../../../components/Input';
import Checkbox from '../../../../../components/checkBox';
import Loader from '../../../../../components/loader';
import { showMessages } from '../../../../../services/genericServices';
import IntegrationDetails from '../integrationItem/integrationDetails';
import IntegrationLogs from '../integrationItem/integrationLogs';

const S3Integration = ({ close, value }) => {
    const isValue = value && Object.keys(value)?.length !== 0;
    const s3Configuration = INTEGRATION_LIST['S3'];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [formFields, setFormFields] = useState({
        name: 's3',
        keys: {
            secret_key: value?.keys?.secret_key || '',
            access_key: value?.keys?.access_key || '',
            region: value?.keys?.region || '',
            bucket_name: value?.keys?.bucket_name || '',
            url: value?.keys?.url || '',
            s3_path_style: value?.keys?.s3_path_style || ''
        }
    });
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [tabValue, setTabValue] = useState('Configuration');
    const tabs = getTabList('Slack');

    useEffect(() => {
        const images = [];
        images.push(INTEGRATION_LIST['S3'].banner.props.src);
        images.push(INTEGRATION_LIST['S3'].insideBanner.props.src);
        const promises = [];

        images.forEach((imageUrl) => {
            const image = new Image();
            promises.push(
                new Promise((resolve) => {
                    image.onload = resolve;
                })
            );
            image.src = imageUrl;
        });

        Promise.all(promises).then(() => {
            setImagesLoaded(true);
        });
    }, []);

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

    const closeModal = (data, disconnect = false) => {
        setTimeout(() => {
            disconnect ? setLoadingDisconnect(false) : setLoadingSubmit(false);
        }, 1000);
        close(data);
        showMessages('success', disconnect ? 'The integration was successfully disconnected' : 'The integration connected successfully');
    };

    const updateIntegration = async (withToken = true) => {
        let newFormFields = { ...formFields };
        if (!withToken) {
            let updatedKeys = { ...formFields.keys };
            updatedKeys['secret_key'] = '';
            newFormFields = { ...newFormFields, keys: updatedKeys };
        }
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATION, { ...newFormFields });
            dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
            closeModal(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const createIntegration = async () => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, { ...formFields });
            dispatch({ type: 'ADD_INTEGRATION', payload: data });
            closeModal(data);
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
            closeModal({}, true);
        } catch (err) {
            setLoadingDisconnect(false);
        }
    };

    return (
        <dynamic-integration is="3xd" className="integration-modal-container">
            {!imagesLoaded && (
                <div className="loader-integration-box">
                    <Loader />
                </div>
            )}
            {imagesLoaded && (
                <>
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

                    <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs} />
                    <Form name="form" form={creationForm} autoComplete="off" className="integration-form">
                        {tabValue === 'Details' && <IntegrationDetails integrateDesc={s3Configuration.integrateDesc} />}
                        {tabValue === 'Logs' && <IntegrationLogs integrationName={'s3'} />}
                        {tabValue === 'Configuration' && (
                            <div className="integration-body">
                                <IntegrationDetails integrateDesc={s3Configuration.integrateDesc} />
                                <div className="api-details">
                                    <p className="title">Integration details</p>
                                    <div className="api-key">
                                        <p>Secret access key</p>
                                        <span className="desc">
                                            When you use S3 compatible storage programmatically, you provide your access keys so that the provider can verify your
                                            identity in programmatic calls. Access keys can be either temporary (short-term) credentials or long-term credentials, such as
                                            for an IAM user, provider provided keys or credentials. <br />
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
                                    <div className="flex-fields">
                                        <div className="input-field">
                                            <p>Region</p>
                                            <Form.Item
                                                name="region"
                                                initialValue={formFields?.keys?.region}
                                                rules={[
                                                    {
                                                        required: true,
                                                        message: 'Please insert region'
                                                    }
                                                ]}
                                            >
                                                <Input
                                                    type="text"
                                                    placeholder="us-east-1"
                                                    fontSize="12px"
                                                    colorType="black"
                                                    backgroundColorType="none"
                                                    borderColorType="gray"
                                                    radiusType="semi-round"
                                                    height="40px"
                                                    value={formFields?.keys?.region}
                                                    onBlur={(e) => updateKeysState('region', e.target.value)}
                                                    onChange={(e) => updateKeysState('region', e.target.value)}
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
                                    <div className="flex-fields">
                                        <div className="input-field">
                                            <p>Endpoint URL (optional)</p>
                                            <Form.Item
                                                name="url"
                                                rules={[
                                                    {
                                                        required: false
                                                    }
                                                ]}
                                                initialValue={formFields?.keys?.url}
                                            >
                                                <Input
                                                    placeholder="Insert custom S3 API endpoint url (Optional; leave empty for AWS)"
                                                    type="text"
                                                    fontSize="12px"
                                                    radiusType="semi-round"
                                                    colorType="black"
                                                    backgroundColorType="none"
                                                    borderColorType="gray"
                                                    height="40px"
                                                    onBlur={(e) => updateKeysState('url', e.target.value)}
                                                    onChange={(e) => updateKeysState('url', e.target.value)}
                                                    value={formFields.keys?.url}
                                                />
                                            </Form.Item>
                                        </div>
                                        <div className="input-field">
                                            <p>Use Path Style</p>
                                            <span className="desc">The URL path contains the s3 bucket name.</span>
                                            <Form.Item
                                                name="s3_path_style"
                                                rules={[
                                                    {
                                                        required: false
                                                    }
                                                ]}
                                                initialValue={formFields?.keys?.s3_path_style}
                                            >
                                                <>
                                                    <Checkbox
                                                        defaultChecked={false}
                                                        checkName="s3_path_style"
                                                        checked={formFields.keys?.s3_path_style === '1' ? true : false}
                                                        onChange={(e) => updateKeysState('s3_path_style', e.target.checked ? '1' : '0')}
                                                    />{' '}
                                                    Enable
                                                </>
                                            </Form.Item>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
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
                                    disabled={isValue && !creationForm.isFieldsTouched()}
                                    onClick={handleSubmit}
                                />
                            </div>
                        </Form.Item>
                    </Form>
                </>
            )}
        </dynamic-integration>
    );
};

export default S3Integration;
