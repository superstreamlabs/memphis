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

import React, { useState, useEffect } from 'react';
import { Divider, Form } from 'antd';
import { IoClose } from 'react-icons/io5';
import { ReactComponent as GithubBranchIcon } from 'assets/images/githubBranchIcon.svg';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import SelectComponent from 'components/select';
import Button from 'components/button';
import { FiPlus } from 'react-icons/fi';

const IntegrationItem = ({ isNew, index, disabled, repo, reposList, updateIntegrationList, removeRepo, updateIntegration, addIsLoading, isEdittingIntegration }) => {
    const [isEditting, setIsEditting] = useState(false);
    const [formFields, setFormFields] = useState({
        type: 'functions',
        repo_name: null,
        repo_owner: null,
        branch: null
    });
    const [branches, setBranches] = useState([]);

    useEffect(() => {
        repo.repo_name && repo.repo_owner && getSourceCodeBranches(repo.repo_name, repo.repo_owner);
        setFormFields({ repo_name: repo.repo_name, repo_owner: repo.repo_owner, branch: repo.branch, type: 'functions' });
    }, [repo]);

    useEffect(() => {
        branches?.length > 0 && isEditting && updateBranch(branches[0]);
    }, [branches]);

    useEffect(() => {
        formFields.branch && updateIntegrationList(formFields, index);
    }, [formFields.branch]);

    useEffect(() => {
        isEditting && formFields?.repo_name && formFields?.repo_owner && getSourceCodeBranches(formFields?.repo_name, formFields?.repo_owner);
    }, [formFields?.repo_name]);

    const updateRepo = (repo) => {
        setIsEditting(true);
        getSourceCodeBranches(repo, reposList[repo]);
        setFormFields((formFields) => ({ ...formFields, ...{ repo_name: repo, repo_owner: reposList[repo], branch: '' } }));
    };

    const updateBranch = (branch) => {
        setIsEditting(true);
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
                <GithubBranchIcon alt="githubBranchIcon" />
                <Form.Item className="button-container">
                    <SelectComponent
                        colorType="black"
                        backgroundColorType="none"
                        radiusType="semi-round"
                        borderColorType="gray"
                        height="32px"
                        width={'90%'}
                        popupClassName="select-options"
                        value={formFields?.repo_name}
                        disabled={disabled}
                        onChange={(e) => {
                            updateRepo(e);
                        }}
                        options={Object?.keys(reposList)}
                    />
                </Form.Item>

                <Form.Item className="button-container">
                    <SelectComponent
                        colorType="black"
                        backgroundColorType="none"
                        radiusType="semi-round"
                        borderColorType="gray"
                        height="32px"
                        width={'90%'}
                        value={formFields?.branch}
                        options={branches || []}
                        popupClassName="select-options"
                        disabled={!isNew}
                        onChange={(e) => {
                            updateBranch(e);
                        }}
                    />
                </Form.Item>

                {!isNew ? (
                    <Button
                        height={'30px'}
                        width={'95px'}
                        placeholder={
                            <div className="repo-button">
                                <IoClose value={{ color: '#FC3400', size: '16' }} /> <span>Remove</span>
                            </div>
                        }
                        borderColorType="red"
                        colorType={'red'}
                        backgroundColorType={'white'}
                        radiusType={'circle'}
                        fontFamily="InterSemiBold"
                        onClick={() => removeRepo(index)}
                        disabled={isEdittingIntegration}
                        fontSize={'14px'}
                        border={'red'}
                    />
                ) : (
                    <Button
                        height={'30px'}
                        width={'90px'}
                        placeholder={
                            !addIsLoading && (
                                <div className="repo-button">
                                    <FiPlus style={{ marginRight: '5px' }} /> <span>Add</span>
                                </div>
                            )
                        }
                        colorType={'white'}
                        radiusType={'circle'}
                        backgroundColorType="purple"
                        fontSize="14px"
                        fontFamily="InterSemiBold"
                        isLoading={addIsLoading}
                        disabled={!formFields?.repo_name || !formFields?.branch}
                        onClick={() => updateIntegration()}
                    />
                )}
            </div>
            <Divider />
        </div>
    );
};

export default IntegrationItem;
