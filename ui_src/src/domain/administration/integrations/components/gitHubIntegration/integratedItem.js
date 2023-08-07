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

const IntegrationItem = ({ index, repo, reposList, updateIntegrationList }) => {
    const [creationForm] = Form.useForm();
    const [state, dispatch] = useContext(Context);
    const [isEditting, setIsEditting] = useState(false);
    const [formFields, setFormFields] = useState({
        type: 'functions'
    });
    const [branches, setBranches] = useState([]);

    useEffect(() => {
        console.log('repo', repo);
        setFormFields({ repo_name: repo.repo_name || repo.repository, repo_owner: repo.owner, branch: repo.branch, type: 'functions' });
    }, [repo]);

    useEffect(() => {
        isEditting && updateIntegrationList(formFields, index);
    }, [formFields.branch]);

    useEffect(() => {
        formFields?.repo_name && formFields?.repo_owner && getSourceCodeBranches(formFields?.repo_name, formFields?.repo_owner);
    }, [formFields?.repo_name]);

    const updateRepo = (repo) => {
        setIsEditting(true);
        setFormFields((formFields) => ({ ...formFields, ...{ repo_name: repo, repo_owner: reposList[repo], branch: '' } }));
    };

    const updateBranch = (branch) => {
        isEditting && setFormFields((formFields) => ({ ...formFields, ...{ branch: branch } }));
        setIsEditting(true);
    };

    const getSourceCodeBranches = async (repo, repo_owner) => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_SOURCE_CODE_BRANCHES}?repo_name=${repo}&repo_owner=${repo_owner}`);
            updateBranch(data?.branches[repo][0]);
            setBranches(data?.branches[repo]);
        } catch (error) {}
    };

    return (
        <div>
            <div className="repos-item" repo={repo}>
                <img src={githubBranchIcon} alt="githubBranchIcon" />
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
                        value={formFields?.branch || branches[0]}
                        options={branches}
                        popupClassName="select-options"
                        onChange={(e) => {
                            updateBranch(e);
                        }}
                    />
                </Form.Item>
            </div>
            <Divider />
        </div>
    );
};

export default IntegrationItem;
