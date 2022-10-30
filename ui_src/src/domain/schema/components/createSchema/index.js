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

import { CheckCircleOutlineRounded, ErrorOutlineRounded } from '@material-ui/icons';
import React, { useEffect, useState } from 'react';
import Schema from 'protocol-buffers-schema';
import Editor from '@monaco-editor/react';
import { Form } from 'antd';

import schemaTypeIcon from '../../../../assets/images/schemaTypeIcon.svg';
import errorModal from '../../../../assets/images/errorModal.svg';
import BackIcon from '../../../../assets/images/backIcon.svg';
import tagsIcon from '../../../../assets/images/tagsIcon.svg';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import RadioButton from '../../../../components/radioButton';
import SelectComponent from '../../../../components/select';
import { httpRequest } from '../../../../services/http';
import Button from '../../../../components/button';
import Input from '../../../../components/Input';
import Modal from '../../../../components/modal';
import TagsList from '../../../../components/tagList';

const schemaTypes = [
    {
        id: 1,
        value: 'Protobuf',
        label: 'Protobuf',
        description: (
            <span>
                The modern. Protocol buffers are Google's language-neutral, platform-neutral, extensible mechanism for serializing structured data – think XML, but
                smaller, faster, and simpler.
            </span>
        )
    },
    {
        id: 2,
        value: 'avro',
        label: 'Avro',
        description: (
            <span>
                The popular. Apache Avro™ is the leading serialization format for record data, and first choice for streaming data pipelines. It offers excellent schema
                evolution.
            </span>
        ),
        disabled: true
    },
    {
        id: 3,
        value: 'json',
        label: 'Json',
        description: <span>The simplest. JSON Schema is a vocabulary that allows you to annotate and validate JSON documents.</span>,
        disabled: true
    }
];

