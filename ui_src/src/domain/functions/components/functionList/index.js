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

import React, { useEffect, useState } from 'react';
import GitHubIntegration from '../../../administration/integrations/components/gitHubIntegration';
import { ReactComponent as PlaceholderFunctionsIcon } from '../../../../assets/images/placeholderFunctions.svg';
import { ReactComponent as SearchIcon } from '../../../../assets/images/searchIcon.svg';
import { ReactComponent as CloneModalIcon } from '../../../../assets/images/cloneModalIcon.svg';
import { ReactComponent as RefreshIcon } from '../../../../assets/images/refresh.svg';
import { ReactComponent as GitHubLogo } from '../../../../assets/images/githubLogo.svg';
import { ReactComponent as RepoIcon } from '../../../../assets/images/repoPurple.svg';
import { ReactComponent as PurpleQuestionMark } from '../../../../assets/images/purpleQuestionMark.svg';
import { ReactComponent as MemphisLogo } from '../../../../assets/images/logo.svg';
import CollapseArrow from '../../../../assets/images/collapseArrow.svg';
import { BiCode } from 'react-icons/bi';
import { MdDone } from 'react-icons/md';
import { AddRounded } from '@material-ui/icons';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
import { parsingDate } from '../../../../services/valueConvertor';
import { isCloud } from '../../../../services/valueConvertor';
import OverflowTip from '../../../../components/tooltip/overflowtip';
import Loader from '../../../../components/loader';
import Button from '../../../../components/button';
import Modal from '../../../../components/modal';
import SearchInput from '../../../../components/searchInput';
import CustomTabs from '../../../../components/Tabs';
import CloudOnly from '../../../../components/cloudOnly';
import FunctionBox from '../functionBox';
import IntegrateFunction from '../integrateFunction';
import FunctionsGuide from '../functionsGuide';
import CloneModal from '../cloneModal';
import { OWNER } from '../../../../const/globalConst';
import { Collapse, Divider, Popover } from 'antd';
import { LOCAL_STORAGE_FUNCTION_PAGE_VIEW } from '../../../../const/localStorageConsts';
const { Panel } = Collapse;

const TABS = [
    {
        name: 'All',
        disabled: false
    },
    {
        name: 'Memphis',
        disabled: false
    },
    {
        name: isCloud() ? (
            'Private'
        ) : (
            <>
                Private <CloudOnly position={'relative'} />
            </>
        ),
        disabled: !isCloud()
    }
];

