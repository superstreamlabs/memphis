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

import React, { useEffect, useState } from 'react';
import { Form } from 'antd';

import Input from '../../../components/Input';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { useHistory } from 'react-router-dom';
import pathDomains from '../../../router';

const CreateFactoryDetails = (props) => {
    const { createFactoryRef } = props;
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        name: '',
        description: ''
    });
    const history = useHistory();

    useEffect(() => {
        createFactoryRef.current = onFinish;
    }, []);

    const handleFactoryNameChange = (e) => {
        setFormFields({ ...formFields, name: e.target.value });
    };
    const handleDescriptionNameChange = (e) => {
        setFormFields({ ...formFields, description: e.target.value });
    };

    const onFinish = async () => {
        const fieldsError = await creationForm.validateFields();
        if (fieldsError?.errorFields) {
            return;
        } else {
            try {
                const bodyRequest = creationForm.getFieldsValue();
                const data = await httpRequest('POST', ApiEndpoints.CREATE_FACTORY, bodyRequest);
                if (data) {
                    history.push(`${pathDomains.factoriesList}/${data.name}`);
                }
            } catch (error) {}
        }
    };

    return (
        <div className="create-factory-form">
            <Form name="form" form={creationForm} autoComplete="off">
                <Form.Item
                    name="name"
                    rules={[
                        {
                            required: true,
                            message: 'Please input factory name!'
                        }
                    ]}
                >
                    <div className="field name">
                        <p>
                            <span className="required-field-mark">* </span>Factory name
                        </p>
                        <Input
                            placeholder="Type factory name"
                            type="text"
                            radiusType="semi-round"
                            colorType="black"
                            backgroundColorType="none"
                            borderColorType="gray"
                            height="40px"
                            fontSize="12px"
                            onBlur={handleFactoryNameChange}
                            onChange={handleFactoryNameChange}
                            value={formFields.name}
                        />
                    </div>
                </Form.Item>
                <Form.Item name="description">
                    <div className="field description">
                        <p>Factory description</p>
                        <Input
                            placeholder="Type factory name"
                            type="textArea"
                            radiusType="semi-round"
                            colorType="black"
                            backgroundColorType="none"
                            borderColorType="gray"
                            numberOfRows="5"
                            fontSize="12px"
                            onBlur={handleDescriptionNameChange}
                            onChange={handleDescriptionNameChange}
                            value={formFields.description}
                        />
                    </div>
                </Form.Item>
            </Form>
        </div>
    );
};

export default CreateFactoryDetails;
