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
        name: '',
        color: ColorPalette[0] //default memphis-purple
    });

    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    useEffect(() => {
        if (searchVal.length > 0) {
            updateFormState('name', searchVal);
            updateFormState('color', ColorPalette[0]);
        }
    }, []);

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
                        }
                    ]}
                    style={{ height: '70px' }}
                >
                    <div className="tag-name">
                        <p className="field-title">Tag</p>
                        <Input
                            placeholder={searchVal || 'Enter tag here'}
                            type="text"
                            radiusType="semi-round"
                            colorType="black"
                            backgroundColorType="none"
                            borderColorType="gray"
                            height="40px"
                            onBlur={(e) => updateFormState('name', e.target.value)}
                            onChange={(e) => updateFormState('name', e.target.value)}
                            value={formFields.name}
                        />
                    </div>
                </Form.Item>
                {/* <div className="color-pick"> */}
                <Form.Item className="form-input" name="color">
                    <ColorPicker onChange={(value) => updateFormState('color', value)} value={formFields.color} />
                </Form.Item>
                {/* </div> */}
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
