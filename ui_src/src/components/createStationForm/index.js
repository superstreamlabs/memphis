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
import React, { useState, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import pathDomains from '../../router';

import { Form } from 'antd';
import TitleComponent from '../titleComponent';
import RadioButton from '../radioButton';
import Switcher from '../switcher';
import Input from '../Input';
import { convertDateToSeconds } from '../../services/valueConvertor';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';

import InputNumberComponent from '../InputNumber';
import SelectComponent from '../select';
import SelectSchema from '../selectSchema';

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

const CreateStationForm = ({ createStationFormRef, getStartedStateRef, finishUpdate, updateFormState, getStarted, setLoading }) => {
    const history = useHistory();
    const [creationForm] = Form.useForm();
    const [allowEdit, setAllowEdit] = useState(true);
    const [actualPods, setActualPods] = useState(null);
    const [retentionType, setRetentionType] = useState(retanionOptions[0].value);
    const [storageType, setStorageType] = useState(storageOptions[0].value);
    const [schemas, setSchemas] = useState([]);
    const [useSchema, setUseSchema] = useState(true);

    useEffect(() => {
        getOverviewData();
        getAllSchemas();
        if (getStarted && getStartedStateRef?.completedSteps > 0) setAllowEdit(false);
        if (getStarted && getStartedStateRef?.formFieldsCreateStation?.retention_type) setRetentionType(getStartedStateRef.formFieldsCreateStation.retention_type);
        createStationFormRef.current = onFinish;
    }, []);

    const getRetentionValue = (formFields) => {
        switch (formFields.retention_type) {
            case 'message_age_sec':
                return convertDateToSeconds(formFields.days, formFields.hours, formFields.minutes, formFields.seconds);
            case 'messages':
                return Number(formFields.retentionMessagesValue);
            case 'bytes':
                return Number(formFields.retentionValue);
        }
    };

    const onFinish = async () => {
        const formFields = await creationForm.validateFields();
        const retentionValue = getRetentionValue(formFields);
        const bodyRequest = {
            name: formFields.name,
            retention_type: formFields.retention_type,
            retention_value: retentionValue,
            storage_type: formFields.storage_type,
            replicas: formFields.replicas,
            schema_name: formFields.schemaValue
        };
        createStation(bodyRequest);
    };

    const getOverviewData = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_MAIN_OVERVIEW_DATA);
            let indexOfBrokerComponent = data?.system_components.findIndex((item) => item.component.includes('broker'));
            indexOfBrokerComponent = indexOfBrokerComponent !== -1 ? indexOfBrokerComponent : 1;
            data?.system_components[indexOfBrokerComponent]?.actual_pods && setActualPods(data?.system_components[indexOfBrokerComponent]?.actual_pods);
        } catch (error) {}
    };

    const getAllSchemas = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_SCHEMAS);
            setSchemas(data);
        } catch (error) {}
    };

    const createStation = async (bodyRequest) => {
        try {
            getStarted && setLoading(true);
            const data = await httpRequest('POST', ApiEndpoints.CREATE_STATION, bodyRequest);
            if (data) {
                if (!getStartedStateRef) history.push(`${pathDomains.stations}/${data.name}`);
                else {
                    finishUpdate(data);
                }
            }
        } catch (error) {
            console.log(error);
        } finally {
            getStarted && setLoading(false);
        }
    };

    return (
        <Form name="form" form={creationForm} autoComplete="off" className="create-station-form-getstarted">
            <div id="e2e-getstarted-step1" className="station-name-section">
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
                    initialValue={getStartedStateRef?.formFieldsCreateStation?.name}
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
                        onBlur={(e) => getStarted && updateFormState('name', e.target.value)}
                        onChange={(e) => getStarted && updateFormState('name', e.target.value)}
                        value={getStartedStateRef?.formFieldsCreateStation?.name}
                        disabled={!allowEdit}
                    />
                </Form.Item>
            </div>
            <div className="retention-type-section">
                <TitleComponent
                    headerTitle="Retention type"
                    typeTitle="sub-header"
                    headerDescription="By which criteria messages will be expelled from the station"
                ></TitleComponent>
                <Form.Item name="retention_type" initialValue={getStarted ? getStartedStateRef?.formFieldsCreateStation?.retention_type : 'message_age_sec'}>
                    <RadioButton
                        className="radio-button"
                        options={retanionOptions}
                        radioValue={getStarted ? getStartedStateRef?.formFieldsCreateStation?.retention_type : retentionType}
                        optionType="button"
                        fontFamily="InterSemiBold"
                        style={{ marginRight: '20px', content: '' }}
                        onChange={(e) => {
                            setRetentionType(e.target.value);
                            if (getStarted) updateFormState('retention_type', e.target.value);
                        }}
                        disabled={!allowEdit}
                    />
                </Form.Item>
                {retentionType === 'message_age_sec' && (
                    <div className="time-value">
                        <div className="days-section">
                            <Form.Item name="days" initialValue={getStartedStateRef?.formFieldsCreateStation?.days || 7}>
                                <InputNumberComponent
                                    min={0}
                                    max={100}
                                    onChange={(e) => getStarted && updateFormState('days', e)}
                                    value={getStartedStateRef?.formFieldsCreateStation?.days}
                                    placeholder={getStartedStateRef?.formFieldsCreateStation?.days || 7}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>days</p>
                        </div>
                        <p className="separator">:</p>
                        <div className="hours-section">
                            <Form.Item name="hours" initialValue={getStartedStateRef?.formFieldsCreateStation?.hours || 0}>
                                <InputNumberComponent
                                    min={0}
                                    max={24}
                                    onChange={(e) => getStarted && updateFormState('hours', e)}
                                    value={getStartedStateRef?.formFieldsCreateStation?.hours}
                                    placeholder={getStartedStateRef?.formFieldsCreateStation?.hours || 0}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>hours</p>
                        </div>
                        <p className="separator">:</p>
                        <div className="minutes-section">
                            <Form.Item name="minutes" initialValue={getStartedStateRef?.formFieldsCreateStation?.minutes || 0}>
                                <InputNumberComponent
                                    min={0}
                                    max={60}
                                    onChange={(e) => getStarted && updateFormState('minutes', e)}
                                    value={getStartedStateRef?.formFieldsCreateStation?.minutes}
                                    placeholder={getStartedStateRef?.formFieldsCreateStation?.minutes || 0}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>minutes</p>
                        </div>
                        <p className="separator">:</p>
                        <div className="seconds-section">
                            <Form.Item name="seconds" initialValue={getStartedStateRef?.formFieldsCreateStation?.seconds || 0}>
                                <InputNumberComponent
                                    min={0}
                                    max={60}
                                    onChange={(e) => getStarted && updateFormState('seconds', e)}
                                    placeholder={getStartedStateRef?.formFieldsCreateStation?.seconds || 0}
                                    value={getStartedStateRef?.formFieldsCreateStation?.seconds}
                                    disabled={!allowEdit}
                                />
                            </Form.Item>
                            <p>seconds</p>
                        </div>
                    </div>
                )}
                {retentionType === 'bytes' && (
                    <div className="retention-type">
                        <Form.Item name="retentionValue" initialValue={getStartedStateRef?.formFieldsCreateStation?.retentionSizeValue || 1000}>
                            <Input
                                placeholder="Type"
                                type="number"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                width="90px"
                                height="38px"
                                onBlur={(e) => getStarted && updateFormState('retentionSizeValue', e.target.value)}
                                onChange={(e) => getStarted && updateFormState('retentionSizeValue', e.target.value)}
                                value={getStartedStateRef?.formFieldsCreateStation?.retentionSizeValue}
                                disabled={!allowEdit}
                            />
                        </Form.Item>
                        <p>bytes</p>
                    </div>
                )}
                {retentionType === 'messages' && (
                    <div className="retention-type">
                        <Form.Item name="retentionMessagesValue" initialValue={getStartedStateRef?.formFieldsCreateStation?.retentionMessagesValue || 10}>
                            <Input
                                placeholder="Type"
                                type="number"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                width="90px"
                                height="38px"
                                onBlur={(e) => getStarted && updateFormState('retentionMessagesValue', e.target.value)}
                                onChange={(e) => getStarted && updateFormState('retentionMessagesValue', e.target.value)}
                                value={getStartedStateRef?.formFieldsCreateStation?.retentionMessagesValue}
                                disabled={!allowEdit}
                            />
                        </Form.Item>
                        <p>messages</p>
                    </div>
                )}
            </div>
            <div className="storage-replicas-container">
                <div className="storage-container">
                    <TitleComponent
                        headerTitle="Storage type"
                        typeTitle="sub-header"
                        headerDescription="Type of message persistence"
                        style={{ description: { width: '240px' } }}
                    ></TitleComponent>
                    <Form.Item name="storage_type" initialValue={getStarted ? getStartedStateRef?.formFieldsCreateStation?.storage_type : 'file'}>
                        <RadioButton
                            options={storageOptions}
                            fontFamily="InterSemiBold"
                            radioValue={getStarted ? getStartedStateRef?.formFieldsCreateStation?.storage_type : storageType}
                            optionType="button"
                            onChange={(e) => {
                                setStorageType(e.target.value);
                                getStarted && updateFormState('storage_type', e.target.value);
                            }}
                            disabled={!allowEdit}
                        />
                    </Form.Item>
                </div>
                <div className="replicas-container">
                    <TitleComponent
                        headerTitle="Replicas"
                        typeTitle="sub-header"
                        headerDescription="Amount of mirrors per message"
                        style={{ description: { width: '240px' } }}
                    ></TitleComponent>
                    <div>
                        <Form.Item name="replicas" initialValue={getStarted ? getStartedStateRef?.formFieldsCreateStation?.replicas : 1}>
                            <InputNumberComponent
                                min={1}
                                max={actualPods && actualPods <= 5 ? actualPods : 5}
                                value={getStarted ? getStartedStateRef?.formFieldsCreateStation?.replicas : 1}
                                onChange={(e) => getStarted && updateFormState('replicas', e)}
                                disabled={!allowEdit}
                            />
                        </Form.Item>
                    </div>
                </div>
            </div>
            {!getStarted && schemas.length > 0 && (
                <div className="schema-type">
                    <Form.Item name="schemaValue">
                        <div className="toggle-add-schema">
                            <TitleComponent headerTitle="Use schema" typeTitle="sub-header"></TitleComponent>
                            <Switcher onChange={() => setUseSchema(!useSchema)} checked={useSchema} />
                        </div>
                        {useSchema && (
                            <SelectSchema
                                height="40px"
                                value={creationForm.schemaValue}
                                colorType="navy"
                                backgroundColorType="none"
                                radiusType="semi-round"
                                width="450px"
                                placeholder="Select schema"
                                options={schemas}
                                onChange={(e) => creationForm.setFieldsValue({ schemaValue: e })}
                                boxShadowsType="gray"
                                popupClassName="select-options"
                                disabled={!allowEdit}
                            />
                        )}
                    </Form.Item>
                </div>
            )}
        </Form>
    );
};
export default CreateStationForm;
