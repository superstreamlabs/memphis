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
import React, { useEffect, useState } from 'react';
import { Form, message, Divider } from 'antd';
import Button from '../../button';
import Input from '../../Input';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import ColorPicker from '../../colorPicker';
import { ColorPalette } from '../../../const/globalConst';
import { showMessages } from '../../../services/genericServices';

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

    useEffect(() => {
        const keyDownHandler = (event) => {
            if (event.key === 'Enter') {
                event.preventDefault();
                onFinish();
            }
        };
        document.addEventListener('keydown', keyDownHandler);
        return () => {
            document.removeEventListener('keydown', keyDownHandler);
        };
    }, []);

    const onFinish = async () => {
        const values = await creationForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            if (allTags.some((tag) => formFields.name === tag.name)) {
                showMessages('warning', 'Tag with this name already exists');
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
