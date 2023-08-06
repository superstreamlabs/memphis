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

const urlSplit = URL.split('/', 3);

const GitHubIntegration = ({ close, value }) => {
    const isValue = value && Object.keys(value)?.length !== 0;
    const githubConfiguration = INTEGRATION_LIST['GitHub'];
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [formFields, setFormFields] = useState({
        name: 'github',
        ui_url: `${urlSplit[0]}//${urlSplit[2]}`,
        keys: {
            type: 'functions'
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

    const updateIntegration = async (withToken = true) => {
        let newFormFields = { ...formFields };
        if (!withToken) {
            let updatedKeys = { ...formFields.keys };
            updatedKeys['auth_token'] = '';
            newFormFields = { ...newFormFields, keys: updatedKeys };
        }
        try {
            const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATION, { ...newFormFields });
            dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
            closeModal(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const createIntegration = async () => {
        try {
            const data = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, { ...formFields });
            dispatch({ type: 'ADD_INTEGRATION', payload: data });
            getIntegration();
            // closeModal(data);
        } catch (err) {
            setLoadingSubmit(false);
        }
    };

    const getIntegration = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_DETAILS}?name=github`);
            console.log(data);
            setRepos(data?.repos);
            // setConnectedRepos(data?.integration?.keys?.connected_repos);
        } catch (error) {}
    };

    useEffect(() => {
        console.log(formFields?.keys?.repo_name, formFields?.keys?.owner);
        formFields?.keys?.repo_name && formFields?.keys?.owner && getSourceCodeBranches(formFields?.keys?.repo_name, formFields?.keys?.owner);
    }, [formFields?.keys?.repo_name]);

    const updateRepo = (repo) => {
        let updatedValue = { ...formFields.keys };
        updatedValue = { ...updatedValue, ...{ repo_name: repo, owner: repos[repo] } };
        setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    };

    const getSourceCodeBranches = async (repo, owner) => {
        console.log('in');
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_SOURCE_CODE_BRANCHES}?repo_name=${repo}&owner=${owner}`);
            console.log(data);
            setBranches(data?.branches[formFields?.keys?.repo_name]);
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
                                        // 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
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
                                        placeholder="****"
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
                            <p className="title">Repos</p>
                            <span className="desc">Lorem Ipsum is simply dummy text of the printing and typesetting industry.</span>
                            <div className="repos-container">
                                <div className="repos-header">
                                    <label>Repo Name</label>
                                    <label>Branch</label>
                                    <label>Type</label>
                                    <label>BTN</label>
                                </div>
                                <div className="repos-body">
                                    {state.integrationsList[0]?.keys?.connected_repos?.map((repo, index) => {
                                        return (
                                            <div className="repos-item" key={index}>
                                                <Form.Item className="button-container">
                                                    <span className="select-repo-span">
                                                        <img src={githubBranchIcon} alt="githubBranchIcon" />
                                                        <SelectComponent
                                                            colorType="black"
                                                            backgroundColorType="none"
                                                            radiusType="semi-round"
                                                            borderColorType="none"
                                                            height="40px"
                                                            width={'180px'}
                                                            popupClassName="select-options"
                                                            value={repo.repository}
                                                            disabled={true}
                                                        />
                                                    </span>
                                                </Form.Item>

                                                <Form.Item className="button-container">
                                                    <SelectComponent
                                                        colorType="black"
                                                        backgroundColorType="none"
                                                        radiusType="semi-round"
                                                        borderColorType="none"
                                                        height="40px"
                                                        width={'180px'}
                                                        value={repo.branch}
                                                        disabled
                                                    />
                                                </Form.Item>
                                                <Form.Item className="button-container">
                                                    <SelectComponent
                                                        colorType="black"
                                                        backgroundColorType="none"
                                                        radiusType="semi-round"
                                                        borderColorType="none"
                                                        height="40px"
                                                        width={'180px'}
                                                        value={'functions'}
                                                        disabled
                                                    />
                                                </Form.Item>
                                                <label onClick={updateIntegration}>btn</label>
                                            </div>
                                        );
                                    })}
                                    <div className="repos-item">
                                        <Form.Item className="button-container">
                                            <span className="select-repo-span">
                                                <img src={githubBranchIcon} alt="githubBranchIcon" />
                                                <SelectComponent
                                                    colorType="black"
                                                    backgroundColorType="none"
                                                    radiusType="semi-round"
                                                    borderColorType="none"
                                                    height="40px"
                                                    width={'180px'}
                                                    popupClassName="select-options"
                                                    options={Object?.keys(repos) || []}
                                                    placeholder={'Select a repo'}
                                                    value={formFields?.keys?.repo_name || ''}
                                                    onChange={(e) => {
                                                        updateRepo(e);
                                                    }}
                                                />
                                            </span>
                                        </Form.Item>

                                        <Form.Item className="button-container">
                                            <SelectComponent
                                                colorType="black"
                                                backgroundColorType="none"
                                                radiusType="semi-round"
                                                borderColorType="none"
                                                height="40px"
                                                width={'180px'}
                                                popupClassName="select-options"
                                                options={branches || []}
                                                // value={selectedBranch || ''}
                                                placeholder={'Select a branch'}
                                                // value={selectedRepo || Object.keys(repos)[0]}
                                                onChange={(e) => updateKeysState('branch', e)}
                                                // onChange={(e) => setSelectedBranch(e)}
                                            />
                                        </Form.Item>
                                        <Form.Item className="button-container">
                                            <SelectComponent
                                                colorType="black"
                                                backgroundColorType="none"
                                                // borderColorType="gray"
                                                radiusType="semi-round"
                                                // backgroundColorType="none"
                                                borderColorType="none"
                                                // radiusType="semi-round"
                                                height="40px"
                                                width={'180px'}
                                                popupClassName="select-options"
                                                options={['functions']}
                                                value={'functions'}
                                                placeholder={'Select type'}
                                                // value={selectedRepo || Object.keys(repos)[0]}
                                                onChange={(e) => updateKeysState('functions', e)}
                                            />
                                        </Form.Item>
                                        <label onClick={updateIntegration}>btn</label>
                                    </div>
                                </div>
                            </div>
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
                            </div>
                        </Form.Item>
                    </Form>
                </>
            )}
        </dynamic-integration>
    );
};

export default GitHubIntegration;
