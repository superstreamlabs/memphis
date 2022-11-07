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
import React, { useEffect, useState } from 'react';
import { Form, message, Divider } from 'antd';
import Button from '../../button';
import Input from '../../Input';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import ColorPicker from '../../colorPicker';
import { ColorPalette } from '../../../const/colorPalette';

const NewTagGenerator = ({ searchVal, allTags, handleFinish, handleCancel }) => {
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        name: searchVal,
        color: ColorPalette[0]
    });

    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    const onFinish = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            if (allTags.some((tag) => formFields.name === tag.name)) {
                message.warning({
                    key: 'memphisWarningMessage',
                    content: 'Tag with this name already exists',
                    duration: 5,
                    style: { cursor: 'pointer' },
                    onClick: () => message.destroy('memphisWarningMessage')
                });
            } else {
                try {
                    let data = {
                        name: values.name,
                        color: values.color
                    };
                    const res = await httpRequest('POST', ApiEndpoints.CREATE_NEW_TAG, data);
                    handleFinish(res);
                } catch (error) {}
            }
        }
    };

    return (
        <div className="new-tag-generator-wrapper">
            <Form name="form" form={creationForm} autoComplete="on" className="create-tag-form">
                <Form.Item
                    className="form-input"
                    name="name"
                    initialValue={searchVal || ''}
                    rules={[
                        {
                            required: true,
                            message: 'Please input tag name!'
                        },
                        {
                            max: 20,
                            message: `Can't be longer than 20!`
                        }
                    ]}
                    style={{ height: '70px' }}
                >
                    <div className="tag-name">
                        <p className="field-title">Tag</p>
                        <Input
                            placeholder={'Enter tag here'}
                            type="text"
                            radiusType="semi-round"
                            colorType="black"
                            backgroundColorType="none"
                            borderColorType="gray"
                            height="40px"
                            onBlur={(e) => updateFormState('name', e.target.value)}
                            onChange={(e) => updateFormState('name', e.target.value)}
                            value={formFields.name}
                            onPressEnter={onFinish}
                        />
                    </div>
                </Form.Item>
                <Form.Item className="form-input" name="color" initialValue={ColorPalette[0]}>
                    <ColorPicker onChange={(value) => updateFormState('color', value)} value={formFields.color} />
                </Form.Item>
                <Divider className="divider" />
                <div className="save-cancel-buttons">
                    <Button
                        width={'80px'}
                        height="36px"
                        placeholder={`Cancel`}
                        colorType="black"
                        radiusType="semi-round"
                        backgroundColorType={'white'}
                        border="gray-light"
                        fontSize="14px"
                        fontWeight="bold"
                        marginBottom="5px"
                        onClick={handleCancel}
                    />
                    <Button
                        width={'60px'}
                        height="36px"
                        placeholder={`Add`}
                        colorType="white"
                        radiusType="semi-round"
                        backgroundColorType={'purple'}
                        fontSize="14px"
                        fontWeight="bold"
                        marginBottom="5px"
                        onClick={onFinish}
                    />
                </div>
            </Form>
        </div>
    );
};

export default NewTagGenerator;
