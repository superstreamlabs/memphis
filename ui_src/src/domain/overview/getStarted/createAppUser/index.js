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

import React, { useState, useEffect, useContext } from 'react';
import Lottie from 'lottie-react';
import { LOCAL_STORAGE_USER_PASS_BASED_AUTH } from '../../../../const/localStorageConsts';
import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import Button from '../../../../components/button';
import { ReactComponent as InformationIcon } from '../../../../assets/images/information.svg';
import { ReactComponent as UserCheckIcon } from '../../../../assets/images/userCheck.svg';
import userCreator from '../../../../assets/lotties/userCreator.json';
import Input from '../../../../components/Input';
import { GetStartedStoreContext } from '..';
import TitleComponent from '../../../../components/titleComponent';
import Copy from '../../../../components/copy';

const screenEnum = {
    CREATE_USER_PAGE: 0,
    DATA_WAITING: 1,
    DATA_RECIEVED: 2
};

const CreateAppUser = (props) => {
    const { createStationFormRef } = props;
    const [isCreatedUser, setCreatedUser] = useState(screenEnum['CREATE_USER_PAGE']);
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [allowEdit, setAllowEdit] = useState(true);
    const [user, setUser] = useState({
        username: '',
        password: ''
    });
    const [confirmPass, setConfirmPass] = useState('');

    useEffect(() => {
        createStationFormRef.current = onNext;
        getStartedDispatch({ type: 'SET_CREATE_APP_USER_DISABLE', payload: false });
        if (getStartedState?.username !== '') {
            getStartedDispatch({ type: 'SET_CREATE_APP_USER_DISABLE', payload: true });
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
            setUser({ username: getStartedState.username, password: getStartedState.password });
            setAllowEdit(false);
        }
    }, []);

    const onNext = () => {
        getStartedDispatch({ type: 'SET_COMPLETED_STEPS', payload: getStartedState?.currentStep });
        getStartedDispatch({ type: 'SET_CURRENT_STEP', payload: getStartedState?.currentStep + 1 });
    };

    const handleCreateUser = async () => {
        getStartedDispatch({ type: 'IS_LOADING', payload: true });
        const bodyRequest = {
            username: user.username,
            password: user.password,
            user_type: 'application'
        };
        try {
            const data = await httpRequest('POST', ApiEndpoints.ADD_USER, bodyRequest);
            setCreatedUser(screenEnum['DATA_WAITING']);

            if (data) {
                setAllowEdit(false);
                getStartedDispatch({ type: 'IS_LOADING', payload: false });

                getStartedDispatch({ type: 'SET_USER', payload: { username: data?.username, password: user.password } });
                getStartedDispatch({ type: 'SET_BROKER_CONNECTION_CREDS', payload: data?.broker_connection_creds });

                getStartedDispatch({ type: 'SET_CREATE_APP_USER_DISABLE', payload: true });
                getStartedDispatch({ type: 'IS_APP_USER_CREATED', payload: true });
                setTimeout(() => {
                    setCreatedUser(screenEnum['DATA_RECIEVED']);
                    getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
                }, 2000);
            }
        } catch (error) {
            getStartedDispatch({ type: 'IS_LOADING', payload: false });
            setCreatedUser(screenEnum['CREATE_USER_PAGE']);
        }
    };

    return (
        <div className="create-station-form-create-app-user">
            <div>
                <div className="create-user-form-field">
                    <TitleComponent headerTitle="Enter user name" typeTitle="sub-header" required={true}></TitleComponent>
                    <Input
                        placeholder="Type user name"
                        type="text"
                        radiusType="semi-round"
                        colorType="black"
                        backgroundColorType="none"
                        borderColorType="gray"
                        width="371px"
                        height="38px"
                        onBlur={(e) => setUser({ ...user, username: e.target.value })}
                        onChange={(e) => setUser({ ...user, username: e.target.value })}
                        value={user.username}
                        disabled={!allowEdit}
                    />
                </div>
                {localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true' && allowEdit && (
                    <div>
                        <div className="create-user-form-field">
                            <TitleComponent headerTitle="Enter password" typeTitle="sub-header" required={true}></TitleComponent>
                            <span className="password-hint">
                                Hint: Password must be at least 8 characters long, contain both uppercase and lowercase, and at least one number and one special
                                character(!?\-@#$%).
                            </span>
                            <Input
                                placeholder="Type password"
                                type="password"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                width="371px"
                                height="38px"
                                onBlur={(e) => setUser({ ...user, password: e.target.value })}
                                onChange={(e) => setUser({ ...user, password: e.target.value })}
                                value={user.password}
                                disabled={!allowEdit}
                            />
                        </div>
                        <div className="create-user-form-field">
                            <TitleComponent headerTitle="Confirm password" typeTitle="sub-header" required={true}></TitleComponent>
                            <Input
                                placeholder="Confirm password"
                                type="password"
                                radiusType="semi-round"
                                colorType="black"
                                backgroundColorType="none"
                                borderColorType="gray"
                                width="371px"
                                height="38px"
                                onBlur={(e) => setConfirmPass(e.target.value)}
                                onChange={(e) => setConfirmPass(e.target.value)}
                                value={confirmPass}
                                disabled={!allowEdit}
                            />
                            {confirmPass != '' && user.password !== confirmPass && <label className="validate-pass">Passwords do not match</label>}
                        </div>
                    </div>
                )}
                <Button
                    placeholder="Create app user"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontWeight="bold"
                    marginTop="25px"
                    disabled={
                        !allowEdit ||
                        user.username?.length === 0 ||
                        (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true' && user.password === '') ||
                        (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true' && user.password !== confirmPass)
                    }
                    onClick={handleCreateUser}
                    isLoading={getStartedState?.isLoading}
                />
            </div>
            {isCreatedUser === screenEnum['DATA_WAITING'] && (
                <div className="creating-the-user-container">
                    <Lottie className="lottie" animationData={userCreator} loop={true} />
                    <p className="create-the-user-header">Please hold a second. The user is being created</p>
                </div>
            )}
            {isCreatedUser === screenEnum['DATA_RECIEVED'] && localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) !== 'true' && (
                <div className="connection-details-container">
                    <div className="user-details-container">
                        <UserCheckIcon alt="usercheck" width={20} height={20} />
                        <p className="user-connection-details">User connection details</p>
                    </div>
                    <div className="container-username-token">
                        <div className="username-container">
                            <p>Username: {getStartedState.username}</p>
                            <Copy data={getStartedState.username} />
                        </div>
                        <div className="token-container">
                            <p>Connection token: {getStartedState?.connectionCreds}</p>
                            <Copy data={getStartedState.connectionCreds} />
                        </div>
                    </div>
                </div>
            )}
            {isCreatedUser === screenEnum['DATA_RECIEVED'] && (
                <div className="information-container">
                    <InformationIcon alt="information" className="information-img" />
                    {localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true' ? (
                        <p className="information">Please save your credentials for future usage.</p>
                    ) : (
                        <p className="information">Please save the generated credentials for future usage.</p>
                    )}
                </div>
            )}
        </div>
    );
};

export default CreateAppUser;
