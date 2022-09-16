// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';
import React, { useState, useEffect, useContext } from 'react';
import { Form, InputNumber } from 'antd';
import TitleComponent from '../../../../components/titleComponent';
import RadioButton from '../../../../components/radioButton';
import Input from '../../../../components/Input';
import { convertDateToSeconds } from '../../../../services/valueConvertor';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import { GetStartedStoreContext } from '..';

const retanionOptions = [
    {
        id: 1,
        value: 'message_age_sec',
        label: 'Time'
    },
    {
        id: 2,
        value: 'bytes',
        label: 'Size'
    },
    {
        id: 3,
        value: 'messages',
        label: 'Messages'
    }
];

const storageOptions = [
    {
        id: 1,
        value: 'file',
        label: 'File'
    },
    {
        id: 2,
        value: 'memory',
        label: 'Memory'
    }
];

const CreateStationForm = (props) => {
    const { createStationFormRef } = props;
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [creationForm] = Form.useForm();
    const [allowEdit, setAllowEdit] = useState(true);

    useEffect(() => {
        checkState();
        createStationFormRef.current = onFinish;
    }, []);

    const checkState = () => {
        if (getStartedState?.formFieldsCreateStation?.name) {
            setAllowEdit(false);
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
        } else {
            setAllowEdit(true);
        }
    };

    const onFinish = async () => {
        if (getStartedState?.completedSteps > 0) {
            getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
        } else {
            let retentionValue = 0;
            try {
                const values = await creationForm.validateFields();
                getStartedDispatch({ type: 'IS_LOADING', payload: true });
                if (values.retention_type === 'message_age_sec') {
                    retentionValue = convertDateToSeconds(values.days, values.hours, values.minutes, values.seconds);
                } else if (values.retention_type === 'bytes') {
                    retentionValue = Number(values.retentionSizeValue);
                } else {
                    retentionValue = Number(values.retentionMessagesValue);
                }
                updateFormState('retention_value', retentionValue);
                const bodyRequest = {
                    name: values.name,
                    retention_type: values.retention_type,
                    retention_value: retentionValue,
                    storage_type: values.storage_type,
                    replicas: values.replicas
                };
                createStation(bodyRequest);
            } catch (error) {}
        }
    };
    const createStation = async (bodyRequest) => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.CREATE_STATION, bodyRequest);
            if (data) {
                getStartedDispatch({ type: 'SET_STATION', payload: data.name });
                getStartedDispatch({ type: 'SET_COMPLETED_STEPS', payload: getStartedState?.currentStep });
                getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
            }
        } catch (error) {
        } finally {
            getStartedDispatch({ type: 'IS_LOADING', payload: false });
        }
    };

    const updateFormState = (field, value) => {
        getStartedDispatch({ type: 'SET_FORM_FIELDS_CREATE_STATION', payload: { field: field, value: value } });
    };

    return (
        <Form name="form" form={creationForm} autoComplete="off" className="create-station-form-getstarted">
            <div id="e2e-getstarted-step1">
                <TitleComponent
                    headerTitle="Enter station name"
                    typeTitle="sub-header"
                    headerDescription="RabbitMQ has queues, Kafka has topics, and Memphis has stations"
                    required={true}
                ></TitleComponent>
                <Form.Item
                    name="name"
                    rules={[
                        {
                            required: true,
                            message: 'Please input station name!'
                        }
                    ]}
                    style={{ height: '50px' }}
                    initialValue={getStartedState?.formFieldsCreateStation.name}
                >
                    <Input
                        placeholder="Type station name"
                        type="text"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="none"
                        borderColorType="gray"
                        width="450px"
                        height="40px"
                        onBlur={(e) => updateFormState('name', e.target.value)}
                        onChange={(e) => updateFormState('name', e.target.value)}
                        value={getStartedState?.formFieldsCreateStation.name}
                        disabled={!allowEdit}
                    />
                </Form.Item>
            </div>
            <div>
                <TitleComponent
                    headerTitle="Retention type"
                    typeTitle="sub-header"
                    headerDescription="By which criteria messages will be expelled from the station"
                ></TitleComponent>
                <Form.Item name="retention_type" initialValue={getStartedState?.formFieldsCreateStation.retention_type}>
                    <RadioButton
                        className="radio-button"
                        options={retanionOptions}
                        radioValue={getStartedState?.formFieldsCreateStation.retention_type}
                        optionType="button"
                        style={{ marginRight: '20px', content: '' }}
                        onChange={(e) => updateFormState('retention_type', e.target.value)}
                        disabled={!allowEdit}
                    />
                </Form.Item>

                {getStartedState?.formFieldsCreateStation.retention_type === 'message_age_sec' && (
                    <div className="time-value">
                        <div className="days-section">
                            <Form.Item name="days" initialValue={getStartedState?.formFieldsCreateStation?.days}>
                                <InputNumber
                                    bordered={false}
                                    min={0}
                                    max={100}
                                    keyboard={true}
                                    onChange={(e) => updateFormState('days', e)}
                                    value={getStartedState?.formFieldsCreateStation?.days}
                                    placeholder={getStartedState?.formFieldsCreateStation?.days || 7}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>days</p>
                        </div>
                        <p className="separator">:</p>
                        <div className="hours-section">
                            <Form.Item name="hours" initialValue={getStartedState?.formFieldsCreateStation?.hours}>
                                <InputNumber
                                    bordered={false}
                                    min={0}
                                    max={24}
                                    keyboard={true}
                                    onChange={(e) => updateFormState('hours', e)}
                                    value={getStartedState?.formFieldsCreateStation?.hours}
                                    placeholder={getStartedState?.formFieldsCreateStation?.hours || 0}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>hours</p>
                        </div>
                        <p className="separator">:</p>
                        <div className="minutes-section">
                            <Form.Item name="minutes" initialValue={getStartedState?.formFieldsCreateStation?.minutes}>
                                <InputNumber
                                    bordered={false}
                                    min={0}
                                    max={60}
                                    keyboard={true}
                                    onChange={(e) => updateFormState('minutes', e)}
                                    value={getStartedState?.formFieldsCreateStation?.minutes}
                                    placeholder={getStartedState?.formFieldsCreateStation?.minutes || 0}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>minutes</p>
                        </div>
                        <p className="separator">:</p>
                        <div className="seconds-section">
                            <Form.Item name="seconds" initialValue={getStartedState?.formFieldsCreateStation?.seconds}>
                                <InputNumber
                                    bordered={false}
                                    min={0}
                                    max={60}
                                    keyboard={true}
                                    onChange={(e) => updateFormState('seconds', e)}
                                    placeholder={getStartedState?.formFieldsCreateStation?.seconds || 0}
                                    value={getStartedState?.formFieldsCreateStation?.seconds}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>seconds</p>
                        </div>
                    </div>
                )}
                {getStartedState?.formFieldsCreateStation.retention_type === 'bytes' && (
                    <div className="retention-type">
                        <Form.Item name="retentionSizeValue" initialValue={getStartedState?.formFieldsCreateStation?.retentionSizeValue}>
                            <Input
                                placeholder="Type"
                                type="number"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                width="90px"
                                height="38px"
                                onBlur={(e) => updateFormState('retentionSizeValue', e.target.value)}
                                onChange={(e) => updateFormState('retentionSizeValue', e.target.value)}
                                value={getStartedState?.formFieldsCreateStation?.retentionSizeValue}
                                disabled={!allowEdit}
                            />
                        </Form.Item>
                        <p>bytes</p>
                    </div>
                )}
                {getStartedState?.formFieldsCreateStation.retention_type === 'messages' && (
                    <div className="retention-type">
                        <Form.Item name="retentionMessagesValue" initialValue={getStartedState?.formFieldsCreateStation?.retentionMessagesValue}>
                            <Input
                                placeholder="Type"
                                type="number"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                width="90px"
                                height="38px"
                                onBlur={(e) => updateFormState('retentionMessagesValue', e)}
                                onChange={(e) => updateFormState('retentionMessagesValue', e)}
                                value={getStartedState?.formFieldsCreateStation?.retentionMessagesValue}
                                disabled={!allowEdit}
                            />
                        </Form.Item>
                        <p>messages</p>
                    </div>
                )}
            </div>
            <div className="storage-replicas-container">
                <div>
                    <TitleComponent
                        headerTitle="Storage type"
                        typeTitle="sub-header"
                        headerDescription="Type of message persistence"
                        style={{ description: { width: '18vw' } }}
                    ></TitleComponent>
                    <Form.Item name="storage_type" initialValue={getStartedState?.formFieldsCreateStation?.storage_type}>
                        <RadioButton
                            options={storageOptions}
                            radioValue={getStartedState?.formFieldsCreateStation?.storage_type}
                            optionType="button"
                            onChange={(e) => updateFormState('storage_type', e.target.value)}
                            disabled={!allowEdit}
                        />
                    </Form.Item>
                </div>
                <div>
                    <TitleComponent
                        headerTitle="Replicas"
                        typeTitle="sub-header"
                        headerDescription="Amount of mirrors per message"
                        style={{ description: { width: '16vw' } }}
                    ></TitleComponent>
                    <div>
                        <Form.Item name="replicas" initialValue={getStartedState?.formFieldsCreateStation?.replicas}>
                            <InputNumber
                                bordered={false}
                                min={1}
                                max={getStartedState?.actualPods && getStartedState?.actualPods <= 5 ? getStartedState?.actualPods : 5}
                                keyboard={true}
                                value={getStartedState?.formFieldsCreateStation?.replicas}
                                onChange={(e) => updateFormState('replicas', e)}
                                disabled={!allowEdit}
                            />
                        </Form.Item>
                    </div>
                </div>
            </div>
        </Form>
    );
};
export default CreateStationForm;
