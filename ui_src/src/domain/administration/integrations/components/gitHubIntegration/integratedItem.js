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
import { Divider, Form, message } from 'antd';

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

const IntegrationItem = ({ index, repo, reposList }) => {
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [formFields, setFormFields] = useState({
        type: 'functions'
    });
    const [repos, setRepos] = useState([]);
    const [branches, setBranches] = useState([]);

    useEffect(() => {
        // getIntegration();
        // console.log(reposList);
        // console.log(repo);
    }, []);

    // useEffect(() => {
    //     const images = [];
    //     images.push(INTEGRATION_LIST['GitHub'].banner.props.src);
    //     images.push(INTEGRATION_LIST['GitHub'].insideBanner.props.src);
    //     images.push(INTEGRATION_LIST['GitHub'].icon.props.src);
    //     const promises = [];

    //     images.forEach((imageUrl) => {
    //         const image = new Image();
    //         promises.push(
    //             new Promise((resolve) => {
    //                 image.onload = resolve;
    //             })
    //         );
    //         image.src = imageUrl;
    //     });

    //     Promise.all(promises).then(() => {
    //         setImagesLoaded(true);
    //     });
    // }, []);

    // const updateKeysState = (field, value) => {
    //     let updatedValue = { ...formFields.keys };
    //     updatedValue[field] = value;
    //     setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    // };

    // const closeModal = (data, disconnect = false) => {
    //     setTimeout(() => {
    //         disconnect ? setLoadingDisconnect(false) : setLoadingSubmit(false);
    //     }, 1000);
    //     close(data);
    //     message.success({
    //         key: 'memphisSuccessMessage',
    //         content: disconnect ? 'The integration was successfully disconnected' : 'The integration connected successfully',
    //         duration: 5,
    //         style: { cursor: 'pointer' },
    //         onClick: () => message.destroy('memphisSuccessMessage')
    //     });
    // };

    // const updateIntegration = async (withToken = true) => {
    //     let newFormFields = { ...formFields };
    //     try {
    //         const data = await httpRequest('POST', ApiEndpoints.UPDATE_INTEGRATION, { ...newFormFields });
    //         dispatch({ type: 'UPDATE_INTEGRATION', payload: data });
    //         // closeModal(data);
    //     } catch (err) {
    //         setLoadingSubmit(false);
    //     }
    // };

    // const createIntegration = async () => {
    //     try {
    //         const data = await httpRequest('POST', ApiEndpoints.CREATE_INTEGRATION, { ...formFields });
    //         dispatch({ type: 'ADD_INTEGRATION', payload: data });
    //         getIntegration();
    //     } catch (err) {
    //         setLoadingSubmit(false);
    //     }
    // };

    // const getIntegration = async () => {
    //     try {
    //         const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_DETAILS}?name=github`);
    //         // console.log(data);
    //         setRepos(data?.repos);
    //         // setConnectedRepos(data?.integration?.keys?.connected_repos);
    //     } catch (error) {}
    // };

    // useEffect(() => {
    //     console.log(formFields?.keys?.repo_name, formFields?.keys?.repo_owner);
    //     formFields?.keys?.repo_name && formFields?.keys?.repo_owner && getSourceCodeBranches(formFields?.keys?.repo_name, formFields?.keys?.repo_owner);
    // }, [formFields?.keys?.repo_name]);

    // const updateRepo = (repo) => {
    //     let updatedValue = { ...formFields.keys };
    //     updatedValue = { ...updatedValue, ...{ repo_name: repo, repo_owner: repos[repo] } };
    //     setFormFields((formFields) => ({ ...formFields, keys: updatedValue }));
    // };

    // const getSourceCodeBranches = async (repo, repo_owner) => {
    //     try {
    //         const data = await httpRequest('GET', `${ApiEndpoints.GET_SOURCE_CODE_BRANCHES}?repo_name=${repo}&repo_owner=${repo_owner}`);
    //         console.log(data);
    //         setBranches(data?.branches[formFields?.keys?.repo_name]);
    //     } catch (error) {}
    // };

    // const disconnect = async () => {
    //     setLoadingDisconnect(true);
    //     try {
    //         await httpRequest('DELETE', ApiEndpoints.DISCONNECT_INTEGRATION, {
    //             name: formFields.name
    //         });
    //         dispatch({ type: 'REMOVE_INTEGRATION', payload: formFields.name });

    //         closeModal({}, true);
    //     } catch (err) {
    //         setLoadingDisconnect(false);
    //     }
    // };
    useEffect(() => {
        console.log(formFields);
    }, [formFields]);

    useEffect(() => {
        formFields?.repo_name && formFields?.repo_owner && getSourceCodeBranches(formFields?.repo_name, formFields?.repo_owner);
    }, [formFields?.repo_name]);

    const updateRepo = (repo) => {
        setFormFields((formFields) => ({ ...formFields, ...{ repo_name: repo, repo_owner: reposList[repo], branch: '' } }));
    };

    const updateBranch = (branch) => {
        setFormFields((formFields) => ({ ...formFields, ...{ branch: branch } }));
    };

    const getSourceCodeBranches = async (repo, repo_owner) => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_SOURCE_CODE_BRANCHES}?repo_name=${repo}&repo_owner=${repo_owner}`);
            setBranches(data?.branches[repo]);
        } catch (error) {}
    };

    return (
        <div>
            <div className="repos-item" repo={repo}>
                <Form.Item className="button-container">
                    <span className="select-repo-span">
                        {/* <img src={githubBranchIcon} alt="githubBranchIcon" /> */}
                        <SelectComponent
                            colorType="black"
                            backgroundColorType="none"
                            radiusType="semi-round"
                            borderColorType="gray"
                            height="32px"
                            width={'180px'}
                            popupClassName="select-options"
                            value={formFields?.repo_name || repo.repository}
                            onChange={(e) => {
                                updateRepo(e);
                            }}
                            options={Object?.keys(reposList)}
                        />
                    </span>
                </Form.Item>

                <Form.Item className="button-container">
                    <SelectComponent
                        colorType="black"
                        backgroundColorType="none"
                        radiusType="semi-round"
                        borderColorType="gray"
                        height="32px"
                        width={'180px'}
                        value={formFields?.branch || repo.branch}
                        options={branches}
                        popupClassName="select-options"
                        onChange={(e) => {
                            updateBranch(e);
                        }}
                    />
                </Form.Item>
                <Form.Item className="button-container">
                    <SelectComponent
                        colorType="black"
                        backgroundColorType="none"
                        radiusType="semi-round"
                        borderColorType="none"
                        height="32px"
                        width={'180px'}
                        value={'functions'}
                        disabled
                    />
                </Form.Item>
                {/* <label onClick={updateIntegration}>btn</label> */}
            </div>
            <Divider />
        </div>
    );
};

export default IntegrationItem;