const SchemaEditorExample = {
    Protobuf: {
        language: 'proto',
        value: `syntax = "proto3";
        message Test {
            string field1 = 1;
            string  field2 = 2;
            int32  field3 = 3;
        }`
    },
    avro: {
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

function CreateSchema({ goBack }) {
    const [creationForm] = Form.useForm();

    const [tagsToDisplay, setTagsToDisplay] = useState([]);

    const [formFields, setFormFields] = useState({
        name: '',
        type: 'Protobuf',
        tags: [],
        schema_content: '',
        message_struct_name: 'test'
    });
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [validateLoading, setValidateLoading] = useState(false);
    const [validateError, setValidateError] = useState('');
    const [validateSuccess, setValidateSuccess] = useState(false);
    const [messageStructName, setMessageStructName] = useState('Test');
    const [messagesStructNameList, setMessagesStructNameList] = useState([]);

    const [modalOpen, setModalOpen] = useState(false);

    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
        if (field === 'schema_content') {
            setValidateSuccess('');
            setValidateError('');
        }
    };

    useEffect(() => {
        updateFormState('schema_content', SchemaEditorExample[formFields?.type]?.value);
    }, []);

    const handleSubmit = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            if (values.type === 'Protobuf') {
                let parser = Schema.parse(values.schema_content).messages;
                if (parser.length === 1) {
                    setMessageStructName(parser[0].name);
                    handleCreateNewSchema();
                } else {
                    setMessageStructName(parser[0].name);
                    setMessagesStructNameList(parser);
                    setModalOpen(true);
                }
            } else {
                handleCreateNewSchema();
            }
        }
    };

    const handleCreateNewSchema = async () => {
        try {
            const values = await creationForm.validateFields();
            debugger;
            setLoadingSubmit(true);
            const data = await httpRequest('POST', ApiEndpoints.CREATE_NEW_SCHEMA, { ...values, message_struct_name: messageStructName });
            if (data) {
                goBack();
            }
        } catch (err) {
            if (err.status === 555) {
                setValidateSuccess('');
                setValidateError(err.data.message);
            }
        }
        setLoadingSubmit(false);
    };

    const handleValidateSchema = async () => {
        setValidateLoading(true);
        try {
            const data = await httpRequest('POST', ApiEndpoints.VALIDATE_SCHEMA, {
                schema_type: formFields?.type,
                schema_content: formFields.schema_content
            });
            if (data.is_valid) {
                setValidateError('');
                setTimeout(() => {
                    setValidateSuccess('Schema is valid');
                    setValidateLoading(false);
                }, 1000);
            }
        } catch (error) {
            if (error.status === 555) {
                setValidateSuccess('');
                setValidateError(error.data.message);
            }
            setValidateLoading(false);
        }
    };

    const checkContent = (_, value) => {
        if (value.length > 0) {
            try {
                Schema.parse(value);
                setValidateSuccess('');
                setValidateError('');
            } catch (error) {
                setValidateSuccess('');
                setValidateError(error.message);
            }
            return Promise.resolve();
        } else {
            setValidateSuccess('');
            setValidateError('Schema content cannot be empty');
            return Promise.reject(new Error());
        }
    };

    const removeTag = (tagName) => {
        let updatedTags = tagsToDisplay.filter((tag) => tag.name !== tagName);
        setTagsToDisplay(updatedTags);
        updateFormState('tags', updatedTags);
    };

    return (
        <div className="create-schema-wrapper">
            <Form name="form" form={creationForm} autoComplete="off" className="create-schema-form">
                <div className="left-side">
                    <div className="header">
                        <div className="flex-title">
                            <img src={BackIcon} onClick={() => goBack()} alt="backIcon" />
                            <p>Create schema</p>
                        </div>
                        <span>
                            Creating a schema will enable you to enforce standardization upon produced data and increase data quality.
                            <a href="https://docs.memphis.dev/memphis-new/getting-started/1-installation" target="_blank">
                                Learn more
                            </a>
                        </span>
                    </div>
                    <Form.Item
                        name="name"
                        rules={[
                            {
                                required: true,
                                message: 'Please input schema name!'
                            }
                        ]}
                        style={{ height: '114px' }}
                    >
                        <div className="schema-field name">
                            <p className="field-title">Schema name</p>
                            <Input
                                placeholder="Type schema name"
                                type="text"
                                radiusType="semi-round"
                                colorType="gray"
                                backgroundColorType="white"
                                borderColorType="gray"
                                fontSize="12px"
                                height="40px"
                                width="200px"
                                onBlur={(e) => updateFormState('name', e.target.value)}
                                onChange={(e) => updateFormState('name', e.target.value)}
                                value={formFields.name}
                            />
                        </div>
                    </Form.Item>
                    <div className="schema-field tags">
                        <div className="title-icon-img">
                            <img className="icon" src={tagsIcon} />
                            <div className="title-desc">
                                <p className="field-title">Tags</p>
                                <p className="desc">Tags will help you organize, search and filter your data</p>
                            </div>
                        </div>
                        <Form.Item name="tags">
                            <TagsList
                                tagsToShow={3}
                                className="tags-list"
                                tags={tagsToDisplay}
                                newEntity={true}
                                addNew={true}
                                editable={true}
                                handleDelete={(tag) => removeTag(tag)}
                                handleTagsUpdate={(tags) => {
                                    updateFormState('tags', tags);
                                    setTagsToDisplay(tags);
                                }}
                            />
                        </Form.Item>
                    </div>
                    <Form.Item name="type" initialValue={formFields.type}>
                        <div className="schema-field type">
                            <div className="title-icon-img">
                                <img className="icon" src={schemaTypeIcon} alt="schemaTypeIcon" />
                                <div className="title-desc">
                                    <p className="field-title">Schema type</p>
                                    <p className="desc">Tags will help you organize, search and filter your data</p>
                                </div>
                            </div>
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
                    <div className="schema-field schema">
                        <div className="title-wrapper">
                            <p className="field-title">Schema definition</p>
                            <Button
                                width="115px"
                                height="34px"
                                placeholder="Validate"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                isLoading={validateLoading}
                                disabled={formFields?.schema_content === ''}
                                onClick={() => handleValidateSchema()}
                            />
                        </div>
                        <div className="editor">
                            <Form.Item
                                name="schema_content"
                                className="schema-item"
                                initialValue={SchemaEditorExample[formFields?.type]?.value}
                                rules={[
                                    {
                                        validator: checkContent
                                    }
                                ]}
                            >
                                <Editor
                                    options={{
                                        minimap: { enabled: false },
                                        scrollbar: { verticalScrollbarSize: 0 },
                                        scrollBeyondLastLine: false,
                                        roundedSelection: false,
                                        formatOnPaste: true,
                                        formatOnType: true,
                                        fontSize: '14px'
                                    }}
                                    height="calc(100% - 5px)"
                                    language={SchemaEditorExample[formFields?.type]?.language}
                                    defaultValue={SchemaEditorExample[formFields?.type]?.value}
                                    value={formFields.schema_content}
                                    onChange={(value) => updateFormState('schema_content', value)}
                                />
                            </Form.Item>
                            <div className={validateError || validateSuccess ? (validateSuccess ? 'validate-note success' : 'validate-note error') : 'validate-note'}>
                                {validateError && <ErrorOutlineRounded />}
                                {validateSuccess && <CheckCircleOutlineRounded />}
                                <p>{validateError || validateSuccess}</p>
                            </div>
                        </div>
                    </div>
                    <Form.Item className="button-container">
                        <Button
                            width="105px"
                            height="34px"
                            placeholder="Close"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => goBack()}
                        />
                        <Button
                            width="125px"
                            height="34px"
                            placeholder="Create schema"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            isLoading={loadingSubmit}
                            disabled={validateError}
                            onClick={handleSubmit}
                        />
                    </Form.Item>
                </div>
            </Form>
            <Modal
                header={<img src={errorModal} alt="errorModal" />}
                width="400px"
                height="300px"
                displayButtons={false}
                clickOutside={() => setModalOpen(false)}
                open={modalOpen}
            >
                <div className="roll-back-modal">
                    <p className="title">Too many message types specified in schema definition</p>
                    <p className="desc">Please choose your master message as a schema definition</p>
                    <SelectComponent
                        value={messageStructName}
                        colorType="black"
                        backgroundColorType="white"
                        borderColorType="gray-light"
                        radiusType="semi-round"
                        minWidth="12vw"
                        width="100%"
                        height="45px"
                        options={messagesStructNameList}
                        iconColor="gray"
                        popupClassName="message-option"
                        onChange={(e) => setMessageStructName(e)}
                    />
                    <div className="buttons">
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Close"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            border="gray-light"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => setModalOpen(false)}
                        />
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Confirm"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => handleCreateNewSchema()}
                        />
                    </div>
                </div>
            </Modal>
        </div>
    );
}

export default CreateSchema;
