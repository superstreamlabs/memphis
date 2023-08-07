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
import { Form, message } from 'antd';

import poisionAlertIcon from '../../../../../assets/images/poisionAlertIcon.svg';
import disconAlertIcon from '../../../../../assets/images/disconAlertIcon.svg';
import schemaAlertIcon from '../../../../../assets/images/schemaAlertIcon.svg';
import githubBranchIcon from '../../../../../assets/images/githubBranchIcon.svg';
import { INTEGRATION_LIST } from '../../../../../const/integrationList';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import Switcher from '../../../../../components/switcher';
import Button from '../../../../../components/button';
import { Context } from '../../../../../hooks/store';
import Input from '../../../../../components/Input';
import SelectComponent from '../../../../../components/select';
import { URL } from '../../../../../config';
import Loader from '../../../../../components/loader';
import IntegrationItem from './integratedItem';

const urlSplit = URL.split('/', 3);

const GitHubIntegration = ({ close, value }) => {
    const isValue = value && Object.keys(value)?.length !== 0;
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

    const [loadingSubmit, setLoadingSubmit] = useState(false);
    const [loadingDisconnect, setLoadingDisconnect] = useState(false);
    const [imagesLoaded, setImagesLoaded] = useState(false);
    const [repos, setRepos] = useState([]);
    const [branches, setBranches] = useState([]);

    useEffect(() => {
        getIntegration();
    }, []);

    useEffect(() => {
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
    }, []);

    const updateKeysConnectedRepos = (repo, index) => {
        if (repo.branch === '') return;
        let updatedValue = { ...formFields.keys };
        if (index && index < updatedValue.connected_repos?.length) updatedValue.connected_repos[index] = repo;
        else if (index === 0 && index === updatedValue.connected_repos?.length) {
            updatedValue.connected_repos.push(repo);
        }
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
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATION, formFields);
            dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
            closeModal(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const createIntegration = async () => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, formFields);
            dispatch({ type: 'ADD_INTEGRATION', payload: data });
            getIntegration();
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const getIntegration = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_DETAILS}?name=github`);
            if (data) {
                updateKeysState('connected_repos', data?.integaraion?.keys?.connected_repos || '');
                // setFormFields((formFields) => ({ ...formFields, ...keys, connected_repos: data?.integaraion?.keys?.connected_repos || [] }));
                // setFormFields(data?.integaraion);
                setRepos(data?.repos);
            } else
                setFormFields({
                    name: 'github',
                    keys: {
                        token: '',
                        connected_repos: []
                    }
                });
        } catch (error) {}
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
                        <div className={!isValue ? 'action-buttons flex-end' : 'action-buttons'}>
                            {isValue && (
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
                                onClick={() => window.open('https://docs.memphis.dev/memphis/dashboard-ui/integrations/notifications/slack', '_blank')}
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
                                    <Button
                                        width="70px"
                                        height="20px"
                                        placeholder={'Connect'}
                                        colorType="white"
                                        radiusType="circle"
                                        backgroundColorType="purple"
                                        fontSize="12px"
                                        fontFamily="InterSemiBold"
                                        // isLoading={loadingSubmit}
                                        // disabled={isValue && !creationForm.isFieldsTouched()}
                                        onClick={createIntegration}
                                    />
                                </span>
                                <span className="desc">The secret key associated with the access key.</span>
                                <Form.Item
                                    name="token"
                                    rules={[
                                        {
                                            required: true,
                                            message: 'Please insert a token.'
                                        }
                                    ]}
                                    initialValue={formFields?.keys?.token}
                                >
                                    <Input
                                        placeholder="ghp_****"
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
                            {formFields?.name && (
                                <div>
                                    <p className="title">Repos</p>
                                    <span className="desc">Lorem Ipsum is simply dummy text of the printing and typesetting industry.</span>
                                    <div className="repos-container">
                                        <div className="repos-header">
                                            <label></label>
                                            <label>Repo Name</label>
                                            <label>Branch</label>
                                            <label>Type</label>
                                        </div>
                                        <div className="repos-body">
                                            {formFields?.keys?.connected_repos?.map((repo, index) => {
                                                return (
                                                    <IntegrationItem
                                                        key={index}
                                                        index={index}
                                                        repo={repo}
                                                        reposList={repos || []}
                                                        updateIntegrationList={(updatedFields, i) => updateKeysConnectedRepos(updatedFields, i)}
                                                    />
                                                );
                                            })}
                                            <IntegrationItem
                                                index={formFields?.keys?.connected_repos?.length}
                                                repo={''}
                                                reposList={repos || []}
                                                updateIntegrationList={(updatedFields, i) => updateKeysConnectedRepos(updatedFields, i)}
                                            />
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
                                    placeholder={isValue ? 'Update' : 'Connect'}
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="14px"
                                    fontFamily="InterSemiBold"
                                    isLoading={loadingSubmit}
                                    // disabled={isValue && !creationForm.isFieldsTouched()}
                                    onClick={updateIntegration}
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
