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
import React, { useState } from 'react';
import { FiGitCommit } from 'react-icons/fi';
import { BiDownload } from 'react-icons/bi';
import { Divider, Rate } from 'antd';
import Button from '../../../../components/button';
import { isCloud, parsingDate } from '../../../../services/valueConvertor';
import { ReactComponent as MemphisFunctionIcon } from '../../../../assets/images/memphisFunctionIcon.svg';
import { ReactComponent as FunctionIcon } from '../../../../assets/images/functionIcon.svg';
import { ReactComponent as CodeBlackIcon } from '../../../../assets/images/codeIconBlack.svg';
import { ReactComponent as GithubBranchIcon } from '../../../../assets/images/githubBranchIcon.svg';

import CustomTabs from '../../../../components/Tabs';
import { OWNER } from '../../../../const/globalConst';
import { FiChevronDown } from 'react-icons/fi';
import { GoRepo } from 'react-icons/go';

import { Language } from '@material-ui/icons';

function FunctionDetails({ selectedFunction, integrated }) {
    const [open, setOpen] = useState(false);
    const [tabValue, setTabValue] = useState('Details');
    const [codeTabValue, setCodeTabValue] = useState('Code');
    const [isTestFunctionModalOpen, setIsTestFunctionModalOpen] = useState(false);

    return (
        <div className="function-drawer-container">
            <div className="drawer-header ">
                <FunctionIcon alt="Function icon" height="120px" width="120px" />
                <div className="right-side">
                    <div className="title">{selectedFunction?.function_name}</div>
                    <div>
                        <deatils is="x3d">
                            <div className="function-owner">
                                {selectedFunction.owner === OWNER && <MemphisFunctionIcon alt="Memphis function icon" height="15px" />}
                                <owner is="x3d">{selectedFunction?.owner === OWNER ? 'Memphis.dev' : selectedFunction?.owner}</owner>
                            </div>
                            <Divider type="vertical" />
                            {/* <downloads is="x3d">
                                <BiDownload className="download-icon" />
                                <label>{Number(1940).toLocaleString()}</label>
                            </downloads>
                            <Divider type="vertical" />
                            <rate is="x3d">
                                <Rate disabled defaultValue={2} className="stars-rate" />
                                <label>(93)</label>
                            </rate>
                            <Divider type="vertical" /> */}
                            <commits is="x3d">
                                <FiGitCommit />
                                <label>Last commit on {parsingDate(selectedFunction?.last_commit, false, false)}</label>
                            </commits>
                        </deatils>
                    </div>

                    <info is="x3d">
                        <repo is="x3d">
                            <GoRepo />
                            <label>{selectedFunction?.repository}</label>
                        </repo>
                        <branch is="x3d">
                            <GithubBranchIcon />
                            <label>{selectedFunction?.branch}</label>
                        </branch>
                        <language is="x3d">
                            <CodeBlackIcon />
                            <label>{selectedFunction?.language}</label>
                        </language>
                    </info>
                    <description is="x3d">{selectedFunction?.description}</description>
                    <actions is="x3d">
                        <Button
                            placeholder={
                                <div className="button-content">
                                    <span>Install</span>
                                    <div className="gradient" />
                                    <FiChevronDown />
                                </div>
                            }
                            backgroundColorType={'purple'}
                            colorType={'white'}
                            radiusType={'circle'}
                            onClick={() => {
                                // installFunction() - not implemented yet
                                return;
                            }}
                            disabled={isCloud() && !integrated}
                        />
                        <Button
                            placeholder={
                                <div className="button-content">
                                    <span>Test</span>
                                    <div className="gradient" />
                                    <FiChevronDown />
                                </div>
                            }
                            backgroundColorType={'orange'}
                            colorType={'black'}
                            radiusType={'circle'}
                            onClick={() => setIsTestFunctionModalOpen(true)}
                            disabled={!selectedFunction?.is_installed}
                        />
                    </actions>
                </div>
            </div>
            <div>
                <CustomTabs tabs={['Details', 'Code']} value={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
                {tabValue === 'Code' && <CustomTabs tabs={['Code', 'Versions']} value={tabValue} onChange={(tabValue) => setCodeTabValue(tabValue)} />}
            </div>
        </div>
    );
}

export default FunctionDetails;
