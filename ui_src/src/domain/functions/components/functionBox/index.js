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

import { IoClose, IoGitBranch } from 'react-icons/io5';
import { FaCode } from 'react-icons/fa';
import { FiGitCommit } from 'react-icons/fi';
import { Drawer } from 'antd';
import React, { useState, useEffect } from 'react';
import { useHistory } from 'react-router-dom';

import { parsingDate } from '../../../../services/valueConvertor';
import TagsList from '../../../../components/tagList';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import pathDomains from '../../../../router';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import Tag from '../../../../components/tag';
import CustomTabs from '../../../../components/Tabs';

function FunctionBox({ key, funcDetails }) {
    const history = useHistory();
    const [functionDetails, setFunctionDetils] = useState(funcDetails);
    const [open, setOpen] = useState(false);
    const [selectedFunction, setSelectedFunction] = useState('');
    const [tabValue, setTabValue] = useState('Code');

    useEffect(() => {
        const url = window.location.href;
        const functionName = url.split('functions/')[1];
        if (functionName === functionDetails?.name) {
            setOpen(true);
            setSelectedFunction(functionName);
        }
    }, []);

    useEffect(() => {
        setFunctionDetils(funcDetails);
    }, [funcDetails]);

    const handleDrawer = (flag) => {
        setOpen(flag);
        if (flag) {
            history.push(`${pathDomains.functions}/${functionDetails?.name}`);
            setSelectedFunction(functionDetails?.name);
        } else {
            history.push(`${pathDomains.functions}`);
            setSelectedFunction('');
        }
    };

    const removeTag = async (tagName, schemaName) => {
        try {
            await httpRequest('DELETE', `${ApiEndpoints.REMOVE_TAG}`, { name: tagName, entity_type: 'schema', entity_name: schemaName });
            functionDetails.tags = functionDetails?.tags.filter((tag) => tag.name !== tagName);
            setFunctionDetils({ ...functionDetails });
        } catch (error) {}
    };

    const updateTags = async (tags) => {
        functionDetails.tags = tags;
        setFunctionDetils({ ...functionDetails });
    };

    return (
        <>
            <div key={functionDetails?.name} className={selectedFunction === functionDetails.name ? 'function-box-wrapper func-selected' : 'function-box-wrapper'}>
                <header is="x3d" onClick={() => handleDrawer(true)}>
                    <div className="function-name">
                        <OverflowTip text={functionDetails?.name} maxWidth={'300px'}>
                            <span>{functionDetails?.name}</span>
                        </OverflowTip>
                    </div>
                    <div className="function-details">
                        <div className="function-repo">
                            <IoGitBranch />
                            <OverflowTip text={functionDetails?.repo} maxWidth={'150px'}>
                                <span>memphiscloud - master</span>
                            </OverflowTip>
                        </div>
                        <div className="function-code-type">
                            <FaCode />
                            <OverflowTip text={functionDetails?.repo} maxWidth={'150px'}>
                                <span>.net</span>
                            </OverflowTip>
                        </div>
                    </div>
                </header>
                <tags is="x3d">
                    <TagsList
                        tagsToShow={3}
                        tags={functionDetails?.tags}
                        editable
                        entityType="schema"
                        entityName={functionDetails?.name}
                        handleDelete={(tag) => removeTag(tag, functionDetails?.name)}
                        handleTagsUpdate={(tags) => updateTags(tags)}
                    />
                </tags>
                <description is="x3d" onClick={() => handleDrawer(true)}>
                    <span>
                        Donec dictum tristique porta. Etiam convallis lorem lobortis nulla molestie, nec tincidunt ex ullamcorper. Quisque ultrices lobortis elit sed
                        euismod. Duis in ultrices dolor, ac rhoncus
                    </span>
                </description>
                <date is="x3d" onClick={() => handleDrawer(true)}>
                    <div className="flex">
                        <FiGitCommit />
                        <p>Commits on {parsingDate(functionDetails?.created_at, false, false)}</p>
                    </div>
                    <Tag editable={false} tag={{ name: 'Community', color: '101, 87, 255' }} rounded={true} />
                </date>
            </div>
            <Drawer
                title={
                    <div>
                        <p>{functionDetails?.name}</p>
                        <CustomTabs tabs={['Code']} value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
                    </div>
                }
                placement="bottom"
                size={'large'}
                className="function-drawer"
                onClose={() => handleDrawer(false)}
                destroyOnClose={true}
                open={open}
                maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
                closeIcon={<IoClose style={{ color: '#D1D1D1', width: '25px', height: '25px' }} />}
            ></Drawer>
        </>
    );
}

export default FunctionBox;
