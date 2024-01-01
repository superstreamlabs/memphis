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
import { ArrowDropDownRounded } from '@material-ui/icons';
import Input from 'components/Input';
import Button from 'components/button';
import SelectComponent from 'components/select';
import { Select } from 'antd';
import { httpRequest } from 'services/http';
import { useGetAllowedActions } from 'services/genericServices';
import { ApiEndpoints } from 'const/apiEndpoints';
import SelectCheckBox from 'components/selectCheckBox';
import RadioButton from 'components/radioButton';
import { generator } from 'services/generator';
import { ReactComponent as RefreshIcon } from 'assets/images/refresh.svg';

import { LOCAL_STORAGE_USER_PASS_BASED_AUTH } from 'const/localStorageConsts';
import { isCloud, showUpgradePlan } from 'services/valueConvertor';
import { Context } from 'hooks/store';
import UpgradePlans from 'components/upgradePlans';

const CreateUserDetails = ({ createUserRef, closeModal, handleLoader, userList, isLoading, clientType = false, selectedRow }) => {
    const [state, dispatch] = useContext(Context);
    const [creationForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        username: '',
        password: ''
    });
    const [userType, setUserType] = useState(selectedRow?.user_type === 'application' ? 'application' : clientType ? 'application' : 'management');
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
    const rbacTypeOptions = [
        {
            id: 1,
            value: 'pattern',
            label: 'Pattern'
        },
        {
            id: 2,
            value: 'stations',
            label: 'Stations'
        }
    ];
    const [stationsList, setStationsList] = useState([]);
    const [rbacTypeWrite, setRbacTypeWrite] = useState('pattern');
    const [rbacTypeRead, setRbacTypeRead] = useState('pattern');
    const [isDisabled, setIsDisabled] = useState(false);

    const getAllowedActions = useGetAllowedActions();

    useEffect(() => {
        getAllStations();
        createUserRef.current = onFinish;
    }, []);

    useEffect(() => {
        selectedRow && setIsDisabled(true);
    }, [selectedRow]);

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
            console.log('formFields', creationForm.getFieldsValue());
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
                    handleLoader(true);
                    const bodyRequest = fieldsValue;
                    if (userType === 'application') {
                        bodyRequest['allow_read_permissions'] =
                            formFields?.allow_read_permissions?.length === 0 ||
                            formFields?.allow_read_permissions === null ||
                            formFields?.allow_read_permissions === undefined
                                ? null
                                : formFields?.allow_read_permissions;
                        bodyRequest['allow_write_permissions'] =
                            formFields?.allow_write_permissions?.length === 0 ||
                            formFields?.allow_write_permissions === null ||
                            formFields?.allow_write_permissions === undefined
                                ? null
                                : formFields?.allow_write_permissions;
                    }
                    const data = await httpRequest('POST', ApiEndpoints.ADD_USER, bodyRequest);
                    if (data) {
                        closeModal(data);
                    }
                } catch (error) {
                    handleLoader(false);
                } finally {
                    handleLoader(false);
                    getAllowedActions();
                }
            }
        } catch (error) {
            handleLoader(false);
        }
    };

    const getAllStations = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_STATIONS}`);
            setStationsList(
                res?.stations?.map((station) => {
                    return { label: station?.station?.name, value: station?.station?.name };
                })
            );
        } catch (err) {
            return;
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
                <div className="fields-section">
                    <div className="field user-type">
                        <Form.Item name="user_type" initialValue={userType}>
                            <SelectCheckBox
                                vertical
                                selectOptions={clientType ? userTypeOptions?.filter((type) => type.value === 'application') : userTypeOptions}
                                handleOnClick={(e) => handleUserTypeChanged(e.value)}
                                selectedOption={userType}
                                disabled={isDisabled}
                            />
                        </Form.Item>
                    </div>
                    <div className="form-section">
                        <p className="fields-title">User details</p>
                        {userType === 'management' && isCloud() && (
                            <div className="form-section-row">
                                <Form.Item
                                    name="username"
                                    rules={[
                                        {
                                            required: true,
                                            message: 'Please input email!'
                                        },
                                        {
                                            message: 'Please enter a valid email address!',
                                            pattern: /^[a-zA-Z0-9._%]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
                                        }
                                    ]}
                                >
                                    <p className="field-title">Email*</p>
                                    <Input
                                        placeholder="Type email"
                                        type="text"
                                        radiusType="semi-round"
                                        maxLength={60}
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('username', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('username', e.target.value);
                                            creationForm.setFieldsValue({ username: e.target.value });
                                        }}
                                        value={formFields?.username || selectedRow?.username}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                                <Form.Item
                                    name="full_name"
                                    rules={[
                                        {
                                            required: true,
                                            message: 'Please input full name!'
                                        },
                                        {
                                            message: 'Please enter a valid full name!',
                                            pattern: /^[A-Za-z\s]+$/i
                                        }
                                    ]}
                                >
                                    <p className="field-title">Full name*</p>
                                    <Input
                                        placeholder="Type full name"
                                        type="text"
                                        maxLength={30}
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('full_name', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('full_name', e.target.value);
                                            creationForm.setFieldsValue({ full_name: e.target.value });
                                        }}
                                        value={formFields?.full_name || selectedRow?.full_name}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                            </div>
                        )}
                        {userType === 'management' && isCloud() && (
                            <div className="form-section-row">
                                <Form.Item name="team">
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
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('team', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('team', e.target.value);
                                            creationForm.setFieldsValue({ team: e.target.value });
                                        }}
                                        value={formFields?.team || selectedRow?.team}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                                <Form.Item name="position">
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
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('position', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('position', e.target.value);
                                            creationForm.setFieldsValue({ position: e.target.value });
                                        }}
                                        value={formFields?.position || selectedRow?.position}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                            </div>
                        )}
                        {userType === 'management' && !isCloud() && (
                            <div className="form-section-row">
                                <Form.Item
                                    name="username"
                                    rules={[
                                        {
                                            required: true,
                                            message: 'Please input username!'
                                        },
                                        {
                                            message: 'Username has to include only letters/numbers and .',
                                            pattern: /^[a-zA-Z0-9_.]*$/
                                        }
                                    ]}
                                >
                                    <p className="field-title">Username*</p>
                                    <Input
                                        placeholder={'Type username'}
                                        type="text"
                                        radiusType="semi-round"
                                        maxLength={60}
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('username', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('username', e.target.value);
                                            creationForm.setFieldsValue({ username: e.target.value });
                                        }}
                                        value={formFields?.username || selectedRow?.username}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
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
                                    <div className="password-title">
                                        <p className="field-title">Set password or generate one*</p>
                                        <span className="generate-btn" onClick={generateNewPassword}>
                                            <RefreshIcon width={14} />
                                            <p className="generate-password-button field-title">Generate</p>
                                        </span>
                                    </div>
                                    <Input
                                        type="text"
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        width="100%"
                                        fontSize="12px"
                                        value={selectedRow ? '*******' : formFields?.password}
                                        onChange={(e) => {
                                            updateFormState('password', e.target.value);
                                            creationForm.setFieldsValue({ ['password']: e.target.value });
                                        }}
                                        onBlur={(e) => {
                                            updateFormState('password', e.target.value);
                                        }}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                            </div>
                        )}
                        {userType === 'management' && !isCloud() && (
                            <div className="form-section-row">
                                <Form.Item
                                    name="full_name"
                                    rules={[
                                        {
                                            message: 'Please enter a valid full name!',
                                            pattern: /^[A-Za-z\s]+$/i
                                        }
                                    ]}
                                >
                                    <p className="field-title">Full name</p>
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
                                        onChange={(e) => {
                                            updateFormState('full_name', e.target.value);
                                            creationForm.setFieldsValue({ full_name: e.target.value });
                                        }}
                                        value={formFields?.full_name || selectedRow?.full_name}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                                <Form.Item name="team">
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
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('team', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('team', e.target.value);
                                            creationForm.setFieldsValue({ team: e.target.value });
                                        }}
                                        value={formFields?.team || selectedRow?.team}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                                <Form.Item name="position">
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
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('position', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('position', e.target.value);
                                            creationForm.setFieldsValue({ position: e.target.value });
                                        }}
                                        value={formFields?.position || selectedRow?.position}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                            </div>
                        )}
                        {userType === 'application' && (
                            <div className="form-section-row">
                                <Form.Item
                                    name="username"
                                    rules={[
                                        {
                                            required: true,
                                            message: 'Please input username!'
                                        },
                                        {
                                            message: 'Username has to include only letters/numbers and .',
                                            pattern: /^[a-zA-Z0-9_.]*$/
                                        }
                                    ]}
                                >
                                    <p className="field-title">Username*</p>
                                    <Input
                                        placeholder={'Type username'}
                                        type="text"
                                        radiusType="semi-round"
                                        maxLength={60}
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('username', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('username', e.target.value);
                                            creationForm.setFieldsValue({ username: e.target.value });
                                        }}
                                        value={formFields?.username || selectedRow?.username}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                                {localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true' && (
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
                                        <div className="password-title">
                                            <p className="field-title">Set password or generate one*</p>
                                            <span className="generate-btn" onClick={generateNewPassword}>
                                                <RefreshIcon width={14} />
                                                <p className="generate-password-button field-title">Generate</p>
                                            </span>
                                        </div>
                                        <Input
                                            type="text"
                                            radiusType="semi-round"
                                            colorType="black"
                                            backgroundColorType="none"
                                            borderColorType="gray"
                                            height="40px"
                                            width="100%"
                                            fontSize="12px"
                                            value={selectedRow ? '*******' : formFields?.password}
                                            onChange={(e) => {
                                                updateFormState('password', e.target.value);
                                                creationForm.setFieldsValue({ password: e.target.value });
                                            }}
                                            onBlur={(e) => {
                                                updateFormState('password', e.target.value);
                                            }}
                                            disabled={isDisabled}
                                        />
                                    </Form.Item>
                                )}
                            </div>
                        )}
                        {userType === 'application' && (
                            <div className="form-section-row">
                                <Form.Item name="description">
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
                                        width="100%"
                                        fontSize="12px"
                                        onBlur={(e) => updateFormState('description', e.target.value)}
                                        onChange={(e) => {
                                            updateFormState('description', e.target.value);
                                            creationForm.setFieldsValue({ description: e.target.value });
                                        }}
                                        value={formFields?.description || selectedRow?.description}
                                        disabled={isDisabled}
                                    />
                                </Form.Item>
                            </div>
                        )}
                    </div>
                    {userType === 'management' && (
                        <>
                            <div className="user-details">
                                <span className="fields-title-container">
                                    <p className="fields-title">Roles</p>
                                    <label className="coming-soon">Coming soon</label>
                                </span>
                                <SelectComponent
                                    suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                                    colorType="black"
                                    backgroundColorType="none"
                                    fontFamily="Inter"
                                    borderColorType="gray"
                                    radiusType="semi-round"
                                    height="40px"
                                    fontSize="12px"
                                    popupClassName="select-options"
                                    placeholder="Select role"
                                    disabled
                                />
                                <span className="fields-title-container">
                                    <p className="fields-title">Permissions</p>
                                    <label className="coming-soon">Coming soon</label>
                                </span>
                                <SelectComponent
                                    suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                                    colorType="black"
                                    backgroundColorType="none"
                                    fontFamily="Inter"
                                    borderColorType="gray"
                                    radiusType="semi-round"
                                    height="40px"
                                    fontSize="12px"
                                    popupClassName="select-options"
                                    placeholder="Select permissions"
                                    disabled
                                />
                                <span className="fields-title-container">
                                    <p className="fields-title">Tenants</p>
                                    <label className="coming-soon">Coming soon</label>
                                </span>
                                <SelectComponent
                                    suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                                    colorType="black"
                                    backgroundColorType="none"
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
                    )}

                    {userType === 'application' && (
                        <div className="form-section">
                            <span className="fields-title-container">
                                <p className="fields-title">Can read from (R)</p>
                                <RadioButton
                                    className="radio-button"
                                    options={rbacTypeOptions}
                                    radioValue={rbacTypeRead}
                                    fontFamily="InterSemiBold"
                                    style={{ marginRight: '20px', content: '' }}
                                    onChange={(e) => setRbacTypeRead(e.target.value)}
                                    disabled={isDisabled}
                                />
                            </span>
                            <div className="form-section-row">
                                <Form.Item
                                    name="allow_read_permissions"
                                    rules={
                                        rbacTypeRead === 'pattern' && [
                                            {
                                                pattern: /^[a-zA-Z0-9_\-., ]+(\..*)?|\*$/,
                                                message: `Only alphanumeric and the '_', '-', '.', '*' characters are allowed`
                                            }
                                        ]
                                    }
                                >
                                    {rbacTypeRead === 'stations' ? (
                                        <Select
                                            suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                                            showArrow
                                            mode="multiple"
                                            style={{ width: '100%' }}
                                            options={stationsList || []}
                                            popupClassName="select-options"
                                            onChange={(e) => {
                                                updateFormState('allow_read_permissions', e);
                                                creationForm.setFieldsValue({ allow_read_permissions: e });
                                            }}
                                            disabled={isDisabled}
                                        />
                                    ) : (
                                        <Select
                                            suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                                            showArrow
                                            mode="tags"
                                            placeholder={'*'}
                                            value={selectedRow?.permissions?.allow_read_permissions || []}
                                            onChange={(e) => {
                                                updateFormState('allow_read_permissions', e);
                                                creationForm.setFieldsValue({ allow_read_permissions: e });
                                            }}
                                            style={{ width: '100%' }}
                                            popupClassName="select-options"
                                            disabled={isDisabled}
                                        ></Select>
                                    )}
                                </Form.Item>
                            </div>
                        </div>
                    )}
                    {userType === 'application' && (
                        <div className="form-section">
                            <span className="fields-title-container">
                                <p className="fields-title">Can write to (W)</p>
                                <RadioButton
                                    className="radio-button"
                                    options={rbacTypeOptions}
                                    radioValue={rbacTypeWrite}
                                    fontFamily="InterSemiBold"
                                    style={{ marginRight: '20px', content: '' }}
                                    onChange={(e) => setRbacTypeWrite(e.target.value)}
                                    disabled={isDisabled}
                                />
                            </span>
                            <div className="form-section-row">
                                <Form.Item
                                    name="allow_write_permissions"
                                    rules={
                                        rbacTypeWrite === 'pattern' && [
                                            {
                                                pattern: /^[a-zA-Z0-9_\-., ]+(\..*)?|\*$/,
                                                message: `Only alphanumeric and the '_', '-', '.', '*' characters are allowed`
                                            }
                                        ]
                                    }
                                >
                                    {rbacTypeWrite === 'stations' ? (
                                        <Select
                                            suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                                            showArrow
                                            mode="multiple"
                                            style={{ width: '100%' }}
                                            options={stationsList || []}
                                            popupClassName="select-options"
                                            onChange={(e) => {
                                                updateFormState('allow_write_permissions', e);
                                                creationForm.setFieldsValue({ allow_write_permissions: e });
                                            }}
                                            disabled={isDisabled}
                                        />
                                    ) : (
                                        <Select
                                            suffixIcon={<ArrowDropDownRounded className="drop-down-icon" />}
                                            showArrow
                                            mode="tags"
                                            placeholder={'*'}
                                            value={selectedRow?.permissions?.allow_write_permissions || []}
                                            onChange={(e) => {
                                                updateFormState('allow_write_permissions', e);
                                                creationForm.setFieldsValue({ allow_write_permissions: e });
                                            }}
                                            style={{ width: '100%', backgroundColor: 'none' }}
                                            popupClassName="select-options"
                                            disabled={isDisabled}
                                        ></Select>
                                    )}
                                </Form.Item>
                            </div>
                        </div>
                    )}

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
                    onClick={() => (selectedRow ? closeModal() : onFinish())}
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
