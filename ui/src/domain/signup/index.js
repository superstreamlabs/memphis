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
// limitations under the License.

import './style.scss';

import React, { useState, useContext, useEffect } from 'react';
import { useHistory } from 'react-router-dom';
import { Form } from 'antd';

import { LOCAL_STORAGE_TOKEN } from '../../const/localStorageConsts';
import betaFullLogo from '../../assets/images/betaFullLogo.svg';
import signup from '../../assets/images/signup.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import sharps from '../../assets/images/sharps.svg';
import { httpRequest } from '../../services/http';
import AuthService from '../../services/auth';
import Button from '../../components/button';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Input from '../../components/Input';
import Switcher from '../../components/switcher';
import { SOCKET_URL } from '../../config';
import io from 'socket.io-client';

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

    useEffect(() => {
        if (localStorage.getItem(LOCAL_STORAGE_TOKEN) && AuthService.isValidToken()) {
            history.push(referer);
        }
    }, []);

    const handleSubmit = async (e) => {
        e.preventDefault();
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
            <section className="signup-container">
                {state.loading ? <Loader></Loader> : ''}
                <img alt="sharps" className="signup-img" src={signup}></img>
                <div className="signup-form">
                    <img alt="logo" className="form-logo" src={betaFullLogo}></img>
                    <p className="signup-sub-title">Let’s get started with memphis</p>
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
                                    placeholder="name@company.com"
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
                                <div id="e2e-tests-password">
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
                            <label className="unselected-toggle">Receive features and releases updates (You can unsubscribe any time)</label>
                        </div>
                        <Form.Item className="button-container">
                            <Button
                                width="470px"
                                height="43px"
                                minWidth="200px"
                                placeholder="Continue"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontWeight="600"
                                isLoading={loadingSubmit}
                                onClick={handleSubmit}
                            />
                        </Form.Item>
                        {error && (
                            <div className="error-message">
                                <p>For some reason we couldn’t process your signup, please reach to support</p>
                            </div>
                        )}
                    </Form>
                </div>
            </section>
        </>
    );
};

export default Signup;
