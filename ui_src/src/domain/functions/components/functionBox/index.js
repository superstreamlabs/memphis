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

import React, { useState, useEffect } from 'react';
import { isCloud, parsingDate } from '../../../../services/valueConvertor';
import { FiGitCommit } from 'react-icons/fi';
import { BiDownload } from 'react-icons/bi';
import { MdOutlineFileDownloadOff } from 'react-icons/md';
import { IoClose } from 'react-icons/io5';
import { GoRepo } from 'react-icons/go';
import { ReactComponent as GithubBranchIcon } from '../../../../assets/images/githubBranchIcon.svg';
import { ReactComponent as MemphisFunctionIcon } from '../../../../assets/images/memphisFunctionIcon.svg';
import { ReactComponent as FunctionIcon } from '../../../../assets/images/functionIcon.svg';
import { Divider, Drawer, Rate } from 'antd';
import FunctionDetails from '../functionDetails';
import { showMessages } from '../../../../services/genericServices';
import TagsList from '../../../../components/tagList';
import CloudOnly from '../../../../components/cloudOnly';
import Button from '../../../../components/button';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import { OWNER } from '../../../../const/globalConst';
import AttachTooltip from '../AttachTooltip';

function FunctionBox({ funcDetails, integrated, installed }) {
    const [functionDetails, setFunctionDetils] = useState(funcDetails);
    const [open, setOpen] = useState(false);
    const [selectedFunction, setSelectedFunction] = useState('');

    useEffect(() => {
        const url = window.location.href;
        const functionName = url.split('functions/')[1];
        if (functionName === functionDetails?.function_name) {
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
            setSelectedFunction(functionDetails);
        } else {
            setSelectedFunction('');
        }
    };

    return (
        <>
            <div
                key={functionDetails?.function_name}
                className={selectedFunction?.function_name === functionDetails.function_name ? 'function-box-wrapper func-selected' : 'function-box-wrapper'}
                onClick={() => handleDrawer(true)}
            >
                <header is="x3d">
                    <div className="function-box-header">
                        <div className="details-section">
                            {funcDetails?.image ? <img src={funcDetails?.image} alt="Function icon" height="40px" /> : <FunctionIcon alt="Function icon" height="40px" />}
                            <div>
                                <div className="function-name">
                                    <OverflowTip text={functionDetails?.function_name} maxWidth={'250px'}>
                                        {functionDetails?.function_name}
                                    </OverflowTip>
                                </div>
                                <deatils is="x3d">
                                    <div className="function-owner">
                                        {funcDetails.owner === OWNER && <MemphisFunctionIcon alt="Memphis function icon" height="15px" />}
                                        <owner is="x3d">{functionDetails?.owner === OWNER ? 'Memphis.dev' : functionDetails?.owner}</owner>
                                    </div>
                                    <Divider type="vertical" />
                                    {funcDetails.owner !== OWNER && (
                                        <>
                                            <repo is="x3d">
                                                <GoRepo />
                                                <label>{functionDetails?.repository}</label>
                                            </repo>
                                            <Divider type="vertical" />
                                            <branch is="x3d">
                                                <GithubBranchIcon />
                                                <label>{functionDetails?.branch}</label>
                                            </branch>
                                            <Divider type="vertical" />
                                        </>
                                    )}
                                    <downloads is="x3d">
                                        <BiDownload className="download-icon" />
                                        <label>{Number(180).toLocaleString()}</label>
                                    </downloads>
                                    <Divider type="vertical" />
                                    <rate is="x3d">
                                        <Rate disabled defaultValue={5} className="stars-rate" />
                                        <label>(50)</label>
                                    </rate>
                                    <Divider type="vertical" />
                                    <commits is="x3d">
                                        <FiGitCommit />
                                        <label>Last commit on {parsingDate(functionDetails?.last_commit, false, false)}</label>
                                    </commits>
                                </deatils>
                                <description is="x3d">{functionDetails?.description}</description>
                            </div>
                        </div>
                        <div
                            onClick={(e) => {
                                e.stopPropagation();
                                // installFunction() - not implemented yet
                            }}
                            className="install-button"
                        >
                            <div className="header-flex">
                                <AttachTooltip disabled={!isCloud() || functionDetails?.install_in_progress || !installed} />
                                {!isCloud() && <CloudOnly position={'relative'} />}
                            </div>
                            <div className="header-flex">
                                <Button
                                    width="100px"
                                    height="34px"
                                    placeholder={
                                        functionDetails?.install_in_progress ? (
                                            ''
                                        ) : installed ? (
                                            <div className="code-btn">
                                                <MdOutlineFileDownloadOff className="Uninstall" />
                                                <label>Uninstall</label>
                                            </div>
                                        ) : (
                                            <div className="code-btn">
                                                <BiDownload className="Install" />
                                                <label>Install</label>
                                            </div>
                                        )
                                    }
                                    purple-light
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType={installed ? 'purple-light' : 'purple'}
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    disabled={!isCloud() || functionDetails?.install_in_progress}
                                    isLoading={functionDetails?.install_in_progress} //Get indication after install function
                                    onClick={() => {
                                        showMessages('success', 'Install function');
                                        return;
                                    }}
                                />
                                {!isCloud() && <CloudOnly position={'relative'} />}
                            </div>
                        </div>
                    </div>
                </header>
                <footer is="x3d">
                    <TagsList tagsToShow={3} tags={functionDetails?.tags} entityType="function" entityName={functionDetails?.function_name} />
                </footer>
            </div>
            <Drawer
                placement="right"
                size={'large'}
                className="function-drawer"
                onClose={() => handleDrawer(false)}
                destroyOnClose={true}
                open={open}
                maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
                closeIcon={<IoClose style={{ color: '#D1D1D1', width: '25px', height: '25px' }} />}
            >
                <FunctionDetails selectedFunction={selectedFunction} integrated={integrated} installed={installed} />
            </Drawer>
        </>
    );
}

export default FunctionBox;
