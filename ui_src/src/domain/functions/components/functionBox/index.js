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

import React, { useState, useEffect, useContext } from 'react';
import Skeleton from 'antd/lib/skeleton';
import { IoIosInformationCircle } from 'react-icons/io';
import { isCloud, parsingDate } from '../../../../services/valueConvertor';
import { FiGitCommit } from 'react-icons/fi';
import { BiDownload } from 'react-icons/bi';
import { IoClose } from 'react-icons/io5';
import { GoRepo } from 'react-icons/go';
import { ReactComponent as GithubBranchIcon } from '../../../../assets/images/githubBranchIcon.svg';
import { ReactComponent as MemphisFunctionIcon } from '../../../../assets/images/memphisFunctionIcon.svg';
import { ReactComponent as FunctionIcon } from '../../../../assets/images/functionIcon.svg';
import { ReactComponent as DeleteIcon } from '../../../../assets/images/deleteIcon.svg';
import { FaArrowCircleUp } from 'react-icons/fa';
import { Divider, Drawer, Rate } from 'antd';
import FunctionDetails from '../functionDetails';
import { showMessages } from '../../../../services/genericServices';
import TagsList from '../../../../components/tagList';
import Button from '../../../../components/button';
import Modal from '../../../../components/modal';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import { OWNER } from '../../../../const/globalConst';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import AttachFunctionModal from '../attachFunctionModal';
import CloudModal from '../../../../components/cloudModal';
import TestMockEvent from '../../components/testFunctionModal/components/testMockEvent';
import { Context } from '../../../../hooks/store';

