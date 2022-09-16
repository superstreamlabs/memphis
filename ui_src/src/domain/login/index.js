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

import React, { useState, useContext, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import { Form } from 'antd';

import { LOCAL_STORAGE_TOKEN } from '../../const/localStorageConsts';
import betaFullLogo from '../../assets/images/betaFullLogo.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import sharps from '../../assets/images/sharps.svg';
import { httpRequest } from '../../services/http';
import AuthService from '../../services/auth';
import Button from '../../components/button';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Input from '../../components/Input';
import { SOCKET_URL } from '../../config';
import io from 'socket.io-client';
import pathDomains from '../../router';

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
            if (data.exist) {
                history.push(pathDomains.signup);
            }
            setIsSignup(data.exist);
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
                    const socket = await io.connect(SOCKET_URL, {
                        path: '/api/socket.io',
                        query: {
                            authorization: data.jwt
                        },
                        reconnection: false
                    });
                    dispatch({ type: 'SET_USER_DATA', payload: data });
                    dispatch({ type: 'SET_SOCKET_DETAILS', payload: socket });
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
            {!isLoading && !isSignup && (
                <section className="loginContainers">
                    {state.loading ? <Loader></Loader> : ''}
                    <div className="desktop-container">
                        <div className="desktop-content">
                            <div className="logoImg">
                                <img alt="logo" src={betaFullLogo}></img>
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
                                            <div id="e2e-login-username">
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
                                            <div id="e2e-login-password">
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
                                        <div id="e2e-login-btn">
                                            <Button
                                                width="19vw"
                                                height="43px"
                                                minWidth="200px"
                                                placeholder="Sign in"
                                                colorType="white"
                                                radiusType="circle"
                                                backgroundColorType="purple"
                                                fontSize="12px"
                                                fontWeight="600"
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
                            <img alt="sharps" src={sharps}></img>
                        </div>
                    </div>
                </section>
            )}
        </>
    );
};

export default Login;
