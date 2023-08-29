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

import React, { useState, useContext, useEffect } from 'react';
import { Form, message, Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

import { ReactComponent as TickCircleIcon } from '../../../../../assets/images/tickCircle.svg';
import { FiPlus } from 'react-icons/fi';
import { INTEGRATION_LIST } from '../../../../../const/integrationList';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Button from '../../../../../components/button';
import { Context } from '../../../../../hooks/store';
import Input from '../../../../../components/Input';
import Loader from '../../../../../components/loader';
import IntegrationItem from './integratedItem';

const GitHubIntegration = ({ close, value }) => {
    const isValue = value && Object?.keys(value)?.length !== 0;
    const githubConfiguration = INTEGRATION_LIST['GitHub'];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [formFields, setFormFields] = useState({
        name: 'github',
        keys: {
            token: '',
            connected_repos: []
        }
    });
    const [loadingCreate, setLoadingCreate] = useState(false);
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);
    const [loadingRepos, setLoadingRepos] = useState(true);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [repos, setRepos] = useState([]);
    const [addNew, setAddNew] = useState(false);
    const [isChanged, setIsChanged] = useState(false);
    const [isIntegrated, setIsIntagrated] = useState(false);

    useEffect(() => {
        value && Object.keys(value).length > 0 && setIsIntagrated(true);
        const images = [];
        images.push(INTEGRATION_LIST['GitHub'].banner.props.src);
        images.push(INTEGRATION_LIST['GitHub'].insideBanner.props.src);
        images.push(INTEGRATION_LIST['GitHub'].icon.props.src);
        const promises = [];

        images.forEach((imageUrl) => {
            const image = new Image();
            promises.push(
                new Promise((resolve) => {
                    image.onload = resolve;
                })
            );
            image.src = imageUrl;
        });

        Promise.all(promises).then(() => {
            setImagesLoaded(true);
        });
        getIntegration();
    }, []);

    function areEqual(arr1, arr2) {
        if (arr1?.length !== arr2?.length) {
            return false;
        }

        for (let i = 0; i < arr1?.length; i++) {
            const obj1 = arr1[i];
            const obj2 = arr2[i];

            if (obj1?.repo_name !== obj2?.repo_name || obj1?.repo_owner !== obj2?.repo_owner || obj1?.branch !== obj2?.branch) {
                return false;
            }
        }

        return true;
    }

    useEffect(() => {
        const results = areEqual(formFields?.keys?.connected_repos, value?.keys?.connected_repos);
        setIsChanged(value ? !results : true);
    }, [formFields]);

    const updateKeysConnectedRepos = (repo, index) => {
        let updatedValue = { ...formFields.keys };
        if (index < updatedValue.connected_repos?.length) updatedValue.connected_repos[index] = repo;
        else if (index === updatedValue.connected_repos?.length) {
            updatedValue.connected_repos.push(repo);
        }
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
        setAddNew(false);
    };

    const cleanEmptyFields = () => {
        let updatedValue = { ...formFields.keys };
        updatedValue.connected_repos = updatedValue.connected_repos.filter((repo) => repo.branch !== '');
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
        return { name: 'github', keys: updatedValue };
    };

    const removeRepoItem = (index) => {
        let updatedValue = { ...formFields.keys };
        updatedValue.connected_repos.splice(index, 1);
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    };

    const updateKeysState = (field, value) => {
        let updatedValue = { ...formFields.keys };
        updatedValue[field] = value;
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    };

    const closeModal = (data, disconnect = false) => {
        setTimeout(() => {
            disconnect ? setLoadingDisconnect(false) : setLoadingSubmit(false);
        }, 1000);
        close(data);
        message.success({
            key: 'memphisSuccessMessage',
            content: disconnect ? 'The integration was successfully disconnected' : 'The integration connected successfully',
            duration: 5,
            style: { cursor: 'pointer' },
            onClick: () => message.destroy('memphisSuccessMessage')
        });
    };

    const updateIntegration = async () => {
        setLoadingSubmit(true);
        const updatedFields = cleanEmptyFields();
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATION, updatedFields);
            dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
            closeModal(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const createIntegration = async () => {
        try {
            setLoadingCreate(true);
            const data = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, formFields);
            dispatch({ type: 'ADD_INTEGRATION', payload: data });
            getIntegration();
            setLoadingCreate(false);
            setIsIntagrated(true);
        } catch (err) {
            setLoadingCreate(false);
        }
    };

    const getIntegration = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_DETAILS}?name=github`);
            if (data) {
                updateKeysState('connected_repos', data?.integaraion?.keys?.connected_repos || []);
                setRepos(data?.repos);
            } else
                setFormFields({
                    name: 'github',
                    keys: {
                        token: '',
                        connected_repos: []
                    }
                });
            setLoadingRepos(false);
        } catch (error) {
            setLoadingRepos(false);
        }
    };

    const disconnect = async () => {
        setLoadingDisconnect(true);
        try {
            await httpRequest('DELETE', ApiEndpoints.DISCONNECT_INTEGRATION, {
                name: formFields.name
            });
            dispatch({ type: 'REMOVE_INTEGRATION', payload: formFields.name });

            closeModal({}, true);
        } catch (err) {
            setLoadingDisconnect(false);
        }
    };

    const antIcon = (
        <LoadingOutlined
            style={{
                fontSize: 24,
                color: '#5A4FE5'
            }}
            spin
        />
    );

    return (
        <dynamic-integration is="3xd" className="integration-modal-container">
            {!imagesLoaded && (
                <div className="loader-integration-box">
                    <Loader />
                </div>
            )}
            {imagesLoaded && (
                <>
                    {githubConfiguration?.insideBanner}
                    <div className="integrate-header">
                        {githubConfiguration.header}
                        <div className={!isIntegrated ? 'action-buttons flex-end' : 'action-buttons'}>
                            {isIntegrated && (
                                <Button
                                    width="100px"
                                    height="35px"
                                    placeholder="Disconnect"
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="red"
                                    border="none"
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    isLoading={loadingDisconnect}
                                    onClick={() => disconnect()}
                                />
                            )}
                            <Button
                                width="140px"
                                height="35px"
                                placeholder="Integration guide"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                border="none"
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                onClick={() => window.open('https://docs.memphis.dev/memphis/integrations-center/source-code/github', '_blank')}
                            />
                        </div>
                    </div>
                    {githubConfiguration.integrateDesc}
                    <Form name="form" form={creationForm} autoComplete="off" className="integration-form">
                        <div className="api-details">
                            <p className="title">API Details</p>
                            <div className="api-key">
                                <span className="connect-bth-gh">
                                    <p>API Token</p>
                                    {isIntegrated ? (
                                        <div className="connected-to-gh">
                                            <TickCircleIcon className="connected" alt="connected" />
                                            &nbsp;Connected
                                        </div>
                                    ) : (
                                        <Button
                                            width="80px"
                                            height="22px"
                                            placeholder={'Connect'}
                                            colorType="white"
                                            radiusType="circle"
                                            backgroundColorType="purple"
                                            fontSize="12px"
                                            fontFamily="InterSemiBold"
                                            disabled={!formFields.keys?.token}
                                            onClick={createIntegration}
                                            isLoading={loadingCreate}
                                        />
                                    )}
                                </span>

                                <span className="desc">The secret key associated with the access key.</span>
                                <Form.Item
                                    name="token"
                                    rules={[
                                        {
                                            required: !value?.keys?.token,
                                            message: 'Please insert a token.'
                                        }
                                    ]}
                                    initialValue={formFields?.keys?.token}
                                >
                                    <Input
                                        placeholder={value?.keys?.token || 'ghp_****'}
                                        type="text"
                                        radiusType="semi-round"
                                        colorType="black"
                                        backgroundColorType="purple"
                                        borderColorType="none"
                                        height="40px"
                                        fontSize="12px"
                                        onBlur={(e) => updateKeysState('token', e.target.value)}
                                        onChange={(e) => updateKeysState('token', e.target.value)}
                                        value={formFields?.keys?.token}
                                    />
                                </Form.Item>
                            </div>
                            {isIntegrated && (
                                <div className="input-field">
                                    <p className="title">Repos</p>
                                    <div className="repos-container">
                                        <div className="repos-header">
                                            <label></label>
                                            <label>REPO NAME</label>
                                            <label>BRANCH</label>
                                        </div>
                                        <div className="repos-body">
                                            {loadingRepos && (
                                                <div className="repos-loader">
                                                    <Spin indicator={antIcon} />
                                                </div>
                                            )}
                                            {formFields?.keys?.connected_repos?.map((repo, index) => {
                                                return (
                                                    <IntegrationItem
                                                        key={index}
                                                        index={index}
                                                        repo={repo}
                                                        reposList={repos || []}
                                                        updateIntegrationList={(updatedFields, i) => updateKeysConnectedRepos(updatedFields, i)}
                                                        removeRepo={(i) => {
                                                            removeRepoItem(i);
                                                        }}
                                                    />
                                                );
                                            })}
                                            {addNew ? (
                                                <IntegrationItem
                                                    index={formFields?.keys?.connected_repos?.length}
                                                    repo={''}
                                                    reposList={repos || []}
                                                    updateIntegrationList={(updatedFields, i) => updateKeysConnectedRepos(updatedFields, i)}
                                                    removeRepo={(i) => {
                                                        removeRepoItem(i);
                                                        setAddNew(false);
                                                    }}
                                                />
                                            ) : (
                                                <div className="add-more-repos" onClick={() => setAddNew(!addNew)}>
                                                    <FiPlus /> <label> {formFields?.keys?.connected_repos?.length === 0 ? `Add the first repo` : `Add more repos`}</label>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                        <Form.Item className="button-container">
                            <div className="button-wrapper">
                                <Button
                                    width="80%"
                                    height="45px"
                                    placeholder="Close"
                                    colorType="black"
                                    radiusType="circle"
                                    backgroundColorType="white"
                                    border="gray-light"
                                    fontSize="14px"
                                    fontFamily="InterSemiBold"
                                    onClick={() => close(value)}
                                />
                                <Button
                                    width="80%"
                                    height="45px"
                                    placeholder={'Update'}
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="14px"
                                    fontFamily="InterSemiBold"
                                    isLoading={loadingSubmit}
                                    onClick={updateIntegration}
                                    disabled={!isIntegrated || (formFields?.keys?.token === '' && !isChanged)}
                                />
                            </div>
                        </Form.Item>
                    </Form>
                </>
            )}
        </dynamic-integration>
    );
};

export default GitHubIntegration;
