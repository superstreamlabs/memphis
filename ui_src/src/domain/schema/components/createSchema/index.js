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

import React, { useState } from 'react';

import BackIcon from '../../../../assets/images/backIcon.svg';
import Input from '../../../../components/Input';
import { Form } from 'antd';
import TagsList from '../../../../components/tagsList';
import RadioButton from '../../../../components/radioButton';
import Editor from '@monaco-editor/react';

const schemaTypes = [
    {
        id: 1,
        value: 0,
        label: 'Protobuf',
        description: (
            <span>
                Contrary to popular belief, Lorem Ipsum is not simply random text. Latin literature from 45 BC{' '}
                <a href="https://docs.memphis.dev/memphis-new/getting-started/1-installation" target="_blank">
                    Learn More
                </a>
            </span>
        )
    },
    {
        id: 2,
        value: 1,
        label: 'Avro',
        description: (
            <span>
                Contrary to popular belief, Lorem Ipsum is not simply random text. Latin literature from 45 BC{' '}
                <a href="https://docs.memphis.dev/memphis-new/getting-started/1-installation" target="_blank">
                    Learn More
                </a>
            </span>
        ),
        disabled: true
    },
    {
        id: 3,
        value: 2,
        label: 'Json',
        description: (
            <span>
                Contrary to popular belief, Lorem Ipsum is not simply random text. Latin literature from 45 BC{' '}
                <a href="https://docs.memphis.dev/memphis-new/getting-started/1-installation" target="_blank">
                    Learn More
                </a>
            </span>
        ),
        disabled: false
    }
];

const SchemaEditorExample = {
    0: {
        language: 'proto',
        value: `syntax = "proto3";
        ​
        message SchemaMemphis {
            string field1 = 1;
            string  field2 = 2;
            int32  field3 = 3;
        }`
    },
    2: {
        language: 'json',
        value: `{
            "type": "record",
            "namespace": "com.example",
            "name": "test-schema",
            "fields": [
               { "name": "username", "type": "string", "default": "-2" },
               { "name": "age", "type": "int", "default": "none" },
               { "name": "phone", "type": "int", "default": "NONE" },
               { "name": "country", "type": "string", "default": "NONE" }
            ]
        }`
    }
};

function CreateSchema({ createNew }) {
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        name: '',
        type: 0,
        tags: [],
        schema: ''
    });
    const [updated, setUpdated] = useState(false);

    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    const onFinish = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        }
    };

    return (
        <div className="create-schema-wrapper">
            <div className="header">
                <div className="flex-title">
                    <img src={BackIcon} onClick={() => createNew()} />
                    <p>Create Schema</p>
                </div>
                <span>
                    Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature
                    <a href="https://docs.memphis.dev/memphis-new/getting-started/1-installation" target="_blank">
                        Learn More
                    </a>
                </span>
            </div>
            <Form name="form" form={creationForm} autoComplete="off" onFinish={onFinish} className="create-schema-form">
                <div className="left-side">
                    <Form.Item
                        name="name"
                        rules={[
                            {
                                required: true,
                                message: 'Please input schema name!'
                            }
                        ]}
                    >
                        <div className="schema-field name">
                            <p className="field-title">Schema Name</p>
                            <Input
                                placeholder="Type schema name"
                                type="text"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                height="40px"
                                width="200px"
                                onBlur={(e) => updateFormState('name', e.target.value)}
                                onChange={(e) => updateFormState('name', e.target.value)}
                                value={formFields.name}
                            />
                        </div>
                    </Form.Item>
                    <Form.Item name="tags">
                        <div className="schema-field tags">
                            <p className="field-title">Tags</p>
                            <TagsList addNew={true} />
                        </div>
                    </Form.Item>
                    <Form.Item name="type">
                        <div className="schema-field type">
                            <p className="field-title">Schema Type</p>
                            <RadioButton
                                vertical={true}
                                options={schemaTypes}
                                radioWrapper="schema-type"
                                radioValue={formFields.type}
                                onChange={(e) => updateFormState('type', e.target.value)}
                            />
                        </div>
                    </Form.Item>
                </div>
                <div className="right-side">
                    <Form.Item name="schema" style={{ height: '100%' }}>
                        <div className="schema-field schema">
                            <p className="field-title">Schema Defination</p>
                            <div className="editor">
                                <Editor
                                    options={{
                                        minimap: { enabled: false },
                                        scrollbar: { verticalScrollbarSize: 0 },
                                        scrollBeyondLastLine: false,
                                        roundedSelection: false,
                                        formatOnPaste: true,
                                        formatOnType: true
                                    }}
                                    language={SchemaEditorExample[formFields?.type]?.language}
                                    value={SchemaEditorExample[formFields?.type]?.value}
                                    onChange={() => setUpdated(true)}
                                />
                            </div>
                        </div>
                    </Form.Item>
                </div>
            </Form>
        </div>
    );
}

export default CreateSchema;
