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

import React, { useState, useEffect, useContext } from 'react';
import Lottie from 'lottie-react';

import { httpRequest } from '../../../../services/http';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import Button from '../../../../components/button';
import CopyIcon from '../../../../assets/images/copy.svg';
import Information from '../../../../assets/images/information.svg';
import UserCheck from '../../../../assets/images/userCheck.svg';
import userCreator from '../../../../assets/lotties/userCreator.json';
import ClickableImage from '../../../../components/clickableImage';
import Input from '../../../../components/Input';
import { GetStartedStoreContext } from '..';
import SelectedClipboard from '../../../../assets/images/selectedClipboard.svg';
import TitleComponent from '../../../../components/titleComponent';

const screenEnum = {
    CREATE_USER_PAGE: 0,
    DATA_WAITING: 1,
    DATA_RECIEVED: 2
};

const CreateAppUser = (props) => {
    const { createStationFormRef } = props;
    const [selectedClipboardUserName, setSelectedClipboardUserName] = useState(false);
    const [selectedClipboardToken, setSelectedClipboardToken] = useState(false);
    const [isCreatedUser, setCreatedUser] = useState(screenEnum['CREATE_USER_PAGE']);
    const [getStartedState, getStartedDispatch] = useContext(GetStartedStoreContext);
    const [allowEdit, setAllowEdit] = useState(true);

    const [username, setUsername] = useState('');

    useEffect(() => {
        createStationFormRef.current = onNext;
        getStartedDispatch({ type: 'SET_NEXT_DISABLE', payload: true });
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

    const onCopy = async (copyParam) => {
        navigator.clipboard.writeText(copyParam);
    };

    const onCopyClick = async (copyValue, setImageState) => {
        onCopy(copyValue);
        setImageState(true);
        setTimeout(() => {
            setImageState(false);
        }, 3000);
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
                    <p className="create-the-user-header">User is getting created</p>
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
                            {selectedClipboardUserName ? (
                                <ClickableImage image={SelectedClipboard} className="copy-icon"></ClickableImage>
                            ) : (
                                <ClickableImage
                                    image={CopyIcon}
                                    alt="copyIcon"
                                    className="copy-icon"
                                    onClick={() => {
                                        onCopyClick(getStartedState.username, setSelectedClipboardUserName);
                                    }}
                                />
                            )}
                        </div>
                        <div className="token-container">
                            <p>Connection token: {getStartedState?.connectionCreds}</p>
                            {selectedClipboardToken ? (
                                <ClickableImage image={SelectedClipboard} className="copy-icon"></ClickableImage>
                            ) : (
                                <ClickableImage
                                    image={CopyIcon}
                                    alt="copyIcon"
                                    className="copy-icon"
                                    onClick={() => {
                                        onCopyClick(getStartedState.connectionCreds, setSelectedClipboardToken);
                                    }}
                                />
                            )}
                        </div>
                    </div>
                </div>
            )}
            {isCreatedUser === screenEnum['DATA_RECIEVED'] && (
                <div className="information-container">
                    <img src={Information} alt="information" className="information-img" />
                    <p className="information">Please note when you close this modal, you will not be able to restore your user details!!</p>
                </div>
            )}
        </div>
    );
};

export default CreateAppUser;
