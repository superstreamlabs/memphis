// Credit for The NATS.IO Authors
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

import React, { useState, useEffect, useContext } from 'react';
import Lottie from 'lottie-react';

import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import Button from '../../../../components/button';
import Information from '../../../../assets/images/information.svg';
import UserCheck from '../../../../assets/images/userCheck.svg';
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

    const [username, setUsername] = useState('');

    useEffect(() => {
        createStationFormRef.current = onNext;
        getStartedDispatch({ type: 'SET_CREATE_APP_USER_DISABLE', payload: false });
        if (getStartedState?.username) {
            getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: false });
            setUsername(getStartedState.username);
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
            username: username,
            user_type: 'application'
        };
        try {
            const data = await httpRequest('POST', ApiEndpoints.ADD_USER, bodyRequest);
            setCreatedUser(screenEnum['DATA_WAITING']);

            if (data) {
                setAllowEdit(false);
                getStartedDispatch({ type: 'IS_LOADING', payload: false });

                getStartedDispatch({ type: 'SET_USER_NAME', payload: data?.username });
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
        <div className="create-station-form-create-app-user" id="e2e-getstarted-step2">
            <div>
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
                    onBlur={(e) => setUsername(e.target.value)}
                    onChange={(e) => setUsername(e.target.value)}
                    value={username}
                    disabled={!allowEdit}
                />
                <Button
                    placeholder="Create app user"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontWeight="bold"
                    marginTop="25px"
                    disabled={!allowEdit || username.length === 0}
                    onClick={handleCreateUser}
                    isLoading={getStartedState?.isLoading}
                />
            </div>
            {isCreatedUser === screenEnum['DATA_WAITING'] && (
                <div className="creating-the-user-container">
                    <Lottie className="lottie" animationData={userCreator} loop={true} />
                    <p className="create-the-user-header">We are creating the user</p>
                </div>
            )}
            {isCreatedUser === screenEnum['DATA_RECIEVED'] && (
                <div className="connection-details-container">
                    <div className="user-details-container">
                        <img src={UserCheck} alt="usercheck" width="20px" height="20px"></img>
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
                    <img src={Information} alt="information" className="information-img" />
                    <p className="information">Please save the generated credentials for future usage.</p>
                </div>
            )}
        </div>
    );
};

export default CreateAppUser;
