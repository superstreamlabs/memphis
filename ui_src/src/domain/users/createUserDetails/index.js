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
import { Form } from 'antd';

import Input from '../../../components/Input';
import RadioButton from '../../../components/radioButton';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import SelectComponent from '../../../components/select';
import { generator } from '../../../services/generator';

const CreateUserDetails = ({ createUserRef, closeModal }) => {
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        username: '',
        password: '',
        user_type: 'management'
    });
    const [passwordType, setPasswordType] = useState(0);
    const userTypeOptions = ['management', 'application'];
    const passwordOptions = [
        {
            id: 1,
            value: 0,
            label: 'Default'
        },
        {
            id: 2,
            value: 1,
            label: 'Custom'
        }
    ];
    const [generatedPassword, setGeneratedPassword] = useState('');

    useEffect(() => {
        createUserRef.current = onFinish;
        generateNewPassword();
    }, []);

    const passwordTypeChange = (e) => {
        setPasswordType(e.target.value);
    };

    const handleUserNameChange = (e) => {
        setFormFields({ ...formFields, username: e.target.value });
    };

    const handlePasswordChange = (password) => {
        setFormFields({ ...formFields, password: password });
    };

    const handleSelectUserType = (e) => {
        setFormFields({ ...formFields, user_type: e });
    };

    const onFinish = async () => {
        const fieldsValue = await creationForm.validateFields();
        if (fieldsValue?.errorFields) {
            return;
        } else {
            if (fieldsValue?.passwordType === 0 ?? passwordType === 0) {
                fieldsValue['password'] = fieldsValue['generatedPassword'];
            }
            try {
                const bodyRequest = fieldsValue;
                const data = await httpRequest('POST', ApiEndpoints.ADD_USER, bodyRequest);
                if (data) {
                    closeModal(data);
                }
            } catch (error) {}
        }
    };

    const generateNewPassword = () => {
        const newPassword = generator();
        setGeneratedPassword(newPassword);
        creationForm.setFieldsValue({ ['generatedPassword']: newPassword });
    };

    return (
        <div className="create-user-form">
            <Form name="form" form={creationForm} autoComplete="off" onFinish={onFinish}>
                <Form.Item
                    name="username"
                    rules={[
                        {
                            required: true,
                            message: 'Please input username!'
                        },
                        {
                            message: 'Username has to include only letters/numbers and . or /',
                            pattern: new RegExp(/^[a-zA-Z0-9_.]*$/)
                        }
                    ]}
                >
                    <div className="field username">
                        <p>
                            <span className="required-field-mark">* </span>Username
                        </p>
                        <Input
                            placeholder="Type username"
                            type="text"
                            radiusType="semi-round"
                            colorType="black"
                            backgroundColorType="none"
                            borderColorType="gray"
                            height="40px"
                            fontSize="12px"
                            onBlur={handleUserNameChange}
                            onChange={handleUserNameChange}
                            value={formFields.name}
                        />
                    </div>
                </Form.Item>
                <div className="field user-type">
                    <p>Type</p>
                    <div className="field username">
                        <Form.Item name="user_type" initialValue={formFields.user_type}>
                            <SelectComponent
                                value={formFields.user_type}
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                options={userTypeOptions}
                                onChange={(e) => handleSelectUserType(e)}
                                popupClassName="select-options"
                            />
                        </Form.Item>
                    </div>
                </div>
                {formFields.user_type === 'management' && (
                    <div className="password-section">
                        <p>Password</p>
                        <Form.Item name="passwordType" initialValue={passwordType}>
                            <RadioButton options={passwordOptions} radioValue={passwordType} onChange={(e) => passwordTypeChange(e)} />
                        </Form.Item>

                        {passwordType === 0 && (
                            <Form.Item name="generatedPassword" initialValue={generatedPassword}>
                                <div className="field password">
                                    <Input
                                        type="text"
                                        disabled
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        fontSize="12px"
                                        value={generatedPassword}
                                    />
                                    <p className="generate-password-button" onClick={() => generateNewPassword()}>
                                        Generate again
                                    </p>
                                </div>
                            </Form.Item>
                        )}
                        {passwordType === 1 && (
                            <div>
                                <div className="field password">
                                    <p>Type password</p>
                                    <Form.Item
                                        name="password"
                                        rules={[
                                            {
                                                required: true,
                                                message: 'Password can not be empty'
                                            }
                                        ]}
                                    >
                                        <Input
                                            placeholder="Type Password"
                                            type="password"
                                            radiusType="semi-round"
                                            colorType="black"
                                            backgroundColorType="none"
                                            borderColorType="gray"
                                            height="40px"
                                            fontSize="12px"
                                        />
                                    </Form.Item>
                                </div>
                                <div className="field confirm">
                                    <p>Confirm Password</p>
                                    <Form.Item
                                        name="confirm"
                                        validateTrigger="onChange"
                                        dependencies={['password']}
                                        rules={[
                                            {
                                                required: true,
                                                message: 'Confirm password can not be empty'
                                            },
                                            ({ getFieldValue }) => ({
                                                validator(rule, value) {
                                                    if (!value || getFieldValue('password') === value) {
                                                        handlePasswordChange(value);
                                                        return Promise.resolve();
                                                    }
                                                    return Promise.reject('Passwords do not match');
                                                }
                                            })
                                        ]}
                                    >
                                        <Input
                                            placeholder="Type Password"
                                            type="password"
                                            radiusType="semi-round"
                                            colorType="black"
                                            backgroundColorType="none"
                                            borderColorType="gray"
                                            height="40px"
                                            fontSize="12px"
                                        />
                                    </Form.Item>
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </Form>
        </div>
    );
};

export default CreateUserDetails;
