// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { useState, useEffect, useContext } from 'react';
import { LoadingOutlined } from '@ant-design/icons';
import { Spin, Badge } from 'antd';
import CustomTabs from '../../../../../components/Tabs';
import FunctionBox from '../../../../functions/components/functionBox';
import FunctionDetails from '../../../../functions/components/functionDetails';
import { getFunctionsTabs } from '../../../../../services/valueConvertor';
import SearchInput from '../../../../../components/searchInput';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import FunctionsApplyModal from '../functionsApplyModal';
import FunctionInputsModal from '../functionInputsModal';
import Modal from '../../../../../components/modal';
import { ReactComponent as SearchIcon } from '../../../../../assets/images/searchIcon.svg';
import { ReactComponent as CheckShieldIcon } from '../../../../../assets/images/checkShieldIcon.svg';
import { ReactComponent as FunctionsModalIcon } from '../../../../../assets/images/vueSaxIcon.svg';
import { ReactComponent as LockIcon } from '../../../../../assets/images/lockIcon.svg';
import { ReactComponent as RefreshIcon } from '../../../../../assets/images/refresh.svg';
import { StationStoreContext } from '../../../';
import { SyncOutlined } from '@ant-design/icons';
import { showMessages } from '../../../../../services/genericServices';
import { OWNER } from '../../../../../const/globalConst';
import TooltipComponent from '../../../../../components/tooltip/tooltip';

