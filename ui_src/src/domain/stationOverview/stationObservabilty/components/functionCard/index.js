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
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import { convertLongNumbers } from '../../../../../services/valueConvertor';
import { StationStoreContext } from '../../../';

export default function FunctionCard({
    onClick,
    stationName,
    partiotionNumber,
    onDeleteFunction,
    functionItem,
    isGeneralView,
    isDeactive = false,
    selected,
    changeActivition,
    requestInfo
}) {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [isActive, setIsActive] = useState(true);
    const [popoverFunctionContextMenu, setPopoverFunctionContextMenu] = useState(false);

    useEffect(() => {
        setIsActive(!isDeactive);
    }, [isDeactive]);

    const functionContextMenuStyles = {
        borderRadius: '8px',
        paddingTop: '5px',
        paddingBottom: '5px',
        marginBottom: '10px',
        width: '150px'
    };

    const getFunctionsOverview = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_FUNCTIONS_OVERVIEW}?station_name=${stationState?.stationMetaData?.name}`);
            stationDispatch({ type: 'SET_STATION_FUNCTIONS', payload: data });
        } catch (e) {
            return;
        }
    };

    const handleDeactivate = async () => {
        const bodyRequest = {
            function_id: functionItem?.installed_id,
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
            function_id: functionItem?.installed_id,
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
            function_id: functionItem?.installed_id,
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
            <div
                className="item-wrap"
                style={{ width: 'initial' }}
                onClick={(e) => {
                    e.stopPropagation();
                    setPopoverFunctionContextMenu(false);
                    requestInfo();
                }}
            >
                <div className="item">
                    <p className="item-title">Information</p>
                </div>
            </div>
        </div>
    );

    return (
        <div className={`ms-function-card ${!functionItem?.activated ? 'deactivated' : undefined}`} onClick={() => isActive && onClick && onClick()}>
            <div className="ms-function-card-badge-left">
                <FunctionProcessingWarningIcon />
                {convertLongNumbers(functionItem?.pending_messages || 0)}
            </div>
            <div className="ms-function-card-top">
                <div className="ms-function-card-badge-top">
                    <FunctionProcessingIcon />
                    {convertLongNumbers(functionItem?.in_process_messages || 0)}
                </div>
                <div className={`ms-function-card-inner ${selected ? 'selected' : undefined}`}>
                    <div className="ms-function-card-header">
                        <div className="ms-function-card-header-title">
                            <FunctionBoxTitleIcon />
                            <div>
                                <span>{functionItem.name}</span>
                                {isGeneralView && <p>Avg. processing time : {functionItem.metrics?.average_processing_time}s</p>}
                            </div>
                        </div>
                        <div className="ms-function-card-header-action">
                            <Popover
                                overlayInnerStyle={functionContextMenuStyles}
                                placement="bottom"
                                content={functionContextMenu}
                                trigger="click"
                                onOpenChange={() => {
                                    setPopoverFunctionContextMenu(!popoverFunctionContextMenu);
                                }}
                                open={popoverFunctionContextMenu}
                            >
                                <HiEllipsisVertical size={16} onClick={(e) => e.stopPropagation()} />
                            </Popover>
                        </div>
                    </div>
                    <div className="ms-function-card-body">
                        <div className="ms-function-card-body-left">
                            <div className="ms-function-card-info-box">
                                <div className="title">Av. Processing Time</div>
                                <div className="subtitle">{functionItem.metrics?.average_processing_time}s</div>
                            </div>
                        </div>
                        <div className="ms-function-card-body-right">
                            <div className="ms-function-card-info-box">
                                <div className="title">Error Rate</div>
                                <div className="subtitle">{functionItem.metrics?.error_rate}%</div>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="ms-function-card-badge-bottom">
                    <FunctionProcessingIcon />
                    {convertLongNumbers(functionItem?.in_process_messages || 0)}
                </div>
            </div>
        </div>
    );
}
