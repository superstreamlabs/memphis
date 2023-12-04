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

import { FiMinusCircle, FiPlus } from 'react-icons/fi';
import React, { useContext, useEffect, useState } from 'react';
import Editor, { loader } from '@monaco-editor/react';
import { useHistory } from 'react-router-dom';
import { Divider, Form, Space } from 'antd';
import * as monaco from 'monaco-editor';

import { StationStoreContext } from '../../domain/stationOverview';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { convertArrayToObject, generateJSONWithMaxLength, isCloud } from '../../services/valueConvertor';
import { ReactComponent as RefreshIcon } from '../../assets/images/refresh.svg';
import InputNumberComponent from '../InputNumber';
import { httpRequest } from '../../services/http';
import TitleComponent from '../titleComponent';
import SelectComponent from '../select';
import Switcher from '../switcher';
import Button from '../button';
import Input from '../Input';
import Copy from '../copy';
import pathDomains from '../../router';
import CloudOnly from '../cloudOnly';

loader.init();
loader.config({ monaco });
const partitons = ['all'];

const ProduceMessages = ({ stationName, cancel, produceMessagesRef, setLoading }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const history = useHistory();
    const [messageExample, setMessageExample] = useState(generateJSONWithMaxLength(isCloud() ? 120 : 55));
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({});

    useEffect(() => {
        produceMessagesRef.current = onFinish;
    }, [messageExample]);

    const generateMessage = () => {
        setMessageExample(generateJSONWithMaxLength(isCloud() ? 120 : 55));
    };

    const handleEditorChange = (newValue) => {
        if (isCloud() || (!isCloud() && newValue.length <= 100)) {
            setMessageExample(newValue);
        }
    };

    const generateEditor = () => {
        return (
            <>
                <Editor
                    options={{
                        minimap: { enabled: false },
                        scrollbar: { verticalScrollbarSize: 0, horizontalScrollbarSize: 0 },
                        scrollBeyondLastLine: false,
                        roundedSelection: false,
                        formatOnPaste: true,
                        formatOnType: true,
                        fontSize: '12px',
                        fontFamily: 'Inter',
                        lineNumbers: 'off'
                    }}
                    className="editor-message"
                    language={'json'}
                    height="calc(100% - 25px)"
                    width="calc(100% - 25px)"
                    value={messageExample}
                    onChange={handleEditorChange}
                />
                <Copy data={messageExample} />
            </>
        );
    };

    const onFinish = async () => {
        const formFields = await creationForm.validateFields();
        if (formFields.message_headers) formFields.message_headers = convertArrayToObject(formFields.message_headers);
        if (formFields.partition_number === 'all') formFields.partition_number = -1;
        const bodyRequest = { ...formFields, message_payload: messageExample, station_name: stationName };

        try {
            setLoading(true);
            await httpRequest('POST', ApiEndpoints.PRODUCE, bodyRequest);
            getStationDetails();
        } catch (error) {
            setLoading(false);
        }
    };

    const sortData = (data) => {
        data.audit_logs?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.messages?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.active_producers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.active_consumers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.destroyed_consumers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.destroyed_producers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.killed_consumers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        data.killed_producers?.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        return data;
    };

    const getStationDetails = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_STATION_DATA}?station_name=${stationName}&partition_number=-1`);
            await sortData(data);
            stationDispatch({ type: 'SET_SOCKET_DATA', payload: data });
            stationDispatch({ type: 'SET_SCHEMA_TYPE', payload: data.schema.schema_type });
            setLoading(false);
            cancel();
        } catch (error) {
            setLoading(false);
            if (error.status === 404) {
                history.push(pathDomains.stations);
            }
        }
    };

    return (
        <div className="produce-modal-wrapper">
            <div className="produce-message">
                <div className="generate-wrapper">
                    <p className="field-title">JSON-based value</p>
                    <div className="generate-action" onClick={() => generateMessage()}>
                        <RefreshIcon width={14} />
                        <span>Generate example</span>
                    </div>
                </div>
                <div className="message-example">
                    <div className="code-content">{generateEditor()}</div>
                </div>
                {!isCloud() && <p>{100 - messageExample.length} characters are left</p>}
            </div>
            <Divider className="seperator" />
            <Form name="form" form={creationForm} onFinish={onFinish} onAbort={cancel} autoComplete="on" className="produce-form">
                <div className="up-form">
                    <p className="field-title">Headers</p>
                    <Form.List name="message_headers">
                        {(fields, { add, remove }) => (
                            <>
                                <div className="headers-wrapper">
                                    {fields.map(({ key, name, ...restField }) => (
                                        <Space
                                            key={key}
                                            style={{
                                                display: 'flex',
                                                marginBottom: 8
                                            }}
                                            align="baseline"
                                        >
                                            <Form.Item
                                                {...restField}
                                                name={[name, 'key']}
                                                rules={[
                                                    {
                                                        required: true,
                                                        message: 'Missing key'
                                                    }
                                                ]}
                                            >
                                                <Input
                                                    placeholder="key"
                                                    type="text"
                                                    radiusType="semi-round"
                                                    colorType="black"
                                                    backgroundColorType="none"
                                                    borderColorType="gray"
                                                    height="40px"
                                                    onBlur={(e) => creationForm.setFieldsValue({[name]: e.target.value})}
                                                    onChange={(e) => creationForm.setFieldsValue({[name]: e.target.value})}
                                                    value={formFields.key}
                                                />
                                            </Form.Item>
                                            <Form.Item
                                                {...restField}
                                                name={[name, 'value']}
                                                rules={[
                                                    {
                                                        required: true,
                                                        message: 'Missing value'
                                                    }
                                                ]}
                                            >
                                                <Input
                                                    placeholder="value"
                                                    type="text"
                                                    radiusType="semi-round"
                                                    colorType="black"
                                                    backgroundColorType="none"
                                                    borderColorType="gray"
                                                    height="40px"
                                                    onBlur={(e) => creationForm.setFieldsValue({[name]: e.target.value})}
                                                    onChange={(e) => creationForm.setFieldsValue({[name]: e.target.value})}
                                                    value={formFields.header}
                                                />
                                            </Form.Item>
                                            <FiMinusCircle className="remove-icon" onClick={() => remove(name)} />
                                        </Space>
                                    ))}
                                </div>
                                <Form.Item>
                                    <div className="add-field" onClick={() => add()}>
                                        <FiPlus />
                                        <span>New header</span>
                                    </div>
                                </Form.Item>
                            </>
                        )}
                    </Form.List>
                    <Divider className="seperator" />
                    <div className="by-pass-switcher">
                        <TitleComponent
                            headerTitle="Bypass schema enforcement"
                            cloudOnly={isCloud() ? false : true}
                            typeTitle="sub-header"
                            headerDescription="Check this box to avoid schema validation"
                        />
                        <Form.Item className="form-input" name="bypass_schema" initialValue={isCloud() ? false : true}>
                            <Switcher disabled={!isCloud()} onChange={(e) => creationForm.setFieldsValue({'bypass_schema': e})} checked={isCloud() ? formFields.bypass_schema : true} />
                        </Form.Item>
                    </div>
                    <Divider className="seperator" />
                    <div className="partition-records-section">
                        <Form.Item className="form-input" name="partition_number" initialValue={partitons[0]}>
                            <div className="header-flex">
                                <p className="field-title">Partition</p>
                                {!isCloud() && <CloudOnly />}
                            </div>
                            <SelectComponent
                                value={formFields.partition_number || partitons[0]}
                                colorType="navy"
                                backgroundColorType={isCloud() ? 'none' : 'disabled'}
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="45px"
                                width="100%"
                                options={partitons}
                                onChange={(e) => creationForm.setFieldsValue({'partition_number': e.target.value})}
                                popupClassName="select-options"
                                disabled={!isCloud()}
                            />
                        </Form.Item>

                        <Form.Item className="form-input" name="amount" initialValue={1}>
                            <div className="header-flex">
                                <p className="field-title">Number of records</p>
                                {!isCloud() && <CloudOnly />}
                            </div>
                            <InputNumberComponent
                                min={1}
                                max={isCloud() ? 1000 : 1}
                                onChange={(e) =>creationForm.setFieldsValue({'amount': e})}
                                value={formFields.amount}
                                placeholder={formFields.amount || 1}
                                disabled={!isCloud()}
                                width="100%"
                            />
                        </Form.Item>
                    </div>
                </div>
            </Form>
        </div>
    );
};

export default ProduceMessages;
