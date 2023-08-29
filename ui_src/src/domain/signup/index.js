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

import React, { useState, useContext, useEffect, useCallback } from 'react';

import { KeyboardArrowRightRounded } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import { Form } from 'antd';

import {
    LOCAL_STORAGE_ACCOUNT_ID,
    LOCAL_STORAGE_INTERNAL_WS_PASS,
    LOCAL_STORAGE_CONNECTION_TOKEN,
    LOCAL_STORAGE_TOKEN,
    LOCAL_STORAGE_USER_PASS_BASED_AUTH
} from '../../const/localStorageConsts';
import { ReactComponent as FullLogo } from '../../assets/images/fullLogo.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { ReactComponent as SignupIcon } from '../../assets/images/signup.svg';
import { httpRequest } from '../../services/http';
import Switcher from '../../components/switcher';
import AuthService from '../../services/auth';
import Button from '../../components/button';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Input from '../../components/Input';
import Tooltip from '../../components/tooltip/tooltip';
import pathDomains from '../../router';
import { connect } from 'nats.ws';
import { ENVIRONMENT, WS_PREFIX, WS_SERVER_URL_PRODUCTION } from '../../config';

const Signup = (props) => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const [signupForm] = Form.useForm();
    const [formFields, setFormFields] = useState({
        username: '',
        full_name: '',
        password: '',
        subscription: true,
        user_type: 'management'
    });
    const [error, setError] = useState('');
    const [systemVersion, setSystemVersion] = useState('1.0.0');
    const [isLoading, setisLoading] = useState(true);

    const referer = props?.location?.state?.referer || '/overview';

    const handleEmailChange = (e) => {
        setFormFields({ ...formFields, username: e.target.value });
    };

    const handleFullNameChange = (e) => {
        setFormFields({ ...formFields, full_name: e.target.value });
    };

    const handlePasswordChange = (e) => {
        setFormFields({ ...formFields, password: e.target.value });
    };

    const switchSubscription = () => {
        setFormFields({ ...formFields, subscription: !formFields.subscription });
    };

    const [loadingSubmit, setLoadingSubmit] = useState(false);

    const getSignupFlag = useCallback(async () => {
        const data = await httpRequest('GET', ApiEndpoints.GET_SIGNUP_FLAG);
        if (!data.show_signup || state.skipSignup) {
            history.push(pathDomains.login);
        }
        setisLoading(false);
    }, []);

    const getSystemVersion = useCallback(async () => {
        const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_INFO);
        if (data) {
            setSystemVersion(data.version);
        }
        setisLoading(false);
    }, []);

    useEffect(() => {
        if (localStorage.getItem(LOCAL_STORAGE_TOKEN) && AuthService.isValidToken()) {
            history.push(referer);
        } else {
            setisLoading(true);
            getSignupFlag().catch(setisLoading(false));
            getSystemVersion().catch(setisLoading(false));
        }
    }, [getSignupFlag, getSystemVersion]);

    const handleSubmit = async (e) => {
        const values = await signupForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            try {
                setLoadingSubmit(true);
                const data = await httpRequest('POST', ApiEndpoints.SIGNUP, formFields, {}, {}, false);
                if (data) {
                    AuthService.saveToLocalStorage(data);
                    try {
                        const ws_port = data.ws_port;
                        const SOCKET_URL = ENVIRONMENT === 'production' ? `${WS_PREFIX}://${WS_SERVER_URL_PRODUCTION}:${ws_port}` : `${WS_PREFIX}://localhost:${ws_port}`;
                        let conn;
                        if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                            const account_id = localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID);
                            const internal_ws_pass = localStorage.getItem(LOCAL_STORAGE_INTERNAL_WS_PASS);
                            conn = await connect({
                                servers: [SOCKET_URL],
                                user: '$memphis_user$' + account_id,
                                pass: internal_ws_pass,
                                timeout: '5000'
                            });
                        } else {
                            const connection_token = localStorage.getItem(LOCAL_STORAGE_CONNECTION_TOKEN);
                            conn = await connect({
                                servers: [SOCKET_URL],
                                token: '::' + connection_token,
                                timeout: '5000'
                            });
                        }
                        dispatch({ type: 'SET_SOCKET_DETAILS', payload: conn });
                    } catch (error) {}
                    dispatch({ type: 'SET_USER_DATA', payload: data });
                    history.push(referer);
                }
            } catch (err) {
                setError(err);
            }
            setLoadingSubmit(false);
        }
    };

    return (
        <>
            {!isLoading && (
                <section className="signup-container">
                    {state.loading ? <Loader></Loader> : ''}
                    <SignupIcon alt="signup-icon" className="signup-icon" />
                    <div className="signup-form">
                        <FullLogo alt="logo" className="form-logo" />
                        <p className="signup-sub-title">Let’s create your first user</p>
                        <Form
                            className="form-fields"
                            name="basic"
                            initialValues={{
                                remember: true
                            }}
                            form={signupForm}
                        >
                            <Form.Item
                                name="username"
                                rules={[
                                    {
                                        required: true,
                                        message: 'Email can not be empty'
                                    },
                                    {
                                        type: 'email',
                                        message: 'Please insert a valid email'
                                    }
                                ]}
                            >
                                <div className="field name">
                                    <p>Your email</p>
                                    <Input
                                        placeholder="name@gmail.com"
                                        type="text"
                                        radiusType="semi-round"
                                        colorType="gray"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        width="470px"
                                        height="43px"
                                        minWidth="200px"
                                        onBlur={handleEmailChange}
                                        onChange={handleEmailChange}
                                        value={formFields.username}
                                    />
                                </div>
                            </Form.Item>
                            <Form.Item
                                name="full_name"
                                rules={[
                                    {
                                        required: true,
                                        message: 'Fullname can not be empty'
                                    }
                                ]}
                            >
                                <div className="field">
                                    <p>Full name</p>
                                    <Input
                                        placeholder="Type your name"
                                        type="text"
                                        radiusType="semi-round"
                                        colorType="gray"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        width="470px"
                                        height="43px"
                                        minWidth="200px"
                                        onBlur={handleFullNameChange}
                                        onChange={handleFullNameChange}
                                        value={formFields.full_name}
                                    />
                                </div>
                            </Form.Item>
                            <Form.Item
                                name="password"
                                rules={[
                                    {
                                        required: true,
                                        message: 'Password can not be empty'
                                    }
                                ]}
                            >
                                <div className="field password">
                                    <p>Password</p>
                                    <Input
                                        placeholder="Password"
                                        type="password"
                                        radiusType="semi-round"
                                        colorType="gray"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        width="470px"
                                        height="43px"
                                        minWidth="200px"
                                        onChange={handlePasswordChange}
                                        onBlur={handlePasswordChange}
                                        value={formFields.password}
                                    />
                                </div>
                            </Form.Item>
                            <p className="future-updates">Features and releases updates</p>
                            <div className="toggle-analytics">
                                <Form.Item name="subscription" initialValue={formFields.subscription} style={{ marginBottom: '0' }}>
                                    <Switcher onChange={() => switchSubscription()} checked={formFields.subscription} checkedChildren="" unCheckedChildren="" />
                                </Form.Item>
                                <label className="unselected-toggle">Receive features and release updates (You can unsubscribe at any time)</label>
                            </div>
                            {error && (
                                <div className="error-message">
                                    <p>For some reason we couldn’t process your signup, please reach to support</p>
                                </div>
                            )}
                            <Form.Item className="button-container">
                                <Button
                                    width="276px"
                                    height="43px"
                                    placeholder={
                                        <div className="placeholder-btn">
                                            <p>Continue</p> <KeyboardArrowRightRounded />
                                        </div>
                                    }
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="12px"
                                    fontWeight="600"
                                    isLoading={loadingSubmit}
                                    onClick={handleSubmit}
                                />
                            </Form.Item>
                        </Form>
                        <div className="version">
                            <p>v{systemVersion}</p>
                        </div>
                    </div>
                </section>
            )}
        </>
    );
};

export default Signup;
