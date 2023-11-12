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
import { Spin } from 'antd';
import CustomTabs from '../../../../../components/Tabs';
import FunctionBox from '../../../../functions/components/functionBox';
import FunctionDetails from '../../../../functions/components/functionDetails';
import { getFunctionsTabs } from '../../../../../services/valueConvertor';
import SearchInput from '../../../../../components/searchInput';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import FunctionsApplyModal from '../functionsApplyModal';
import Modal from '../../../../../components/modal';
import { ReactComponent as SearchIcon } from '../../../../../assets/images/searchIcon.svg';
import { ReactComponent as CheckShieldIcon } from '../../../../../assets/images/checkShieldIcon.svg';
import { ReactComponent as FunctionsModalIcon } from '../../../../../assets/images/vueSaxIcon.svg';
import { StationStoreContext } from '../../../';

import { OWNER } from '../../../../../const/globalConst';

const FunctionsModal = ({ applyFunction }) => {
    const [functionList, setFunctionList] = useState([]);
    const [isIntegrated, setIsIntegrated] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [tabValue, setTabValue] = useState('all');
    const [searchInput, setSearchInput] = useState('');
    const [isApplyModalOpen, setIsApplyModalOpen] = useState(false);
    const [filteredData, setFilteredData] = useState([]);
    const [clickedFunction, setClickedFunction] = useState(null);
    const [selectedFunction, setSelectedFunction] = useState(null);
    const [stationState, stationDispatch] = useContext(StationStoreContext);

    useEffect(() => {
        getAllFunctions();
    }, []);

    useEffect(() => {
        let result = functionList;
        if (tabValue === 'Private') {
            result = result.filter((func) => func?.owner !== OWNER && func?.is_valid);
        } else if (tabValue === 'Memphis') {
            result = result.filter((func) => func?.owner === OWNER && func?.is_valid);
        } else {
            result = result.filter((func) => func?.is_valid);
        }
        if (searchInput.length > 0) {
            result = result.filter(
                (func) =>
                    (func?.function_name?.toLowerCase()?.includes(searchInput?.toLowerCase()) && func?.is_valid) ||
                    (func?.description?.toLowerCase()?.includes(searchInput.toLowerCase()) && func?.is_valid)
            );
        }
        setFilteredData(result);
    }, [tabValue, searchInput, functionList]);

    const getAllFunctions = async () => {
        setIsLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_FUNCTIONS);
            setIsIntegrated(data?.scm_integrated);
            setFunctionList([...data?.installed, ...data?.other] || []);
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
        const requestBodey = {
            function_id: selectedFunction?.id,
            visible_step: stationState?.stationFunctions?.functions?.length + 1,
            ordering_matter: ordering,
            activate: true
        };
        applyFunction(requestBodey);
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

    const TABS = getFunctionsTabs();

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
            {clickedFunction ? (
                <FunctionDetails
                    selectedFunction={clickedFunction}
                    integrated={false}
                    installed={true}
                    onBackToFunction={() => {
                        setClickedFunction(null);
                    }}
                    clickApply={() => {
                        setSelectedFunction(clickedFunction);
                        stationState?.stationFunctions?.functions?.length === 0 ? setIsApplyModalOpen(true) : onFunctionApply(clickedFunction);
                    }}
                    handleUnInstall={() => handleUnInstall(clickedFunction)}
                />
            ) : (
                <>
                    <div className="fdm-header modal-header">
                        <div className="header-img-container">
                            <FunctionsModalIcon />
                        </div>
                        <p>Functions</p>
                        <label>Say Goodbye to Manual Business Logic! Use Functions Instead of Building Complicated Clients.</label>
                    </div>
                    <div className="fdm-body">
                        <CustomTabs tabs={TABS} tabValue={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
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
                                        key={index}
                                        funcDetails={functionItem}
                                        integrated={isIntegrated}
                                        isTagsOn={false}
                                        installed={true}
                                        onApply={() => {
                                            setSelectedFunction(functionItem);
                                            stationState?.stationFunctions?.functions?.length === 0 ? setIsApplyModalOpen(true) : onFunctionApply(functionItem);
                                        }}
                                        onClick={() => {
                                            setClickedFunction(functionItem);
                                        }}
                                    />
                                ))}
                            </div>
                        )}
                    </div>
                </>
            )}

            <Modal
                width={'403px'}
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
                    successText={'Apply'}
                />
            </Modal>
        </>
    );
};

export default FunctionsModal;