const FunctionsModal = ({ open, clickOutside, applyFunction, referredFunction }) => {
    const [functionList, setFunctionList] = useState([]);
    const [isIntegrated, setIsIntegrated] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [refreshIndeicator, setRefreshIndicator] = useState(false);
    const [tabValue, setTabValue] = useState('all');
    const [searchInput, setSearchInput] = useState('');
    const [isApplyModalOpen, setIsApplyModalOpen] = useState(false);
    const [isInputsModalOpen, setIsInputsModalOpen] = useState(false);
    const [filteredData, setFilteredData] = useState([]);
    const [clickedFunction, setClickedFunction] = useState(null);
    const [selectedFunction, setSelectedFunction] = useState(null);
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [tabsCounter, setTabsCounter] = useState([0, 0, 0]);

    const TABS = getFunctionsTabs();

    useEffect(() => {
        getAllFunctions();
        setClickedFunction(null);
    }, [open]);

    useEffect(() => {
        const memphisCount = filteredData?.filter((func) => func?.owner === OWNER)?.length;
        const privateCount = filteredData?.filter((func) => func?.owner !== OWNER)?.length;
        setTabsCounter([memphisCount + privateCount, memphisCount, privateCount]);
    }, [filteredData]);

    useEffect(() => {
        let shouldRefresh = false;
        shouldRefresh = functionList?.some((func) => {
            return func?.installed_in_progress;
        });
        setRefreshIndicator(shouldRefresh);
    }, [functionList]);

    useEffect(() => {
        let result = functionList;
        if (tabValue === 'Private') {
            result = result?.filter((func) => func?.owner !== OWNER);
        } else if (tabValue === 'Memphis') {
            result = result?.filter((func) => func?.owner === OWNER);
        }
        if (searchInput.length > 0) {
            result = result?.filter(
                (func) =>
                    func?.function_name?.toLowerCase()?.includes(searchInput?.toLowerCase()) || func?.description?.toLowerCase()?.includes(searchInput.toLowerCase())
            );
        }
        setFilteredData(result);
    }, [tabValue, searchInput, functionList]);

    const getAllFunctions = async () => {
        setIsLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_FUNCTIONS);
            setIsIntegrated(data?.scm_integrated);

            let updatedData = { ...data };

            const installed = updatedData?.installed
                ?.map((func, index) => {
                    if (func?.owner === OWNER) {
                        func.stars = Math.random() + 4;
                        func.rates = Math.floor(Math.random() * (80 - 50 + 1)) + 50;
                        func.forks = Math.floor(Math.random() * (100 - 80 + 1)) + 80;
                    }
                    return func;
                })
                ?.sort((a, b) => (a.function_name > b.function_name ? 1 : -1));

            const other = updatedData?.other
                ?.filter((func) => func?.is_valid)
                ?.map((func, index) => {
                    if (func?.owner === OWNER) {
                        func.stars = Math.random() + 4;
                        func.rates = Math.floor(Math.random() * (80 - 50 + 1)) + 50;
                        func.forks = Math.floor(Math.random() * (100 - 80 + 1)) + 80;
                    }
                    return func;
                })
                ?.sort((a, b) => (a.function_name > b.function_name ? 1 : -1));

            setFunctionList([...installed, ...other] || []);
            setTimeout(() => {
                setIsLoading(false);
            }, 500);
        } catch (error) {
            setIsLoading(false);
        }
    };
    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const onFunctionApply = async (selectedFunction, ordering) => {
        let inputsObject = {};
        selectedFunction?.inputs?.forEach((item) => {
            inputsObject[item.name] = item.value;
        });
        const requestBody = {
            function_id: selectedFunction?.id,
            visible_step: stationState?.stationFunctions?.functions?.length + 1,
            ordering_matter: ordering,
            activate: true,
            inputs: inputsObject
        };
        applyFunction(requestBody);
    };

    const handleUnInstall = async (clickedFunction) => {
        const bodyRequest = {
            function_name: clickedFunction?.function_name,
            repo: clickedFunction?.repo,
            owner: clickedFunction?.owner,
            branch: clickedFunction?.branch,
            scm_type: clickedFunction?.scm,
            compute_engine: clickedFunction?.compute_engine
        };
        try {
            await httpRequest('DELETE', ApiEndpoints.UNINSTALL_FUNCTION, bodyRequest);
            getAllFunctions();
            setClickedFunction(null);
        } catch (e) {
        } finally {
        }
    };

    const handleInstall = async (clickedFunction) => {
        const bodyRequest = {
            function_name: clickedFunction?.function_name,
            repo: clickedFunction?.repo,
            owner: clickedFunction?.owner,
            branch: clickedFunction?.branch,
            scm_type: clickedFunction?.scm,
            by_memphis: clickedFunction?.by_memphis
        };
        try {
            await httpRequest('POST', ApiEndpoints.INSTALL_FUNCTION, bodyRequest);
            showMessages('success', `We are ${clickedFunction?.updates_available ? 'updating' : 'installing'} the function for you. We will let you know once its done`);
            getAllFunctions();
        } catch (e) {
            return;
        }
    };

    const handleInputsChange = (inputs) => {
        const newFunction = { ...selectedFunction };
        newFunction.inputs = inputs;
        setSelectedFunction(newFunction);
    };

    const antIcon = (
        <LoadingOutlined
            style={{
                fontSize: 24,
                color: '#5A4FE5'
            }}
            spin
        />
    );
    return (
        <>
            <Modal open={open} clickOutside={clickOutside} displayButtons={false} className="ms-function-details-modal" height="95vh" width="1200px">
                <div className="function-modal-container">
                    {clickedFunction ? (
                        <FunctionDetails
                            selectedFunction={clickedFunction}
                            integrated={false}
                            onBackToFunction={() => {
                                setClickedFunction(null);
                            }}
                            clickApply={() => {
                                setSelectedFunction(clickedFunction);
                                clickedFunction?.inputs?.length > 0
                                    ? setIsInputsModalOpen(true)
                                    : stationState?.stationFunctions?.functions?.length === 0
                                    ? setIsApplyModalOpen(true)
                                    : onFunctionApply(clickedFunction);
                            }}
                            handleUnInstall={() => handleUnInstall(clickedFunction)}
                            handleInstall={() => handleInstall(clickedFunction)}
                        />
                    ) : (
                        <>
                            <div className="fdm-header modal-header">
                                <div className="header-img-container">
                                    <FunctionsModalIcon />
                                </div>
                                <p>Functions</p>
                                <span className="title-section">
                                    <label>Say Goodbye to Manual Business Logic! Use Functions Instead of Building Complicated Clients.</label>
                                    <span className="update-refresh">
                                        {refreshIndeicator && <Badge dot />}
                                        <div className="refresh-btn">
                                            {isLoading ? (
                                                <Spin indicator={<SyncOutlined style={{ color: '#6557FF', fontSize: '16px' }} spin />} />
                                            ) : (
                                                <RefreshIcon alt="refreshIcon" style={{ path: { color: '#6557FF' } }} onClick={getAllFunctions} />
                                            )}
                                        </div>
                                    </span>
                                </span>
                            </div>
                            <div className="fdm-body">
                                <CustomTabs tabs={TABS} tabValue={tabValue} onChange={(tabValue) => setTabValue(tabValue)} tabsCounter={tabsCounter} />
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
                                    className="mrb-15"
                                />
                                {isLoading ? (
                                    <div className="loader-container">
                                        <Spin indicator={antIcon} />
                                    </div>
                                ) : (
                                    <div className="functions-list">
                                        {filteredData?.map((functionItem, index) => (
                                            <FunctionBox
                                                key={`func-index-${index}`}
                                                funcDetails={functionItem}
                                                integrated={isIntegrated}
                                                referredFunction={referredFunction}
                                                isTagsOn={false}
                                                onApply={() => {
                                                    setSelectedFunction(functionItem);
                                                    functionItem?.inputs?.length > 0
                                                        ? setIsInputsModalOpen(true)
                                                        : stationState?.stationFunctions?.functions?.length === 0
                                                        ? setIsApplyModalOpen(true)
                                                        : onFunctionApply(functionItem);
                                                }}
                                                onClick={() => {
                                                    setClickedFunction(functionItem);
                                                }}
                                                startInstallation={() => getAllFunctions()}
                                            />
                                        ))}
                                    </div>
                                )}
                            </div>
                        </>
                    )}
                </div>
            </Modal>
            <FunctionInputsModal
                open={isInputsModalOpen}
                clickOutside={() => setIsInputsModalOpen(false)}
                rBtnClick={() => {
                    stationState?.stationFunctions?.functions?.length === 0 ? setIsApplyModalOpen(true) : onFunctionApply(selectedFunction);
                    setIsInputsModalOpen(false);
                }}
                rBtnText={stationState?.stationFunctions?.functions?.length === 0 ? 'Next' : 'Apply'}
                clickedFunction={selectedFunction}
            />
            <Modal
                width={'400px'}
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <CheckShieldIcon />
                        </div>
                        <p>Events Orderding</p>
                        <label>Should Functions maintain the initial order of the events as they received?</label>
                    </div>
                }
                open={isApplyModalOpen}
                clickOutside={() => setIsApplyModalOpen(false)}
                displayButtons={false}
            >
                <FunctionsApplyModal
                    onCancel={() => setIsApplyModalOpen(false)}
                    onApply={(e) => {
                        onFunctionApply(selectedFunction, e);
                        setIsApplyModalOpen(false);
                    }}
                    successText={'Next'}
                />
            </Modal>
        </>
    );
};

export default FunctionsModal;
