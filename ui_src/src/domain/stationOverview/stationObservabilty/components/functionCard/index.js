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
import React, { useState, useEffect, useContext } from 'react';
import { Popover } from 'antd';
import { HiEllipsisVertical } from 'react-icons/hi2';
import { ReactComponent as FunctionBoxTitleIcon } from '../../../../../assets/images/functionCardIcon.svg';
import { ReactComponent as FunctionProcessingIcon } from '../../../../../assets/images/proccessingIcon.svg';
import { ReactComponent as FunctionProcessingWarningIcon } from '../../../../../assets/images/processingWarningIcon.svg';
import { IoClose } from 'react-icons/io5';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import { convertLongNumbers } from '../../../../../services/valueConvertor';
import { StationStoreContext } from '../../../';
import FunctionDetails from '../../../../functions/components/functionDetails';
import { Drawer } from 'antd';
import Tooltip from '../../../../../components/tooltip/tooltip';

export default function FunctionCard({
    onClick,
    onClickMenu,
    stationName,
    partiotionNumber,
    onDeleteFunction,
    functionItem,
    isDeactive = false,
    selected,
    changeActivition
}) {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [isActive, setIsActive] = useState(true);
    const [popoverFunctionContextMenu, setPopoverFunctionContextMenu] = useState(false);
    const [openFunctionDetails, setOpenFunctionDetails] = useState(false);
    const [selectedFunction, setSelectedFunction] = useState();

    useEffect(() => {
        setIsActive(!isDeactive);
    }, [isDeactive]);

    useEffect(() => {
        let func = functionItem;
        func.stars = Math.random() + 4;
        func.rates = Math.floor(Math.random() * (80 - 50 + 1)) + 50;
        func.forks = Math.floor(Math.random() * (100 - 80 + 1)) + 80;
        setSelectedFunction(func);
    }, [functionItem]);

    const functionContextMenuStyles = {
        borderRadius: '8px',
        paddingTop: '5px',
        paddingBottom: '5px',
        marginBottom: '10px',
        width: '150px'
    };

    const getFunctionsOverview = async () => {
        try {
            const data = await httpRequest(
                'GET',
                `${ApiEndpoints.GET_FUNCTIONS_OVERVIEW}?station_name=${stationState?.stationMetaData?.name}&partition=${stationState?.stationPartition || -1}`
            );
            stationDispatch({ type: 'SET_STATION_FUNCTIONS', payload: data });
        } catch (e) {
            return;
        }
    };

    const handleDeactivate = async () => {
        const bodyRequest = {
            function_id: functionItem?.id,
            station_name: stationName,
            partition: partiotionNumber,
            visible_step: functionItem?.visible_step
        };
        try {
            await httpRequest('POST', ApiEndpoints.DEACTIVATE_FUNCTION, bodyRequest);
            changeActivition(false);
        } catch (error) {
            return;
        }
    };

    const handleActivate = async () => {
        const bodyRequest = {
            function_id: functionItem?.id,
            station_name: stationName,
            partition: partiotionNumber,
            visible_step: functionItem?.visible_step
        };
        try {
            await httpRequest('POST', ApiEndpoints.ACTIVATE_FUNCTION, bodyRequest);
            changeActivition(true);
        } catch (error) {
            return;
        }
    };

    const handleDelete = async () => {
        const bodyRequest = {
            function_id: functionItem?.id,
            station_name: stationName,
            partition: partiotionNumber,
            visible_step: functionItem?.visible_step
        };
        try {
            await httpRequest('POST', ApiEndpoints.DETACH_FUNCTION, bodyRequest);
            getFunctionsOverview();
            onDeleteFunction();
        } catch (error) {
            return;
        }
    };

    const functionContextMenu = (
        <div className="menu-content">
            <div
                className="item-wrap"
                style={{ width: 'initial' }}
                onClick={(e) => {
                    e.stopPropagation();
                    setPopoverFunctionContextMenu(false);
                    setOpenFunctionDetails(true);
                }}
            >
                <div className="item">
                    <p className="item-title">Information</p>
                </div>
            </div>
            <div
                className="item-wrap"
                style={{ width: 'initial' }}
                onClick={(e) => {
                    e.stopPropagation();
                    setPopoverFunctionContextMenu(false);
                    functionItem?.activated ? handleDeactivate() : handleActivate();
                }}
            >
                <div className="item">
                    <p className="item-title">{functionItem?.activated ? 'Deactivate' : 'Activate'}</p>
                </div>
            </div>
            <div
                className="item-wrap"
                style={{ width: 'initial' }}
                onClick={(e) => {
                    e.stopPropagation();
                    setPopoverFunctionContextMenu(false);
                    handleDelete();
                }}
            >
                <div className="item">
                    <p className="item-title">Delete</p>
                </div>
            </div>
        </div>
    );

    return (
        <div className={`ms-function-card`} onClick={() => isActive && onClick && onClick()}>
            <Tooltip text="Awaiting messages">
                <div className={`ms-function-card-badge-left badge ${!selectedFunction?.activated ? 'deactivated' : undefined}`}>
                    <FunctionProcessingWarningIcon />
                    {convertLongNumbers(selectedFunction?.pending_messages || 0)}
                </div>
            </Tooltip>
            <div className={`ms-function-card-top`}>
                {
                    <Tooltip text="In process">
                        <div className={`ms-function-card-badge-top badge ${!selectedFunction?.activated ? 'deactivated' : undefined}`}>
                            <FunctionProcessingIcon />
                            {convertLongNumbers(functionItem?.in_process_messages || 0)}
                        </div>
                    </Tooltip>
                }
                <div className={`ms-function-card-inner ${selected ? 'selected' : undefined}`}>
                    <div className="ms-function-card-header">
                        <div className="ms-function-card-header-action">
                            <Popover
                                overlayInnerStyle={functionContextMenuStyles}
                                placement="bottom"
                                content={functionContextMenu}
                                trigger="click"
                                onOpenChange={(e) => {
                                    setPopoverFunctionContextMenu(!popoverFunctionContextMenu);
                                    e && onClickMenu();
                                }}
                                open={popoverFunctionContextMenu}
                            >
                                <HiEllipsisVertical size={16} onClick={(e) => e.stopPropagation()} />
                            </Popover>
                        </div>
                    </div>

                    <div className={`ms-function-card-title ${!selectedFunction?.activated ? 'deactivated-function' : undefined}`}>
                        <FunctionBoxTitleIcon />
                        <span>{selectedFunction?.name}</span>
                    </div>
                </div>
                <Tooltip text="Dead-letter">
                    <div className={`ms-function-card-badge-bottom badge ${!selectedFunction?.activated ? 'deactivated' : undefined}`}>
                        <FunctionProcessingIcon />
                        {convertLongNumbers(functionItem?.dls_msgs_count || 0)}
                    </div>
                </Tooltip>
            </div>
            <Drawer
                placement="right"
                size={'large'}
                className="function-drawer"
                onClose={(e) => {
                    e.stopPropagation();
                    setOpenFunctionDetails(false);
                }}
                destroyOnClose={true}
                open={openFunctionDetails}
                maskStyle={{ background: 'rgba(16, 16, 16, 0.2)' }}
                closeIcon={<IoClose style={{ color: '#D1D1D1', width: '25px', height: '25px' }} />}
            >
                <FunctionDetails selectedFunction={selectedFunction} integrated={true} stationView />
            </Drawer>
        </div>
    );
}
