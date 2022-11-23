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

import React, { useState } from 'react';
import confImg1 from '../../../../../assets/images/confImg1.svg';
import figmaIcon from '../../../../../assets/images/figmaIcon.svg';
import { INTEGRATION_LIST } from '../../../../../const/integrationList';
import { FiberManualRecord } from '@material-ui/icons';
import { diffDate } from '../../../../../services/valueConvertor';
import Input from '../../../../../components/Input';
import { Form } from 'antd';
import { URL } from '../../../../../config';
import Switcher from '../../../../../components/switcher';
import Button from '../../../../../components/button';

const urlSplit = URL.split('/', 3)[0];

const SlackIntegration = ({ close }) => {
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        ui_url: `${urlSplit[0]}//${urlSplit[2]}`,
        auth_token: '',
        channel_id: '',
        poison_message_alert: true,
        schema_validation_fail_alert: true
    });
    const slackConfiguration = INTEGRATION_LIST[0];

    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    const handleSubmit = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            try {
                // const data = await httpRequest('POST', ApiEndpoints.CREATE_NEW_SCHEMA, { ...values, message_struct_name: messageName || messageStructName });
            } catch (err) {}
        }
    };

    return (
        <slack-integration is="3xd" className="integration-modal-container">
            {slackConfiguration?.insideBanner}
            {slackConfiguration.header}
            {slackConfiguration.integrateDesc}
            {/* <div className="header">
                {slackConfiguration?.icon}
                <div className="details">
                    <p>{slackConfiguration?.name}</p>
                    <>
                        <span>by {slackConfiguration.by}</span>
                        <FiberManualRecord />
                        <span>Last update: {diffDate(slackConfiguration.date)} </span>
                    </>
                </div>
            </div> */}
            {/* <div className="integrate-description">
                <p>Description</p>
                <span className="content">
                    Receive alerts and notifications directly to your chosen slack channel for faster response and better real-time observability. Read More
                </span>
            </div> */}
            <Form name="form" form={creationForm} autoComplete="off" className="integration-form">
                <div className="api-details">
                    <p className="title">API details</p>
                    <div className="api-key">
                        <p>API KEY</p>
                        <span className="desc">
                            There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour
                        </span>
                        <Form.Item
                            name="auth_token"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert auth token.'
                                }
                            ]}
                        >
                            <Input
                                placeholder="Insert auth token"
                                type="text"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="purple"
                                borderColorType="none"
                                height="40px"
                                fontSize="12px"
                                onBlur={(e) => updateFormState('auth_token', e.target.value)}
                                onChange={(e) => updateFormState('auth_token', e.target.value)}
                                value={creationForm.auth_token}
                            />
                        </Form.Item>
                    </div>
                    <div className="channel-id">
                        <p>Channel ID</p>
                        <span className="desc">There are many variations of passages of Lorem Ipsum available.</span>
                        <Form.Item
                            name="channel_id"
                            rules={[
                                {
                                    required: true,
                                    message: 'Please insert channel id'
                                }
                            ]}
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
                                onBlur={(e) => updateFormState('channel_id', e.target.value)}
                                onChange={(e) => updateFormState('channel_id', e.target.value)}
                                value={formFields.channel_id}
                            />
                        </Form.Item>
                    </div>
                    <div className="notification-option">
                        <p>Notify me when:</p>
                        <span className="desc">Memphis will send only the selected triggers</span>
                        <>
                            <div className="option-wrapper">
                                <div className="option-name">
                                    <img src={confImg1} />
                                    <p>POISION_MESSAGE</p>
                                </div>
                                <Switcher
                                    onChange={() => updateFormState('poison_message_alert', !formFields.poison_message_alert)}
                                    checked={formFields.poison_message_alert}
                                />
                            </div>
                            <div className="option-wrapper">
                                <div className="option-name">
                                    <img src={confImg1} />
                                    <p>SCHEMA_VALIDATION_FAIL </p>
                                </div>
                                <Switcher
                                    onChange={() => updateFormState('schema_validation_fail_alert', !formFields.schema_validation_fail_alert)}
                                    checked={formFields.schema_validation_fail_alert}
                                />
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
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={close}
                        />
                        <Button
                            width="80%"
                            height="45px"
                            placeholder="Integrate"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            // isLoading={loadingSubmit}
                            // disabled={validateError}
                            onClick={handleSubmit}
                        />
                    </div>
                </Form.Item>
            </Form>
        </slack-integration>
    );
};

export default SlackIntegration;
