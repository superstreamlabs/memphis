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

import React, { useState, useContext, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import { Form } from 'antd';

import {
    LOCAL_STORAGE_ACCOUNT_ID,
    LOCAL_STORAGE_WS_HOST,
    LOCAL_STORAGE_INTERNAL_WS_PASS,
    LOCAL_STORAGE_CONNECTION_TOKEN,
    LOCAL_STORAGE_TOKEN,
    LOCAL_STORAGE_USER_PASS_BASED_AUTH
} from '../../const/localStorageConsts';
import { ReactComponent as FullLogo } from '../../assets/images/fullLogo.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { ReactComponent as SharpsIcon } from '../../assets/images/sharps.svg';
import { httpRequest } from '../../services/http';
import AuthService from '../../services/auth';
import Button from '../../components/button';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Input from '../../components/Input';
import pathDomains from '../../router';
import { connect } from 'nats.ws';
import { WS_PREFIX } from '../../config';

const Login = (props) => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const [loginForm] = Form.useForm(); // form controller
    const [formFields, setFormFields] = useState({
        username: '',
        password: ''
    });
    const [error, setError] = useState('');
    const referer = props?.location?.state?.referer || '/overview';
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [isLoading, setisLoading] = useState(false);
    const [isSignup, setIsSignup] = useState(false);

    useEffect(() => {
        if (localStorage.getItem(LOCAL_STORAGE_TOKEN) && AuthService.isValidToken()) {
            history.push(referer);
        } else {
            getSignupFlag();
        }
    }, []);

    const getSignupFlag = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_SIGNUP_FLAG);
            if (data.show_signup && !state.skipSignup) {
                history.push(pathDomains.signup);
            }
            setIsSignup();
            setisLoading(false);
        } catch (error) {
            setisLoading(false);
        }
    };

    const handleUserNameChange = (e) => {
        setFormFields({ ...formFields, username: e.target.value });
    };

    const handlePasswordChange = (e) => {
        setFormFields({ ...formFields, password: e.target.value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        const values = await loginForm.validateFields();
        if (values?.errorFields) {
            return;
        } else {
            try {
                setLoadingSubmit(true);
                const { username, password } = formFields;
                const data = await httpRequest('POST', ApiEndpoints.LOGIN, { username, password }, {}, {}, false);
                if (data) {
                    AuthService.saveToLocalStorage(data);
                    try {
                        let wsHost = localStorage.getItem(LOCAL_STORAGE_WS_HOST);
                        wsHost = `${WS_PREFIX}://${wsHost}`;
                        let conn;
                        if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                            const account_id = localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID);
                            const internal_ws_pass = localStorage.getItem(LOCAL_STORAGE_INTERNAL_WS_PASS);
                            conn = await connect({
                                servers: [wsHost],
                                user: '$memphis_user$' + account_id,
                                pass: internal_ws_pass,
                                timeout: '5000'
                            });
                        } else {
                            const connection_token = localStorage.getItem(LOCAL_STORAGE_CONNECTION_TOKEN);
                            conn = await connect({
                                servers: [wsHost],
                                token: '::' + connection_token,
                                timeout: '5000'
                            });
                        }
                        dispatch({ type: 'SET_SOCKET_DETAILS', payload: conn });
                    } catch (error) {
                        throw new Error(error);
                    }
                    dispatch({ type: 'SET_USER_DATA', payload: data });
                    history.push(referer);
                }
            } catch (err) {
                setError(err);
                setLoadingSubmit(false);
                console.log(err);
            }
            setLoadingSubmit(false);
        }
    };

    return (
        <>
            {!isLoading && !isSignup && (
                <section className="loginContainers">
                    {state.loading ? <Loader></Loader> : ''}
                    <div className="desktop-container">
                        <div className="desktop-content">
                            <div className="logoImg">
                                <FullLogo alt="logo" />
                            </div>
                            <div className="title">
                                <p>Hey Memphiser,</p>
                                <p>Welcome</p>
                            </div>
                            <div className="login-form">
                                <Form
                                    name="basic"
                                    initialValues={{
                                        remember: true
                                    }}
                                    form={loginForm}
                                >
                                    <Form.Item
                                        name="username"
                                        rules={[
                                            {
                                                required: true,
                                                message: 'Username can not be empty'
                                            }
                                        ]}
                                    >
                                        <div className="field name">
                                            <p>Username / Email</p>
                                            <div>
                                                <Input
                                                    placeholder="Type username / email"
                                                    type="text"
                                                    radiusType="semi-round"
                                                    colorType="gray"
                                                    backgroundColorType="none"
                                                    borderColorType="gray"
                                                    width="19vw"
                                                    height="43px"
                                                    minWidth="200px"
                                                    onBlur={handleUserNameChange}
                                                    onChange={handleUserNameChange}
                                                    value={formFields.username}
                                                />
                                            </div>
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
                                            <div>
                                                <Input
                                                    placeholder="Password"
                                                    type="password"
                                                    radiusType="semi-round"
                                                    colorType="gray"
                                                    backgroundColorType="none"
                                                    borderColorType="gray"
                                                    width="19vw"
                                                    height="43px"
                                                    minWidth="200px"
                                                    onChange={handlePasswordChange}
                                                    onBlur={handlePasswordChange}
                                                    value={formFields.password}
                                                />
                                            </div>
                                        </div>
                                    </Form.Item>
                                    <Form.Item className="button-container">
                                        <div>
                                            <Button
                                                width="19vw"
                                                height="43px"
                                                minWidth="200px"
                                                placeholder="Sign in"
                                                colorType="white"
                                                radiusType="circle"
                                                backgroundColorType="purple"
                                                fontSize="14px"
                                                fontFamily="InterBold"
                                                isLoading={loadingSubmit}
                                                onClick={handleSubmit}
                                            />
                                        </div>
                                    </Form.Item>

                                    {error && (
                                        <div className="error-message">
                                            <p>The username and password you entered did not match our records. Please double-check and try again.</p>
                                        </div>
                                    )}
                                </Form>
                            </div>
                        </div>
                        <div className="brand-shapes">
                            <SharpsIcon alt="sharps" />
                        </div>
                    </div>
                </section>
            )}
        </>
    );
};

export default Login;
