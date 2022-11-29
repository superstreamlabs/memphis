// Credit for The NATS.IO Authors
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

import React, { useState, useContext } from 'react';
import { Form } from 'antd';

import poisionAlertIcon from '../../../../../assets/images/poisionAlertIcon.svg';
import disconAlertIcon from '../../../../../assets/images/disconAlertIcon.svg';
import schemaAlertIcon from '../../../../../assets/images/schemaAlertIcon.svg';
import { INTEGRATION_LIST } from '../../../../../const/integrationList';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Switcher from '../../../../../components/switcher';
import Button from '../../../../../components/button';
import { Context } from '../../../../../hooks/store';
import Input from '../../../../../components/Input';
import { URL } from '../../../../../config';

const urlSplit = URL.split('/', 3);

const SlackIntegration = ({ close, value }) => {
    const isValue = value && Object.keys(value)?.length !== 0;
    const slackConfiguration = INTEGRATION_LIST[0];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);

    const [formFields, setFormFields] = useState({
        name: 'slack',
        ui_url: `${urlSplit[0]}//${urlSplit[2]}`,
        keys: {
            auth_token: value?.keys?.auth_token || '',
            channel_id: value?.keys?.channel_id || ''
        },
        properties: {
            poison_message_alert: value?.properties?.poison_message_alert ? true : false,
            schema_validation_fail_alert: value?.properties?.schema_validation_fail_alert ? true : false,
            disconnection_events_alert: value?.properties?.disconnection_events_alert ? true : false
        }
    });
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);

    const updateKeysState = (field, value) => {
        let updatedValue = { ...formFields.keys };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    };
    const updatePropertiesState = (field, value) => {
        let updatedValue = { ...formFields.properties };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, properties: updatedValue }));
    };

    const handleSubmit = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            setLoadingSubmit(true);
            if (isValue) {
                updateIntegration();
            } else {
                createIntegration();
            }
        }
    };

    const updateIntegration = async () => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATIONL, { ...formFields });
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
            await httpRequest('DELETE', `${ApiEndpoints.DELETE_INTEGRATION}?=`, {
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
        <slack-integration is="3xd" className="integration-modal-container">
            {slackConfiguration?.insideBanner}
            <div className="integrate-header">
                {slackConfiguration.header}
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
                        onClick={() => window.open('https://docs.memphis.dev/memphis/dashboard-ui/integrations/notifications/slack', '_blank')}
                    />
                </div>
            </div>
            {slackConfiguration.integrateDesc}
            <Form name="form" form={creationForm} autoComplete="off" className="integration-form">
                <div className="api-details">
                    <p className="title">API details</p>
                    <div className="api-key">
                        <p>API KEY</p>
                        <span className="desc">Copy and paste your slack 'Bot User OAuth Token' here</span>
                        <Form.Item
                            name="auth_token"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert auth token.'
                                }
                            ]}
                            initialValue={formFields?.keys?.auth_token}
                        >
                            <Input
                                placeholder="xoxb-****"
                                type="text"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="purple"
                                borderColorType="none"
                                height="40px"
                                fontSize="12px"
                                onBlur={(e) => updateKeysState('auth_token', e.target.value)}
                                onChange={(e) => updateKeysState('auth_token', e.target.value)}
                                value={formFields?.keys?.auth_token}
                            />
                        </Form.Item>
                    </div>
                    <div className="channel-id">
                        <p>Channel ID</p>
                        <span className="desc">To which slack channel should Memphis push notifications?</span>
                        <Form.Item
                            name="channel_id"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert channel id'
                                }
                            ]}
                            initialValue={formFields?.keys?.channel_id}
                        >
                            <Input
                                placeholder="#C0P4ISJH06K"
                                type="text"
                                fontSize="12px"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                height="40px"
                                onBlur={(e) => updateKeysState('channel_id', e.target.value)}
                                onChange={(e) => updateKeysState('channel_id', e.target.value)}
                                value={formFields.keys?.channel_id}
                            />
                        </Form.Item>
                    </div>
                    <div className="notification-option">
                        <p>Notify me when:</p>
                        <span className="desc">Memphis will send only the selected triggers</span>
                        <>
                            <div className="option-wrapper">
                                <div className="option-name">
                                    <img src={poisionAlertIcon} />
                                    <div className="name-des">
                                        <p>New poison message</p>
                                        <span>
                                            Messages that cause a consumer group to repeatedly require a delivery (possibly due to a consumer failure) such that the
                                            message is never processed completely and acknowledged
                                        </span>
                                    </div>
                                </div>
                                <Form.Item name="poison_message_alert">
                                    <Switcher
                                        onChange={() => updatePropertiesState('poison_message_alert', !formFields.properties.poison_message_alert)}
                                        checked={formFields.properties?.poison_message_alert}
                                    />
                                </Form.Item>
                            </div>
                            <div className="option-wrapper">
                                <div className="option-name">
                                    <img src={schemaAlertIcon} />
                                    <div className="name-des">
                                        <p>Schema validation failure</p>
                                        <span>Triggered once a client fails in schema validation</span>
                                    </div>
                                </div>
                                <Form.Item name="schema_validation_fail_alert">
                                    <Switcher
                                        onChange={() => updatePropertiesState('schema_validation_fail_alert', !formFields.properties.schema_validation_fail_alert)}
                                        checked={formFields.properties?.schema_validation_fail_alert}
                                    />
                                </Form.Item>
                            </div>
                            <div className="option-wrapper">
                                <div className="option-name">
                                    <img src={disconAlertIcon} />
                                    <div className="name-des">
                                        <p>Disconnected clients</p>
                                        <span>Triggered once a producer/consumer get disconnected</span>
                                    </div>
                                </div>
                                <Form.Item name="schema_validation_fail_alert">
                                    <Switcher
                                        onChange={() => updatePropertiesState('disconnection_events_alert', !formFields.properties.disconnection_events_alert)}
                                        checked={formFields.properties?.disconnection_events_alert}
                                    />
                                </Form.Item>
                            </div>
                        </>
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
        </slack-integration>
    );
};

export default SlackIntegration;
