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

import React, { useEffect, useState, useRef } from 'react';

import { ApiEndpoints } from '../../../../const/apiEndpoints';
import CustomSelect from '../../../../components/customSelect';
import { ReactComponent as RefreshIcon } from '../../../../assets/images/refresh.svg';
import { httpRequest } from '../../../../services/http';
import Button from '../../../../components/button';
import Input from '../../../../components/Input';
import Copy from '../../../../components/copy';
import CreateUserDetails from '../../../users/createUserDetails';
import { Drawer } from 'antd';
import { LOCAL_STORAGE_ACCOUNT_ID, LOCAL_STORAGE_USER_PASS_BASED_AUTH } from '../../../../const/localStorageConsts';
import { isCloud } from '../../../../services/valueConvertor';

const GenerateTokenModal = ({ host, close, returnToken, restProducer, stationName }) => {
    const [isLoading, setIsLoading] = useState(true);
    const [generateLoading, setGenerateLoading] = useState(false);
    const [appUsers, setAppUsers] = useState([]);
    const [formFields, setFormFields] = useState({
        username: appUsers[0]?.name || '',
        connection_token: '',
        account_id: isCloud() ? Number(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)) : 1,
        token_expiry_in_minutes: restProducer ? 525600 : 123,
        refresh_token_expiry_in_minutes: 10000092
    });
    const [userToken, setUserToken] = useState({});
    const [tokenTitle, setTokenTitle] = useState('Connection token');
    const [tokenPlaceHolder, setTokenPlaceHolder] = useState('Generated during user creation');
    const [addUserModalFlip, setAddUserModalFlip] = useState(false);
    const [createUserLoader, setCreateUserLoader] = useState(false);
    const createUserRef = useRef(null);

    const updateState = (field, value) => {
        let updatedValue = { ...formFields };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, ...updatedValue }));
    };

    const getAppUsers = async () => {
        setIsLoading(true);
        try {
            let data = await httpRequest('GET', ApiEndpoints.GET_APP_USERS);
            if (data) {
                let newObjectArray = data.map(({ username: name }) => name);
                updateState('username', newObjectArray[0]);
                setAppUsers(newObjectArray);
                setIsLoading(false);
            }
        } catch (erro) {
            setIsLoading(false);
        }
    };

    const handleAddUser = async (userData) => {
        await getAppUsers();
        setCreateUserLoader(false);
        setAddUserModalFlip(false);
        updateState('username', userData.username);
    };

    useEffect(() => {
        getAppUsers();
        if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
            setFormFields({
                username: appUsers[0]?.name || '',
                password: '',
                account_id: isCloud() ? Number(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)) : 1,
                token_expiry_in_minutes: restProducer ? 525600 : 123,
                refresh_token_expiry_in_minutes: 10000092
            });
            setTokenTitle('Password');
            setTokenPlaceHolder('User password');
        }
        return () => {};
    }, []);

    const generateToken = async () => {
        setGenerateLoading(true);
        try {
            let data = await httpRequest('POST', ApiEndpoints.GENERATE_TOKEN, { ...formFields }, {}, {}, false, 0, host);
            if (data) {
                setUserToken({
                    jwt: data.jwt,
                    jwt_refresh_token: data.jwt_refresh_token
                });
                setGenerateLoading(false);
                returnToken(data);
            }
        } catch (erro) {
            setGenerateLoading(false);
        }
    };

    return (
        <div className="generate-modal-wrapper">
            {!isLoading && (
                <>
                    {restProducer ? (
                        <p className="desc">
                            Produce and consume data using the REST protocol. Great for webhooks ingestion and connecting your Memphis with other platforms
                        </p>
                    ) : (
                        <p className="desc">
                            JWT token can be generated using a REST call, but for better convenience, it can also be generated through the GUI.
                            <br />
                            <br />
                            By default, tokens are generated with 15-minutes expiration time for security purposes and can be refreshed using the "refresh token"
                        </p>
                    )}
                    {((Object.keys(userToken).length === 0 && !restProducer) || restProducer) && (
                        <>
                            <div className="user-password-section">
                                <div className="app-username">
                                    <p className="field-title">Client-type user</p>
                                    <CustomSelect
                                        placeholder="choose your app user"
                                        value={formFields?.username || appUsers[0]}
                                        options={appUsers}
                                        onChange={(e) => updateState('username', e)}
                                        type="user"
                                        handleCreateNew={() => setAddUserModalFlip(true)}
                                    />
                                </div>
                                <div className="app-token">
                                    <p className="field-title">{tokenTitle}</p>
                                    <Input
                                        placeholder={tokenPlaceHolder}
                                        type="password"
                                        fontSize="12px"
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        height="40px"
                                        onBlur={(e) => {
                                            if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                                                updateState('password', e.target.value);
                                            } else {
                                                updateState('connection_token', e.target.value);
                                            }
                                        }}
                                        onChange={(e) => {
                                            if (localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true') {
                                                updateState('password', e.target.value);
                                            } else {
                                                updateState('connection_token', e.target.value);
                                            }
                                        }}
                                        value={localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'true' ? formFields.password : formFields.connection_token}
                                    />
                                </div>
                            </div>
                        </>
                    )}
                    {Object.keys(userToken).length !== 0 && !restProducer && (
                        <>
                            <div className="api-token">
                                <p className="field-title">JWT token</p>
                                <div className="input-and-copy">
                                    <Input
                                        width="98%"
                                        numberOfRows="5"
                                        type="textArea"
                                        fontSize="12px"
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        disabled={true}
                                        value={userToken?.jwt}
                                    />
                                    <Copy data={userToken?.jwt} width={20} />
                                </div>
                            </div>

                            <div className="api-token">
                                <p className="field-title">JWT refresh token</p>
                                <div className="input-and-copy">
                                    <Input
                                        width="98%"
                                        numberOfRows="5"
                                        type="textArea"
                                        fontSize="12px"
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="none"
                                        borderColorType="gray"
                                        disabled={true}
                                        value={userToken?.jwt_refresh_token}
                                    />
                                    <Copy data={userToken?.jwt_refresh_token} width={20} />
                                </div>
                                <div className="generate-again" onClick={() => setUserToken({})}>
                                    <RefreshIcon width={14} />
                                    <span>Generate again</span>
                                </div>
                            </div>
                        </>
                    )}

                    {Object.keys(userToken).length !== 0 && restProducer && (
                        <div className="user-password-section">
                            <div className="api-token">
                                <p className="field-title">Produce</p>
                                <div className="input-and-copy">
                                    <span className="url-span">{`https://${host}.restgw.cloud.memphis.dev/stations/${stationName}/produce/single?authorization=${userToken?.jwt_refresh_token}`}</span>
                                    <Copy data={userToken?.jwt_refresh_token} width={20} />
                                </div>
                            </div>
                            <div className="api-token">
                                <p className="field-title">Consume</p>
                                <div className="input-and-copy">
                                    <span className="url-span">{`https://${host}.restgw.cloud.memphis.dev/stations/${stationName}/consume/batch?authorization=${userToken?.jwt_refresh_token}`}</span>
                                    <Copy data={userToken?.jwt_refresh_token} width={20} />
                                </div>
                            </div>
                        </div>
                    )}

                    {(Object.keys(userToken).length === 0 && !restProducer) || restProducer ? (
                        <Button
                            width="100%"
                            height="36px"
                            placeholder={Object.keys(userToken).length !== 0 && restProducer ? 'Regenerate' : 'Generate'}
                            colorType="white"
                            radiusType="semi-round"
                            backgroundColorType={'purple'}
                            fontSize="14px"
                            fontWeight="bold"
                            disabled={formFields.connection_token === '' && formFields.password === ''}
                            isLoading={generateLoading}
                            onClick={generateToken}
                        />
                    ) : (
                        <Button
                            width="100%"
                            height="36px"
                            placeholder="Close"
                            colorType="white"
                            radiusType="semi-round"
                            backgroundColorType={'purple'}
                            fontSize="14px"
                            fontWeight="bold"
                            disabled={!userToken}
                            isLoading={generateLoading}
                            onClick={close}
                        />
                    )}
                </>
            )}
            <Drawer
                placement="right"
                title="Add a new user"
                onClose={() => {
                    setCreateUserLoader(false);
                    setAddUserModalFlip(false);
                }}
                destroyOnClose={true}
                width="650px"
                open={addUserModalFlip}
            >
                <CreateUserDetails
                    clientType
                    createUserRef={createUserRef}
                    closeModal={(userData) => handleAddUser(userData)}
                    handleLoader={(e) => setCreateUserLoader(e)}
                    isLoading={createUserLoader}
                />
            </Drawer>
        </div>
    );
};

export default GenerateTokenModal;
