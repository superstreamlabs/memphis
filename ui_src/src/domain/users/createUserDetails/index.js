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

import React, { useContext, useEffect, useState } from 'react';
import { Form } from 'antd';
import { HiLockClosed } from 'react-icons/hi';
import Input from '../../../components/Input';
import RadioButton from '../../../components/radioButton';
import Button from '../../../components/button';
import SelectComponent from '../../../components/select';
import { httpRequest } from '../../../services/http';
import { useGetAllowedActions } from '../../../services/genericServices';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import SelectCheckBox from '../../../components/selectCheckBox';
import { generator } from '../../../services/generator';
import { ReactComponent as RefreshIcon } from '../../../assets/images/refresh.svg';

import { LOCAL_STORAGE_USER_PASS_BASED_AUTH } from '../../../const/localStorageConsts';
import { isCloud, showUpgradePlan } from '../../../services/valueConvertor';
import { Context } from '../../../hooks/store';
import UpgradePlans from '../../../components/upgradePlans';

const CreateUserDetails = ({ createUserRef, closeModal, handleLoader, userList, isLoading, clientType = false }) => {
    const [state, dispatch] = useContext(Context);
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        username: '',
        password: ''
    });
    const [userType, setUserType] = useState(clientType ? 'application' : 'management');
    const [userViolation, setUserViolation] = useState(false);
    const userTypeOptions = [
        {
            id: 1,
            value: 'management',
            label: 'Management',
            desc: 'For management and console access',
            disabled: false
        },
        {
            id: 2,
            value: 'application',
            label: 'Client',
            desc: 'For client-based authentication with the broker',
            disabled: false
        }
    ];

    const getAllowedActions = useGetAllowedActions();
    useEffect(() => {
        createUserRef.current = onFinish;
    }, []);

    const updateFormState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    const checkPlanViolation = () => {
        const usersLimits = state?.userData?.entitlements && state?.userData?.entitlements['feature-management-users']?.limits;
        const usersExceeded = userList?.management_users?.length === usersLimits;
        setUserViolation(usersExceeded);

        return !usersExceeded;
    };

    const onFinish = async () => {
        try {
            let canCreate = isCloud() ? false : true;
            const fieldsValue = await creationForm.validateFields();
            if (fieldsValue?.errorFields) {
                handleLoader(false);
                return;
            } else {
                if (isCloud()) canCreate = checkPlanViolation(formFields);
                if (!canCreate) {
                    handleLoader(false);
                    return;
                }
                try {
                    const bodyRequest = fieldsValue;
                    const data = await httpRequest('POST', ApiEndpoints.ADD_USER, bodyRequest);
                    if (data) {
                        closeModal(data);
                    }
                } catch (error) {
                    handleLoader(false);
                } finally {
                    getAllowedActions();
                }
            }
        } catch (error) {
            handleLoader(false);
        }
    };

    const generateNewPassword = () => {
        const newPassword = generator();
        updateFormState('password', newPassword);
        creationForm.setFieldsValue({ ['password']: newPassword });
    };

    const handleUserTypeChanged = (value) => {
        setUserType(value);
        creationForm.setFieldValue('user_type', value);
    };

    return (
        <div className="create-user-form">
            <Form name="form" className="user-form" form={creationForm} autoComplete="off">
                <div>
                    <div className="field user-type">
                        <Form.Item name="user_type" initialValue={userType}>
                            <SelectCheckBox
                                vertical
                                selectOptions={clientType ? userTypeOptions?.filter((type) => type.value === 'application') : userTypeOptions}
                                handleOnClick={(e) => handleUserTypeChanged(e.value)}
                                selectedOption={userType}
                            />
                        </Form.Item>
                    </div>
                    <div className="user-details">
                        <p className="fields-title">User details</p>
                        <Form.Item
                            name="username"
                            rules={[
                                {
                                    required: true,
                                    message: userType === 'management' && isCloud() ? 'Please input email!' : 'Please input username!'
                                },
                                {
                                    message:
                                        userType === 'management' && isCloud()
                                            ? 'Please enter a valid email address!'
                                            : 'Username has to include only letters/numbers and .',
                                    pattern: userType === 'management' && isCloud() ? /^[a-zA-Z0-9._%]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/ : /^[a-zA-Z0-9_.]*$/
                                }
                            ]}
                        >
                            <div className="field username">
                                <p className="field-title">{userType === 'management' && isCloud() ? 'Email*' : 'Username*'}</p>
                                <Input
                                    placeholder={userType === 'management' && isCloud() ? 'Type email' : 'Type username'}
                                    type="text"
                                    radiusType="semi-round"
                                    maxLength={60}
                                    colorType="black"
                                    backgroundColorType="none"
                                    borderColorType="gray"
                                    height="40px"
                                    fontSize="12px"
                                    onBlur={(e) => updateFormState('username', e.target.value)}
                                    onChange={(e) => updateFormState('username', e.target.value)}
                                    value={formFields.name}
                                />
                            </div>
                        </Form.Item>
                        {userType === 'management' && (
                            <>
                                {userType === 'management' && (
                                    <Form.Item
                                        name="full_name"
                                        rules={[
                                            {
                                                required: isCloud() ? true : false,
                                                message: 'Please input full name!'
                                            },
                                            {
                                                message: 'Please enter a valid full name!',
                                                pattern: /^[A-Za-z\s]+$/i
                                            }
                                        ]}
                                    >
                                        <div className="field fullname">
                                            <p className="field-title">{isCloud() ? 'Full name*' : 'Full name'}</p>
                                            <Input
                                                placeholder="Type full name"
                                                type="text"
                                                maxLength={30}
                                                radiusType="semi-round"
                                                colorType="black"
                                                backgroundColorType="none"
                                                borderColorType="gray"
                                                height="40px"
                                                fontSize="12px"
                                                onBlur={(e) => updateFormState('full_name', e.target.value)}
                                                onChange={(e) => updateFormState('full_name', e.target.value)}
                                                value={formFields.full_name}
                                            />
                                        </div>
                                    </Form.Item>
                                )}
                                <div className="user-row">
                                    <Form.Item name="team">
                                        <div className="field team">
                                            <p className="field-title">Team</p>
                                            <Input
                                                placeholder="Type your team"
                                                type="text"
                                                maxLength={20}
                                                radiusType="semi-round"
                                                colorType="black"
                                                backgroundColorType="none"
                                                borderColorType="gray"
                                                height="40px"
                                                fontSize="12px"
                                                onBlur={(e) => updateFormState('team', e.target.value)}
                                                onChange={(e) => updateFormState('team', e.target.value)}
                                                value={formFields.team}
                                            />
                                        </div>
                                    </Form.Item>
                                    <Form.Item name="position">
                                        <div className="field position">
                                            <p className="field-title">Position</p>
                                            <Input
                                                placeholder="Type your position"
                                                type="text"
                                                maxLength={30}
                                                radiusType="semi-round"
                                                colorType="black"
                                                backgroundColorType="none"
                                                borderColorType="gray"
                                                height="40px"
                                                fontSize="12px"
                                                onBlur={(e) => updateFormState('position', e.target.value)}
                                                onChange={(e) => updateFormState('position', e.target.value)}
                                                value={formFields.position}
                                            />
                                        </div>
                                    </Form.Item>
                                </div>
                            </>
                        )}
                        {userType === 'application' && (
                            <>
                                <Form.Item name="description">
                                    <div className="field description">
                                        <p className="field-title">Description</p>
                                        <Input
                                            placeholder="Type your description"
                                            type="text"
                                            maxLength={100}
                                            radiusType="semi-round"
                                            colorType="black"
                                            backgroundColorType="none"
                                            borderColorType="gray"
                                            height="40px"
                                            fontSize="12px"
                                            onBlur={(e) => updateFormState('description', e.target.value)}
                                            onChange={(e) => updateFormState('description', e.target.value)}
                                            value={formFields.description}
                                        />
                                    </div>
                                </Form.Item>
                            </>
                        )}
                    </div>

                    {((userType === 'management' && !isCloud()) ||
                        (userType === 'application' && localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true')) && (
                        <div className="password-section">
                            <p className="fields-title">Set password</p>
                            <Form.Item
                                name="password"
                                rules={[
                                    {
                                        required: true,
                                        message: 'Password can not be empty'
                                    },
                                    {
                                        pattern: /^(?=.*[A-Z])(?=.*[a-z])(?=.*\d)(?=.*[!?\-@#$%])[A-Za-z\d!?\-@#$%]{8,}$/,
                                        message:
                                            'Password must be at least 8 characters long, contain both uppercase and lowercase, and at least one number and one special character(!?-@#$%)'
                                    }
                                ]}
                            >
                                <div className="field password">
                                    <div className="password-title">
                                        <p className="field-title">Set password or generate one</p>
                                        <span className="generate-btn" onClick={generateNewPassword}>
                                            <RefreshIcon width={14} />
                                            <p className="generate-password-button">Generate password</p>
                                        </span>
                                    </div>
                                    <Input
                                        type="text"
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        fontSize="12px"
                                        value={formFields?.password}
                                        onChange={(e) => {
                                            updateFormState('password', e.target.value);
                                            // setGeneratedPassword(e.target.value);
                                            // creationForm.setFieldsValue({ ['generatedPassword']: e.target.value });
                                        }}
                                        onBlur={(e) => {
                                            updateFormState('password', e.target.value);
                                        }}
                                    />
                                </div>
                            </Form.Item>
                        </div>
                    )}
                    <>
                        <div className="user-details">
                            <span className="coming-soon-container">
                                <p className="fields-title">Roles</p>
                                <label className="coming-soon">Coming soon</label>
                            </span>
                            <SelectComponent
                                colorType="black"
                                backgroundColorType="light-gray"
                                fontFamily="Inter"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                fontSize="12px"
                                popupClassName="select-options"
                                placeholder="Select role"
                                disabled
                            />
                            <span className="coming-soon-container">
                                <p className="fields-title">Permissions</p>
                                <label className="coming-soon">Coming soon</label>
                            </span>
                            <SelectComponent
                                colorType="black"
                                backgroundColorType="light-gray"
                                fontFamily="Inter"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                fontSize="12px"
                                popupClassName="select-options"
                                placeholder="Select permissions"
                                disabled
                            />
                            <span className="coming-soon-container">
                                <p className="fields-title">Tenants</p>
                                <label className="coming-soon">Coming soon</label>
                            </span>
                            <SelectComponent
                                colorType="black"
                                backgroundColorType="light-gray"
                                fontFamily="Inter"
                                borderColorType="gray"
                                radiusType="semi-round"
                                height="40px"
                                fontSize="12px"
                                popupClassName="select-options"
                                placeholder="Select tenants"
                                disabled
                            />
                        </div>
                    </>
                    {userViolation && (
                        <div className="show-violation-form">
                            <div className="flex-line">
                                <HiLockClosed className="lock-feature-icon" />
                                <p>Your current plan allows {state?.userData?.entitlements['feature-management-users']?.limits} management users</p>
                            </div>
                            {showUpgradePlan() && (
                                <UpgradePlans
                                    content={
                                        <div className="upgrade-button-wrapper">
                                            <p className="upgrade-plan">Upgrade now</p>
                                        </div>
                                    }
                                    isExternal={false}
                                />
                            )}
                        </div>
                    )}
                </div>
                <Button
                    placeholder={'Add'}
                    colorType={'white'}
                    onClick={onFinish}
                    fontSize={'14px'}
                    fontWeight={500}
                    border="none"
                    backgroundColorType={'purple'}
                    height="40px"
                    width="50%"
                    radiusType="circle"
                    isLoading={isLoading}
                />
            </Form>
        </div>
    );
};

export default CreateUserDetails;