function FunctionList({ tabPrivate }) {
    const [isLoading, setisLoading] = useState(true);
    const [modalIsOpen, modalFlip] = useState(false);
    const [cloneTooltipIsOpen, cloneTooltipIsOpenFlip] = useState(false);
    const [integrated, setIntegrated] = useState(false);
    const [installedFunctionList, setInstalledFunctionList] = useState([]);
    const [otherFunctionList, setOtherFunctionList] = useState([]);
    const [filteredInstalledData, setFilteredInstalledData] = useState([]);
    const [filteredOtherData, setFilteredOtherData] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [tabValue, setTabValue] = useState(tabPrivate ? 'Private' : 'All');
    const [isFunctionsGuideOpen, setIsFunctionsGuideOpen] = useState(false);
    const [isCloneModalOpen, setIsCloneModalOpen] = useState(false);
    const [connectedRepos, setConnectedRepos] = useState([]);
    const [clickedRefresh, setClickedRefresh] = useState(false);
    const ExpandIcon = ({ isActive }) => <img className={isActive ? 'collapse-arrow open' : 'collapse-arrow close'} src={CollapseArrow} alt="collapse-arrow" />;

    const content = (
        <div className="git-repos-list">
            <div>
                <div className="git-repos-item">
                    <div className="left-section">
                        <MemphisLogo alt="repo" className="repo-item-icon-memphis" />
                        <span className="repo-data">
                            <OverflowTip text="memphis-dev-functions" center={false}>
                                memphis-dev-functions
                            </OverflowTip>
                            <OverflowTip text="master" width={'170px'} center={false}>
                                <label className="last-modified">master</label>
                            </OverflowTip>
                        </span>
                        <MdDone alt="Healty" />
                    </div>
                </div>
                <Divider />
            </div>
            {connectedRepos?.map((repo, index) => (
                <div key={index}>
                    <div className="git-repos-item">
                        <div className="left-section">
                            <RepoIcon alt="repo" className="repo-item-icon" />
                            <span className="repo-data">
                                <OverflowTip text={repo?.repo_name} center={false}>
                                    {repo?.repo_name}
                                </OverflowTip>
                                <OverflowTip text={`${repo?.branch} | ${parsingDate(repo?.last_stnc, false, false)}`} width={'170px'} center={false}>
                                    <label className="last-modified">
                                        {repo?.branch} | Last synced on {parsingDate(repo?.last_stnc, false, false)}
                                    </label>
                                </OverflowTip>
                            </span>
                            <MdDone alt="Healty" />
                        </div>
                    </div>
                    <Divider />
                </div>
            ))}
        </div>
    );

    useEffect(() => {
        getAllFunctions();
        if (localStorage.getItem(LOCAL_STORAGE_FUNCTION_PAGE_VIEW) !== 'true' && isCloud()) {
            setIsFunctionsGuideOpen(true);
            localStorage.setItem(LOCAL_STORAGE_FUNCTION_PAGE_VIEW, true);
        }
    }, []);

    const getAllFunctions = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_FUNCTIONS);
            setIntegrated(data?.scm_integrated);
            setInstalledFunctionList(data?.installed);
            setOtherFunctionList(data?.other);
            setConnectedRepos(data?.connected_repos);
            setTimeout(() => {
                setisLoading(false);
            }, 500);
        } catch (error) {
            setisLoading(false);
        }
    };

    const fetchFunctions = async () => {
        getAllFunctions();
        modalFlip(false);
    };

    useEffect(() => {
        let resultsInstalled = installedFunctionList;
        let resultsOther = otherFunctionList;
        if (tabValue === 'Private') {
            resultsInstalled = resultsInstalled.filter((func) => func?.owner !== OWNER);
            resultsOther = resultsOther.filter((func) => func?.owner !== OWNER);
        } else if (tabValue === 'Memphis') {
            resultsInstalled = resultsInstalled.filter((func) => func?.owner === OWNER);
            resultsOther = resultsOther.filter((func) => func?.owner === OWNER);
        }
        if (searchInput.length > 0) {
            resultsInstalled = resultsInstalled.filter(
                (func) =>
                    func?.function_name?.toLowerCase()?.includes(searchInput?.toLowerCase()) || func?.description?.toLowerCase()?.includes(searchInput.toLowerCase())
            );
            resultsOther = resultsOther.filter(
                (func) =>
                    func?.function_name?.toLowerCase()?.includes(searchInput?.toLowerCase()) || func?.description?.toLowerCase()?.includes(searchInput.toLowerCase())
            );
        }
        setFilteredInstalledData(resultsInstalled);
        setFilteredOtherData(resultsOther);
    }, [tabValue, searchInput, installedFunctionList, otherFunctionList]);

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };
    const handleCloseFunctionModal = () => {
        setIsFunctionsGuideOpen(false);
    };
    const handleNewFunctionModal = () => {
        handleCloseFunctionModal();
        modalFlip(true);
    };

    const renderNoFunctionsFound = () => (
        <div className="no-function-to-display">
            <PlaceholderFunctionsIcon width={150} alt="placeholderFunctions" />
            <p className="title">No functions found</p>
            <p className="sub-title">Please try to search again</p>
        </div>
    );

    const renderFunctionBoxes = (filter) =>
        !isCloud() ? (
            <>
                {filteredOtherData?.map((func, index) => (
                    <FunctionBox key={index} funcDetails={func} integrated={integrated} />
                ))}
            </>
        ) : filter === 'installed' ? (
            <>
                {filteredInstalledData?.map((func, index) => (
                    <FunctionBox key={index} funcDetails={func} integrated={integrated} installed />
                ))}
            </>
        ) : (
            <>
                {filteredOtherData?.map((func, index) => (
                    <FunctionBox key={index} funcDetails={func} integrated={integrated} />
                ))}
            </>
        );

    const drawCollapse = () => {
        if (isCloud() && tabValue === 'Private' && !integrated) return <IntegrateFunction onClick={() => setIsFunctionsGuideOpen(true)} />;
        const noFunctionsContent = filteredInstalledData?.length === 0 && filteredOtherData === 0 ? renderNoFunctionsFound() : null;
        const installedFunctionBoxesContent = filteredInstalledData?.length !== 0 ? <div className="cards-wrapper">{renderFunctionBoxes('installed')}</div> : null;
        const otherFunctionBoxesContent = filteredOtherData?.length !== 0 ? <div className="cards-wrapper">{renderFunctionBoxes('other')}</div> : null;

        if (!installedFunctionBoxesContent && !otherFunctionBoxesContent) return null;
        return (
            <div className="function-list-collapse">
                {!isCloud() && <div>{otherFunctionBoxesContent || noFunctionsContent}</div>}
                {isCloud() && !integrated && tabValue === 'Private' && (
                    <div className="cards-wrapper">
                        <IntegrateFunction onClick={() => setIsFunctionsGuideOpen(true)} />
                    </div>
                )}
                {isCloud() && (
                    <>
                        <Collapse defaultActiveKey={['1']} accordion={true} expandIcon={({ isActive }) => <ExpandIcon isActive={isActive} />} ghost>
                            <Panel
                                header={<div className="panel-header">{`Installed ${filteredInstalledData?.length > 0 && `(${filteredInstalledData?.length})`}`}</div>}
                                key={1}
                            >
                                <div>{installedFunctionBoxesContent || noFunctionsContent}</div>
                            </Panel>
                        </Collapse>
                        <Collapse defaultActiveKey={['2']} accordion={true} expandIcon={({ isActive }) => <ExpandIcon isActive={isActive} />} ghost>
                            <Panel header={<div className="panel-header">{`Other ${filteredOtherData?.length > 0 && `(${filteredOtherData?.length})`}`}</div>} key={2}>
                                <div>{otherFunctionBoxesContent || noFunctionsContent}</div>
                            </Panel>
                        </Collapse>
                    </>
                )}
            </div>
        );
    };
    const renderContent = () => {
        console.log(filteredInstalledData);
        console.log(filteredOtherData);
        const noFunctionsContent = filteredInstalledData?.length === 0 && filteredOtherData?.length ? renderNoFunctionsFound() : null;
        console.log(noFunctionsContent);
        return drawCollapse() || noFunctionsContent;
    };

    return (
        <div className="function-container">
            <div className="header-wraper">
                <div className="main-header-wrapper">
                    <div className="header-flex-wrapper">
                        <label className="main-header-h1">Functions</label>
                        <PurpleQuestionMark className="info-icon" alt="Integration info" onClick={() => setIsFunctionsGuideOpen(true)} />
                    </div>
                    <span className="memphis-label">Serverless functions to process ingested events "on the fly"</span>
                </div>
                <div className="action-section">
                    {isCloud() && integrated && (
                        <Button
                            width={'36px'}
                            height={'34px'}
                            placeholder={
                                <div className="button-content">{isLoading ? '' : <RefreshIcon alt="refreshIcon" style={{ path: { color: '#6557FF' } }} />}</div>
                            }
                            backgroundColorType={'white'}
                            colorType="black"
                            radiusType="circle"
                            border={'gray-light'}
                            isLoading={isLoading}
                            onClick={getAllFunctions}
                        />
                    )}
                    <Popover
                        placement="top"
                        title={
                            <div
                                className="git-repo git-refresh-title"
                                onClick={() => {
                                    if (!isCloud()) return; //Open cloud only banner
                                    modalFlip(true);
                                    setClickedRefresh(false);
                                }}
                            >
                                <AddRounded className="add" fontSize="small" />
                                <label>Add repositories</label>
                            </div>
                        }
                        content={content}
                        trigger="click"
                        overlayClassName="repos-popover"
                        open={clickedRefresh}
                        onOpenChange={(open) => setClickedRefresh(open)}
                    >
                        <connectedRepos is="x3d">
                            <GitHubLogo alt="github icon" />
                            <label>Connected Git Repository</label>
                            <Divider type="vertical" />
                            <img src={CollapseArrow} alt="arrow" className={clickedRefresh ? 'open' : 'collapse-arrow'} />
                        </connectedRepos>
                    </Popover>
                    <Popover
                        placement="bottomLeft"
                        content={<CloneModal />}
                        width="540px"
                        trigger="click"
                        overlayClassName="clone-popover"
                        open={cloneTooltipIsOpen}
                        onOpenChange={(open) => cloneTooltipIsOpenFlip(open)}
                    >
                        <Button
                            width="100px"
                            height="34px"
                            placeholder={
                                <span className="code-btn">
                                    <BiCode size={18} />
                                    <label>Code</label>
                                </span>
                            }
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            aria-controls="usecse-menu"
                            aria-haspopup="true"
                            onClick={() => cloneTooltipIsOpenFlip(true)}
                        />
                    </Popover>
                </div>
            </div>
            <div className="function-tabs">
                <CustomTabs tabs={TABS} defaultActiveKey={tabPrivate ? 'Private' : 'All'} tabValue={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
            </div>
            <SearchInput
                placeholder="Search here"
                colorType="navy"
                backgroundColorType="gray-dark"
                width="100%"
                height="34px"
                borderRadiusType="circle"
                borderColorType="none"
                boxShadowsType="none"
                iconComponent={<SearchIcon alt="searchIcon" />}
                onChange={handleSearch}
                value={searchInput}
            />
            <div className="function-list">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {!isLoading && renderContent()}
            </div>
            <Modal className="integration-modal" height="95vh" width="720px" displayButtons={false} clickOutside={() => modalFlip(false)} open={modalIsOpen}>
                <GitHubIntegration
                    close={(data) => {
                        if (Object.keys(data).length > 0) {
                            fetchFunctions();
                        } else {
                            modalFlip(false);
                        }
                    }}
                    value={{}}
                />
            </Modal>
            <Modal
                className="new-function-modal"
                width={'681px'}
                height={'95vh'}
                displayButtons={false}
                clickOutside={handleCloseFunctionModal}
                open={isFunctionsGuideOpen}
            >
                <FunctionsGuide handleClose={handleCloseFunctionModal} handleConfirm={handleNewFunctionModal} handleCloneClick={() => setIsCloneModalOpen(true)} />
            </Modal>
            <Modal
                header={<CloneModalIcon alt="cloneModalIcon" />}
                width="540px"
                displayButtons={false}
                clickOutside={() => setIsCloneModalOpen(false)}
                open={isCloneModalOpen}
            >
                <CloneModal />
            </Modal>
        </div>
    );
}

export default FunctionList;