function FunctionBox({ funcDetails, integrated, isTagsOn = true, onClick = null, onApply, doneUninstall, startInstallation, funcIndex, referredFunction }) {
    const [state, dispatch] = useContext(Context);
    const [functionDetails, setFunctionDetils] = useState(funcDetails);
    const [open, setOpen] = useState(false);
    const [selectedFunction, setSelectedFunction] = useState(null);
    const [isValid, setIsValid] = useState(false);
    const [chooseStationModal, setChooseStationModal] = useState(false);
    const [cloudModal, setCloudModal] = useState(false);
    const [loader, setLoader] = useState(false);
    const [openUpgradeModal, setOpenUpgradeModal] = useState(false);
    const [isTestFunctionModalOpen, setIsTestFunctionModalOpen] = useState(false);

    useEffect(() => {
        const url = window.location.href;
        const functionName = url.split('functions/')[1];
        if (functionName === functionDetails?.function_name) {
            !onClick && setOpen(true);
            setSelectedFunction(functionName);
        }
    }, []);

    useEffect(() => {
        setFunctionDetils(funcDetails);
        setIsValid(funcDetails?.is_valid);
    }, [funcDetails]);

    const handleDrawer = (flag) => {
        setOpen(flag);
        if (flag) {
            setSelectedFunction(functionDetails);
        } else {
            setSelectedFunction(null);
        }
    };

    const handleInstall = async () => {
        const bodyRequest = {
            function_name: functionDetails?.function_name,
            repo: functionDetails?.repo,
            owner: functionDetails?.owner,
            branch: functionDetails?.branch,
            scm_type: functionDetails?.scm,
            by_memphis: funcDetails?.by_memphis
        };
        try {
            await httpRequest('POST', ApiEndpoints.INSTALL_FUNCTION, bodyRequest);
            showMessages('success', `We are ${functionDetails?.updates_available ? 'updating' : 'installing'} the function for you. We will let you know once its done`);
            startInstallation(funcIndex, functionDetails?.updates_available);
        } catch (e) {
            return;
        }
    };

    const handleUnInstall = async () => {
        setLoader(true);
        const bodyRequest = {
            function_name: functionDetails?.function_name,
            repo: functionDetails?.repo,
            owner: functionDetails?.owner,
            branch: functionDetails?.branch,
            scm_type: functionDetails?.scm,
            compute_engine: functionDetails?.compute_engine
        };
        try {
            await httpRequest('DELETE', ApiEndpoints.UNINSTALL_FUNCTION, bodyRequest);
            doneUninstall(funcIndex);
        } catch (e) {
        } finally {
            setLoader(false);
        }
    };

    return (
        <>
            <div
                key={functionDetails?.function_name}
                className={
                    selectedFunction?.function_name === functionDetails.function_name || referredFunction?.function_name === functionDetails.function_name
                        ? 'function-box-wrapper func-selected'
                        : 'function-box-wrapper'
                }
                onClick={() => (onClick ? onClick() : isCloud() ? handleDrawer(true) : setCloudModal(true))}
            >
                <header is="x3d">
                    <div className={`function-box-header ${!isTagsOn && 'station'}`}>
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
                                                <label>{functionDetails?.repo}</label>
                                            </repo>
                                            <Divider type="vertical" />
                                            <branch is="x3d">
                                                <GithubBranchIcon />
                                                <label>{functionDetails?.branch}</label>
                                            </branch>
                                            <Divider type="vertical" />
                                        </>
                                    )}
                                    {functionDetails?.owner === OWNER && (
                                        <>
                                            <downloads is="x3d">
                                                <BiDownload className="download-icon" />
                                                <label>{Number(funcDetails?.forks).toLocaleString()}</label>
                                            </downloads>
                                            <Divider type="vertical" />
                                            <rate is="x3d">
                                                <Rate disabled defaultValue={functionDetails?.stars} className="stars-rate" />
                                                <label>{`(${funcDetails?.rates})`}</label>
                                            </rate>
                                            <Divider type="vertical" />
                                        </>
                                    )}
                                    <commits is="x3d">
                                        <FiGitCommit />
                                        <label>Last modified on {parsingDate(functionDetails?.installed_updated_at, true, true)}</label>
                                    </commits>
                                </deatils>
                                <description is="x3d">
                                    {isValid ? (
                                        functionDetails?.description
                                    ) : (
                                        <Skeleton.Button
                                            active={false}
                                            style={{ width: '200px', height: '7px', borderRadius: '2px', background: '#e8e8e8', minWidth: '40px' }}
                                        />
                                    )}
                                </description>
                            </div>
                        </div>
                        <div onClick={(e) => e.stopPropagation()} className="install-button">
                            {!isValid && (
                                <div className="warning">
                                    <IoIosInformationCircle style={{ fontSize: '20px', color: '#fc3400' }} />
                                    <OverflowTip text={functionDetails?.invalid_reason} maxWidth={'260px'} textColor="#fc3400">
                                        <label className="warning-message">{functionDetails?.invalid_reason}</label>
                                    </OverflowTip>
                                </div>
                            )}
                            {isCloud() && functionDetails?.installed && (
                                <Button
                                    placeholder="Test"
                                    width={'100px'}
                                    backgroundColorType={'orange'}
                                    colorType={'black'}
                                    radiusType={'circle'}
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    onClick={() => setIsTestFunctionModalOpen(true)}
                                    disabled={!functionDetails?.installed || functionDetails?.installed_in_progress}
                                />
                            )}
                            {isValid && (!isCloud() || functionDetails?.installed) && (
                                <Button
                                    width="100px"
                                    height="34px"
                                    placeholder={
                                        isCloud() && !state?.allowedActions?.can_apply_functions ? (
                                            <span className="attach-btn">
                                                <label>Attach</label>
                                                <FaArrowCircleUp className="lock-feature-icon" />
                                            </span>
                                        ) : (
                                            <span className="attach-btn">Attach</span>
                                        )
                                    }
                                    purple-light
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType={'purple'}
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    disabled={isCloud() && functionDetails?.installed_in_progress}
                                    onClick={() => {
                                        if (!isCloud()) setCloudModal(true);
                                        else if (isTagsOn) setChooseStationModal(true);
                                        else {
                                            state?.allowedActions?.can_apply_functions ? onApply() : setOpenUpgradeModal(true);
                                        }
                                    }}
                                />
                            )}

                            {isCloud() && (
                                <Button
                                    width={functionDetails?.installed && !functionDetails?.updates_available ? '34px' : '100px'}
                                    height="34px"
                                    placeholder={
                                        functionDetails?.installed_in_progress ? (
                                            ''
                                        ) : functionDetails?.installed ? (
                                            <div className="code-btn">
                                                {functionDetails?.updates_available ? <BiDownload className="Install" /> : <DeleteIcon className="Uninstall" />}
                                                {functionDetails?.updates_available ? <label>Update</label> : null}
                                            </div>
                                        ) : (
                                            <div className="code-btn">
                                                <BiDownload className="Install" />
                                                <label>Install</label>
                                            </div>
                                        )
                                    }
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType={functionDetails?.installed && !functionDetails?.updates_available ? 'white' : 'purple'}
                                    border={functionDetails?.installed && !functionDetails?.updates_available ? 'gray-light' : null}
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    disabled={!isValid || functionDetails?.installed_in_progress}
                                    isLoading={loader || functionDetails?.installed_in_progress}
                                    onClick={() => {
                                        !functionDetails?.installed || functionDetails?.updates_available ? handleInstall() : handleUnInstall();
                                    }}
                                />
                            )}
                        </div>
                    </div>
                </header>
                {isTagsOn && isValid && (
                    <footer is="x3d">
                        <TagsList tagsToShow={3} tags={functionDetails?.tags} entityType="function" entityName={functionDetails?.function_name} />
                    </footer>
                )}
                {isTagsOn && !isValid && (
                    <footer is="x3d">
                        <Skeleton.Button active={false} />
                        <Skeleton.Button active={false} />
                        <Skeleton.Button active={false} />
                    </footer>
                )}
            </div>
            <Modal
                header={
                    <div className="modal-header">
                        <p>Attach a function</p>
                    </div>
                }
                displayButtons={false}
                height="400px"
                width="352px"
                clickOutside={() => setChooseStationModal(false)}
                open={chooseStationModal}
                hr={true}
                className="use-schema-modal"
            >
                <AttachFunctionModal selectedFunction={functionDetails} />
            </Modal>
            <CloudModal open={cloudModal} handleClose={() => setCloudModal(false)} type={'cloud'} />
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
                {selectedFunction && (
                    <FunctionDetails
                        selectedFunction={selectedFunction}
                        integrated={integrated}
                        handleUnInstall={handleUnInstall}
                        handleInstall={handleInstall}
                        clickApply={() => setChooseStationModal(true)}
                    />
                )}
            </Drawer>
            <CloudModal type="upgrade" open={openUpgradeModal} handleClose={() => setOpenUpgradeModal(false)} />
            <Modal width={'75vw'} height={'80vh'} clickOutside={() => setIsTestFunctionModalOpen(false)} open={isTestFunctionModalOpen} displayButtons={false}>
                <TestMockEvent functionDetails={funcDetails} open={isTestFunctionModalOpen} />
            </Modal>
        </>
    );
}

export default FunctionBox;
