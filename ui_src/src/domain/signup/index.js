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

import React, { useState, useContext, useEffect, useCallback } from 'react';
import { KeyboardArrowRightRounded } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import io from 'socket.io-client';
import { Form } from 'antd';

import { LOCAL_STORAGE_TOKEN } from '../../const/localStorageConsts';
import betaFullLogo from '../../assets/images/betaFullLogo.svg';
import betaBadge from '../../assets/images/betaBadge.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import signup from '../../assets/images/signup.svg';
import { httpRequest } from '../../services/http';
import Switcher from '../../components/switcher';
import AuthService from '../../services/auth';
import Button from '../../components/button';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Input from '../../components/Input';
import { SOCKET_URL } from '../../config';
import pathDomains from '../../router';

const Signup = (props) => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const [signupForm] = Form.useForm(); // form controller
    const [formFields, setFormFields] = useState({
        username: '',
        full_name: '',
        password: '',
        subscription: true,
        user_type: 'management'
    });
    const [error, setError] = useState('');
    const [systemVersion, setSystemVersion] = useState('');
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
        if (!data.exist) {
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
            {!isLoading && (
                <section className="signup-container">
                    {state.loading ? <Loader></Loader> : ''}
                    <img alt="signup-img" className="signup-img" src={signup}></img>
                    <div className="signup-form">
                        <img alt="logo" className="form-logo" src={betaFullLogo}></img>
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
                                <div className="field name" id="e2e-tests-field-email">
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
                                <div className="field" id="e2e-tests-field-fullname">
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
                                    <div id="e2e-tests-field-password">
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
                                <div id="e2e-tests-signup-btn">
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
                                </div>
                            </Form.Item>
                        </Form>
                        <div className="version">
                            <p>v0.3.5</p>
                            <img src={betaBadge} />
                        </div>
                    </div>
                </section>
            )}
        </>
    );
};

export default Signup;
