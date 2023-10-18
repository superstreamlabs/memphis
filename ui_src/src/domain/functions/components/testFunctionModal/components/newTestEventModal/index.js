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
import { useState, useEffect } from 'react';
import Editor from '@monaco-editor/react';
import Input from '../../../../../../components/Input';
import Button from '../../../../../../components/button';
import { FaPlay } from 'react-icons/fa';
import { ApiEndpoints } from '../../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../../services/http';

import TestResult from '../testResult';
const TEMPLATE_CARDS = [
    {
        name: 'Generate synthetic data',
        description: 'In case you prefer to generate a random test event',
        isAvailable: true
    },
    {
        name: 'An event from a station',
        description: 'Create a test event from a synthetic data',
        isAvailable: false
    },
    {
        name: 'An event from a source',
        description: 'Create a test event from a data source',
        isAvailable: false
    }
];

const DEFAULT_TEXT = `"type": "record",
"namespace": "com.example",
"name": "test-schema",
"fields": [
       { "name": "Master message", "type": "string", "default": "NONE" },
       { "name": "age", "type": "int", "default": "-1" },
       { "name": "phone", "type": "string", "default": "NONE" },
       { "name": "country", "type": "string", "default": "NONE" }
]`;
const NewTestEventModal = ({ onCancel, updateTestEvents, editEvent }) => {
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [content, setContent] = useState(DEFAULT_TEXT);
    const [isLoading, setIsLoading] = useState(false);
    const [selectedTemplate, setSelectedTemplate] = useState(0);
    const [isTested, setIsTested] = useState(false);

    const getEvent = async () => {
        try {
            setIsLoading(true);
            const res = await httpRequest('GET', encodeURI(`${ApiEndpoints.GET_TEST_EVENT}?test_event_name=${editEvent}`));
            setName(res.name);
            setContent(res.content);
            setDescription(res.description);
        } catch (err) {
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        if (editEvent) {
            getEvent();
        }
    }, [editEvent]);

    const handleSave = async () => {
        try {
            setIsLoading(true);

            if (editEvent) {
                const bodyRequest = {
                    name: editEvent,
                    new_name: name,
                    description,
                    content
                };
                await httpRequest('PUT', ApiEndpoints.UPDATE_TEST_EVENT, bodyRequest);
            } else {
                const bodyRequest = {
                    name,
                    description,
                    content
                };
                await httpRequest('POST', ApiEndpoints.CREATE_TEST_EVENT, bodyRequest);
            }
            await updateTestEvents();
            onCancel();
        } catch (e) {
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="newTestEventModal-container">
            <div className="overflow-wrapper">
                <div className="form-wrapper">
                    <div className="name form">
                        <p className="title">Template name</p>
                        <Input
                            placeholder={'name'}
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            colorType={'black'}
                            backgroundColorType={'none'}
                            height={'38px'}
                            borderColorType="gray"
                            radiusType="semi-round"
                            type="text"
                        />
                    </div>
                    <div className="description form">
                        <p className="title">Description</p>
                        <Input
                            placeholder={'description'}
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            colorType={'black'}
                            backgroundColorType={'none'}
                            height={'38px'}
                            borderColorType="gray"
                            radiusType="semi-round"
                            type="text"
                        />
                    </div>
                </div>
                <div className="template-wrapper">
                    <p className="title">Select Template</p>
                    <div className="templates">
                        {TEMPLATE_CARDS.map((item, index) => (
                            <div className={`template-card ${index > 0 ? 'disabled' : 'selected'}`} key={index} tabIndex={0} onClick={() => setSelectedTemplate(index)}>
                                {!item.isAvailable && <div className="badge">Coming Soon</div>}
                                <p className="title">{item.name}</p>
                                <p className="description">{item.description}</p>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="divider" />
                <div className="action-wrapper">
                    <p className="title">Generate Fake Data</p>

                    <div className="actions">
                        <Button
                            placeholder={
                                <div className="button-content">
                                    <span>Import data</span>
                                    <div className="badge">Coming soon</div>
                                </div>
                            }
                            backgroundColorType={'white'}
                            border={'gray-light'}
                            colorType={'black'}
                            fontSize={'12px'}
                            fontFamily={'InterSemiBold'}
                            radiusType={'circle'}
                            height={'34px'}
                            width={'119px'}
                            disabled={true}
                        />
                        <Button
                            placeholder={
                                <div className="button-content">
                                    <div className="badge">Coming soon</div>

                                    <span>Regenerate</span>
                                </div>
                            }
                            backgroundColorType={'white'}
                            border={'gray-light'}
                            colorType={'black'}
                            fontSize={'12px'}
                            fontFamily={'InterSemiBold'}
                            radiusType={'circle'}
                            height={'34px'}
                            width={'186px'}
                            disabled={true}
                        />
                        <div className="line" />
                        <Button
                            placeholder={
                                <div className="button-content">
                                    <FaPlay />
                                    <span>Test</span>
                                </div>
                            }
                            backgroundColorType={'orange'}
                            border={'none'}
                            colorType={'black'}
                            fontSize={'12px'}
                            fontFamily={'InterSemiBold'}
                            radiusType={'circle'}
                            height={'34px'}
                            width={'99px'}
                            onClick={() => setIsTested(true)}
                        />
                    </div>
                </div>
                <div className="text-area-wrapper">
                    <div className={`text-wrapper ${isTested ? 'width-50' : ''}  `}>
                        <Editor
                            options={{
                                minimap: { enabled: false },
                                scrollbar: { verticalScrollbarSize: 3 },
                                scrollBeyondLastLine: false,
                                roundedSelection: false,
                                formatOnPaste: true,
                                formatOnType: true,
                                fontSize: '14px',
                                fontFamily: 'Inter',
                                lineNumbers: 'off',
                                glyphMargin: false,
                                folding: false,
                                lineDecorationsWidth: 0,
                                lineNumbersMinChars: 0,
                                automaticLayout: true
                            }}
                            language={'json'}
                            height="100%"
                            defaultValue={DEFAULT_TEXT}
                            value={content}
                            key={isTested ? 'tested' : 'not-tested'}
                            onChange={(value) => {
                                setContent(value);
                            }}
                        />
                    </div>
                    {isTested && <TestResult name={name} />}
                </div>
            </div>

            <div className="footer">
                <Button
                    placeholder={'Cancel'}
                    backgroundColorType={'white'}
                    border={'gray-light'}
                    colorType={'black'}
                    fontSize={'14px'}
                    fontFamily={'InterSemibold'}
                    radiusType={'circle'}
                    width={'168px'}
                    height={'34px'}
                    onClick={onCancel}
                />

                <Button
                    placeholder={'Save'}
                    backgroundColorType={'purple'}
                    border={'none'}
                    colorType={'white'}
                    fontSize={'14px'}
                    fontFamily="InterSemibold"
                    radiusType={'circle'}
                    width={'168px'}
                    height={'34px'}
                    isLoading={isLoading}
                    onClick={handleSave}
                />
            </div>
        </div>
    );
};

export default NewTestEventModal;
