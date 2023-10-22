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
import { ReactComponent as GithubActiveConnectionIcon } from '../../../../assets/images/githubActiveConnectionIcon.svg';
import { ReactComponent as SearchIcon } from '../../../../assets/images/searchIcon.svg';
import { ReactComponent as CloneModalIcon } from '../../../../assets/images/cloneModalIcon.svg';
import { ReactComponent as RefreshIcon } from '../../../../assets/images/refresh.svg';
import { ReactComponent as GitHubLogo } from '../../../../assets/images/githubLogo.svg';
import { ReactComponent as RepoIcon } from '../../../../assets/images/repoPurple.svg';
import CollapseArrow from '../../../../assets/images/collapseArrow.svg';
import { AddRounded } from '@material-ui/icons';
import { ApiEndpoints } from '../../../../const/apiEndpoints';
import { httpRequest } from '../../../../services/http';
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

function FunctionList() {
    const [isLoading, setisLoading] = useState(true);
    const [modalIsOpen, modalFlip] = useState(false);
    const [integrated, setIntegrated] = useState(false);
    const [functionList, setFunctionList] = useState([]);
    const [filteredData, setFilteredData] = useState([]);
    const [searchInput, setSearchInput] = useState('');
    const [filterItem, setFilterItem] = useState(null);
    const [tabValue, setTabValue] = useState('All');
    const [isFunctionsGuideOpen, setIsFunctionsGuideOpen] = useState(false);
    const [isCloneModalOpen, setIsCloneModalOpen] = useState(false);
    const [connectedRepos, setConnectedRepos] = useState([]);
    const [clickedRefresh, setClickedRefresh] = useState(false);
    const ExpandIcon = ({ isActive }) => <img className={isActive ? 'collapse-arrow open' : 'collapse-arrow close'} src={CollapseArrow} alt="collapse-arrow" />;

    const content = (
        <div className="git-repos-list">
            {connectedRepos?.map((repo, index) => (
                <>
                    <div className="git-repos-item" key={index} onClick={() => setFilterItem(index)}>
                        <div className="left-section">
                            <RepoIcon alt="repo" className={`repo-item-icon ${filterItem === index && 'filtered'}`} />
                            <span className="repo-data">
                                <label className="git-repo">{repo?.repo_name}</label>
                                <label className="last-modified">{repo?.branch}</label>
                            </span>
                        </div>
                    </div>
                    <Divider />
                </>
            ))}
            {filterItem !== null && (
                <div className="git-repos-item" onClick={() => setFilterItem(null)}>
                    <div className="left-section">
                        <RepoIcon alt="repo" className="repo-item-icon" />
                        <span className="repo-data">
                            <label className="git-repo">Show all</label>
                        </span>
                    </div>
                </div>
            )}
        </div>
    );

    useEffect(() => {
        getAllFunctions();
        isCloud() && getIntegrationDetails();
    }, []);

    const getIntegrationDetails = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_INTEGRATION_DETAILS}?name=github`);
            setConnectedRepos(data?.integration?.keys?.connected_repos || []);
        } catch (error) {}
    };

    const getAllFunctions = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_FUNCTIONS);
            setIntegrated(data.scm_integrated);
            setFunctionList(data?.functions);
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
        let results = functionList;
        if (tabValue === 'Private') {
            results = results.filter((func) => func?.owner !== OWNER);
        } else if (tabValue === 'Memphis') {
            results = results.filter((func) => func?.owner === OWNER);
        }
        if (searchInput.length > 0) {
            results = results.filter(
                (func) => func?.function_name?.toLowerCase()?.includes(searchInput.toLowerCase()) || func?.description?.toLowerCase()?.includes(searchInput.toLowerCase())
            );
        }
        if (filterItem) {
            results = results.filter((func) => func?.repository === connectedRepos[filterItem]?.repo_name && func?.branch === connectedRepos[filterItem]?.branch);
        }
        setFilteredData(results);
    }, [tabValue, searchInput, functionList, filterItem]);

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
        !integrated ? (
            <>
                {filteredData?.map((func, index) => (
                    <FunctionBox key={index} funcDetails={func} integrated={integrated} />
                ))}
            </>
        ) : filter === 'installed' ? (
            <>
                {filteredData
                    .filter((func) => func?.is_installed)
                    ?.map((func, index) => (
                        <FunctionBox key={index} funcDetails={func} integrated={integrated} />
                    ))}
            </>
        ) : (
            <>
                {filteredData
                    .filter((func) => !func?.is_installed)
                    ?.map((func, index) => (
                        <FunctionBox key={index} funcDetails={func} integrated={integrated} />
                    ))}
            </>
        );

    const drawCollapse = () => {
        if (isCloud() && tabValue === 'Private' && !integrated) return <IntegrateFunction onClick={() => setIsFunctionsGuideOpen(true)} />;
        const noFunctionsContent = filteredData?.length === 0 ? renderNoFunctionsFound() : null;
        const installedFunctionBoxesContent = filteredData?.length !== 0 ? <div className="cards-wrapper">{renderFunctionBoxes('installed')}</div> : null;
        const otherFunctionBoxesContent = filteredData?.length !== 0 ? <div className="cards-wrapper">{renderFunctionBoxes('other')}</div> : null;

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
                            <Panel header={<div className="panel-header">Installed</div>} key={1}>
                                <div>{installedFunctionBoxesContent || noFunctionsContent}</div>
                            </Panel>
                        </Collapse>
                        <Collapse defaultActiveKey={['2']} accordion={true} expandIcon={({ isActive }) => <ExpandIcon isActive={isActive} />} ghost>
                            <Panel header={<div className="panel-header">Other</div>} key={2}>
                                <div>{otherFunctionBoxesContent || noFunctionsContent}</div>
                            </Panel>
                        </Collapse>
                    </>
                )}
            </div>
        );
    };
    const renderContent = () => {
        const noFunctionsContent = filteredData?.length === 0 ? renderNoFunctionsFound() : null;
        return drawCollapse() || noFunctionsContent;
    };

    return (
        <div className="function-container">
            <div className="header-wraper">
                <div className="main-header-wrapper">
                    <div className="header-flex-wrapper">
                        <label className="main-header-h1">
                            Functions <label className="length-list">{filteredData?.length > 0 && `(${filteredData?.length})`}</label>
                        </label>
                        {isCloud() && integrated && (
                            <>
                                <div className="integrated-wrapper">
                                    <GithubActiveConnectionIcon alt="integratedIcon" />
                                    <OverflowTip text={'Integrated with GitHub'} maxWidth={'180px'}>
                                        <span>{'Integrated with GitHub'}</span>
                                    </OverflowTip>
                                </div>
                                <Button
                                    width={'100px'}
                                    height={'34px'}
                                    placeholder={
                                        <div className="button-content">
                                            {!isLoading && <RefreshIcon alt="refreshIcon" style={{ path: { color: '#6557FF' } }} />}
                                            <span>Fetch</span>
                                        </div>
                                    }
                                    backgroundColorType={'white'}
                                    colorType="black"
                                    radiusType="circle"
                                    border={'gray-light'}
                                    isLoading={isLoading}
                                    onClick={getAllFunctions}
                                />
                            </>
                        )}
                    </div>
                    <span className="memphis-label">Serverless functions to process ingested events "on the fly"</span>
                </div>
                <div className="action-section">
                    {isCloud() && !integrated && (
                        <Button
                            width="166px"
                            height="34px"
                            placeholder="Integrate with GitHub"
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            boxShadowStyle="float"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            aria-controls="usecse-menu"
                            aria-haspopup="true"
                            onClick={() => setIsFunctionsGuideOpen(true)}
                        />
                    )}
                    {isCloud() && integrated && (
                        <Popover
                            placement="top"
                            title={
                                connectedRepos?.length > 0 && (
                                    <div
                                        className="git-repo git-refresh-title"
                                        onClick={() => {
                                            modalFlip(true);
                                            setClickedRefresh(false);
                                        }}
                                    >
                                        <label>Add repositories</label>
                                        <AddRounded className="add" fontSize="small" />
                                    </div>
                                )
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
                    )}
                    <Button
                        width="166px"
                        height="34px"
                        placeholder="Download template"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontFamily="InterSemiBold"
                        aria-controls="usecse-menu"
                        aria-haspopup="true"
                        onClick={() => setIsCloneModalOpen(true)}
                    />
                </div>
            </div>
            <div className="function-tabs">
                <CustomTabs tabs={TABS} tabValue={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
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
                width="435px"
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
