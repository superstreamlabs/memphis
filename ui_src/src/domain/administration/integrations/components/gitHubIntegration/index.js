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

import { FiPlus } from 'react-icons/fi';
import { INTEGRATION_LIST, getTabList } from '../../../../../const/integrationList';
import { INTEGRATION_LIST, getTabList } from '../../../../../const/integrationList';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Button from '../../../../../components/button';
import { Context } from '../../../../../hooks/store';
import CustomTabs from '../../../../../components/Tabs';
import Loader from '../../../../../components/loader';
import IntegrationItem from './integratedItem';
import { showMessages } from '../../../../../services/genericServices';
import IntegrationDetails from '../integrationItem/integrationDetails';
import IntegrationLogs from '../integrationItem/integrationLogs';
import { ReactComponent as GithubNoConnectionIcon } from '../../../../../assets/images/noConnectionIcon.svg';

const GitHubIntegration = ({ close, value }) => {
    const githubConfiguration = INTEGRATION_LIST['GitHub'];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [applicationName, setApplicationName] = useState(null);
    const [formFields, setFormFields] = useState({
        name: 'github',
        keys: {
            token: '',
            connected_repos: []
        }
    });
    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);
    const [loadingRepos, setLoadingRepos] = useState(true);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [repos, setRepos] = useState([]);
    const [addNew, setAddNew] = useState(false);
    const [isChanged, setIsChanged] = useState(false);
    const [isIntegrated, setIsIntagrated] = useState(false);
    const [tabValue, setTabValue] = useState('Configuration');
    const tabs = getTabList('GitHub');

    useEffect(() => {
        value && Object.keys(value).length > 0 && setIsIntagrated(true);
        const images = [];
        images.push(INTEGRATION_LIST['GitHub'].banner.props.src);
        images.push(INTEGRATION_LIST['GitHub'].insideBanner.props.src);
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
    }, [value]);

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
        setIsIntagrated(false);

        showMessages('success', disconnect ? 'The integration was successfully disconnected' : 'The integration connected successfully');
    };

    const updateIntegration = async () => {
        setLoadingSubmit(true);
        const updatedFields = cleanEmptyFields();
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATION, updatedFields);
            dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
            setAddNew(false);
            await getIntegration();
        } catch (err) {
        } finally {
            setLoadingSubmit(false);
        }
    };

    const getIntegration = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_DETAILS}?name=github`);
            if (data) {
                if (data?.integration) {
                    setIsIntagrated(true);
                } else {
                    setIsIntagrated(false);
                }

                updateKeysState('connected_repos', data?.integration?.keys?.connected_repos || []);
                setRepos(data?.repos);
                setApplicationName(data?.application_name);
            } else
                setFormFields({
                    name: 'github',
                    keys: {
                        token: '',
                        connected_repos: []
                    }
                });
        } catch (error) {
        } finally {
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
                        <div className={'action-buttons flex-end'}>
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
                        </div>
                    </div>
                    <CustomTabs value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabs={tabs} />
                    <Form name="form" form={creationForm} autoComplete="off" className="integration-form">
                        {tabValue === 'Details' && <IntegrationDetails integrateDesc={githubConfiguration.integrateDesc} />}
                        {tabValue === 'Logs' && <IntegrationLogs integrationName={'github'} />}
                        {tabValue === 'Configuration' && (
                            <div className="integration-body">
                                <IntegrationDetails integrateDesc={githubConfiguration.integrateDesc} />
                                <div className="api-details">
                                    {!isIntegrated && (
                                        <div className="noConnection-wrapper">
                                            <GithubNoConnectionIcon />
                                            <p className="noConnection-title">Connect to get more details</p>
                                            <p className="noConnection-subtitle">Lorem Ipsum is simply dummy text of the printing and typesetting industry.</p>

                                            <Button
                                                height="35px"
                                                placeholder="Connect"
                                                colorType="white"
                                                radiusType="circle"
                                                backgroundColorType="purple"
                                                border="none"
                                                fontSize="12px"
                                                fontFamily="InterSemiBold"
                                                disabled={!applicationName}
                                                onClick={() => window.location.assign(`https://github.com/apps/${applicationName}/installations/select_target`)}
                                            />
                                        </div>
                                    )}
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
                                                                type={index === formFields?.keys?.connected_repos?.length - 1 && addNew}
                                                                updateIntegration={updateIntegration}
                                                                addIsLoading={loadingSubmit}
                                                            />
                                                        );
                                                    })}
                                                    {!addNew && (
                                                        <div
                                                            className="add-more-repos"
                                                            onClick={() => {
                                                                updateKeysConnectedRepos(
                                                                    {
                                                                        type: 'functions',
                                                                        repo_name: '',
                                                                        repo_owner: '',
                                                                        branch: ''
                                                                    },
                                                                    formFields.keys?.connected_repos?.length
                                                                );
                                                                setAddNew((prev) => !prev);
                                                            }}
                                                        >
                                                            <FiPlus />
                                                            <label> {formFields?.keys?.connected_repos?.length === 0 ? `Add the first repo` : `Add more repos`}</label>
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}
                        <Form.Item className="button-container">
                            <div className="button-wrapper button-wrapper-single-item  ">
                                <div></div>
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
                            </div>
                        </Form.Item>
                    </Form>
                </>
            )}
        </dynamic-integration>
    );
};

export default GitHubIntegration;
