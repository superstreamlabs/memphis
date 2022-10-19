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

const NewTagGenerator = ({ searchVal, allTags, handleFinish }) => {
    const [saveVisible, setSaveVisible] = useState(false);
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        name: ''
    });
    const [tagColor, setTagColor] = useState('purple');
    const colors = ['purple', 'magenta', 'red', 'orange', 'gold', 'lime', 'green', 'cyan', 'blue'];
    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    useEffect(() => {
        if (searchVal.length > 0) {
            updateFormState('name', searchVal);
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
                        color: tagColor
                    };
                    const res = await httpRequest('POST', ApiEndpoints.CREATE_NEW_TAG, data);
                    handleFinish(res);
                } catch (error) {}
            }
        }
    };

    const handleColorChange = (color) => {
        setTagColor(color);
    };

    return (
        <div className="new-tag-generator-wrapper">
            <Form name="form" form={creationForm} autoComplete="on" className="create-tag-form">
                <Form.Item
                    className="form-input"
                    name="name"
                    initialValue={searchVal ? searchVal : ''}
                    rules={[
                        {
                            required: true,
                            message: 'Please input tag name!'
                        }
                    ]}
                    style={{ height: '70px' }}
                >
                    <div className="tag-name">
                        <p className="field-title">
                            <span className="required-field-mark">* </span>Tag name <div className="color-circle" style={{ backgroundColor: tagColor }}></div>
                        </p>
                        <Input
                            placeholder={searchVal ? searchVal : 'Type tag name'}
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
                <div className="color-pick">
                    {/* <CirclePicker colors={colors} onChange={handleColorChange} /> */}
                    <ColorPicker colors={colors} onChange={handleColorChange} />
                    <Divider className="divider" />
                    {/* {colors.map((color) => (
                        <li key={color} className="color-picker">
                            <div className="color-picker" style={{ backgroundColor: color }}></div>
                        </li>
                    ))} */}
                </div>
                <Button
                    width={'200px'}
                    height="30px"
                    placeholder={`Create Tag`}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType={'purple'}
                    fontSize="14px"
                    fontWeight="bold"
                    htmlType="submit"
                    marginTop="20px"
                    marginLeft="30px"
                    marginBottom="5px"
                    onClick={onFinish}
                />
            </Form>
        </div>
    );
};

export default NewTagGenerator;
