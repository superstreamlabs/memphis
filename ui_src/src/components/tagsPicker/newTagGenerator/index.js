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
import Button from '../../button';
import Input from '../../Input';
import { CirclePicker } from 'react-color';

const NewTagGenerator = ({ searchVal, allTags, handleFinish }) => {
    const [saveVisible, setSaveVisible] = useState(false);
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        name: ''
    });
    const [tagColor, setTagColor] = useState('#00A5FF');
    const colors = [
        '#00A5FF',
        '#e91e63',
        '#9c27b0',
        '#673ab7',
        '#3f51b5',
        '#2196f3',
        '#03a9f4',
        '#00bcd4',
        '#009688',
        '#4caf50',
        '#8bc34a',
        '#cddc39',
        '#ffeb3b',
        '#ffc107',
        '#ff9800',
        '#ff5722',
        '#795548',
        '#607d8b'
    ];
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
        } else if (!saveVisible) {
            return;
        } else {
            handleFinish({
                name: values.name,
                color: tagColor
            });
        }
    };

    useEffect(() => {
        if (formFields.name !== '') {
            if (allTags.some((tag) => formFields.name === tag.name)) {
                setSaveVisible(false);
            } else {
                setSaveVisible(true);
            }
        } else {
            setSaveVisible(false);
        }
    }, [formFields.name]);

    const handleColorChange = (color) => {
        setTagColor(color.hex);
    };

    return (
        <div className="new-tag-generator-wrapper">
            <Form name="form" form={creationForm} autoComplete="on" onFinish={onFinish} className="create-tag-form">
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
                    <CirclePicker colors={colors} onChange={handleColorChange} />
                </div>
                {saveVisible && (
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
                        marginTop="10px"
                        marginLeft="20px"
                        marginBottom="5px"
                        onClick={onFinish}
                    />
                )}
            </Form>
        </div>
    );
};

export default NewTagGenerator;
