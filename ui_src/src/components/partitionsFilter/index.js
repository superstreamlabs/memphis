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

import React, { createContext, useContext, useEffect, useReducer, useState } from 'react';
import { StringCodec, JSONCodec } from 'nats.ws';
import { Divider, Popover } from 'antd';

import { filterType, labelType, CircleLetterColor } from '../../const/globalConst';
import searchIcon from '../../assets/images/searchIcon.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import partitionIcon from '../../assets/images/partitionIcon.svg';
import { httpRequest } from '../../services/http';
import Button from '../button';
import CollapseArrow from '../../assets/images/collapseArrow.svg';
import { StationStoreContext } from '../../domain/stationOverview';

const PartitionsFilter = ({ partitions_number, height }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [isOpen, setIsOpen] = useState(false);
    const [selectedPartition, setSelectedPartition] = useState(0);

    const handleApply = (i) => {
        setSelectedPartition(i);
        console.log(i);
        stationDispatch({ type: 'SET_STATION_PARTITION', payload: i });
        setIsOpen(false);
    };

    const handleOpenChange = () => {
        setIsOpen(!isOpen);
    };

    const getItems = () => {
        let elements = [];
        for (let i = 0; i <= partitions_number; i++) {
            elements.push(
                <div className="el" key={i} onClick={() => handleApply(i)}>
                    <span>
                        <img src={partitionIcon} alt="PartitionIcon" /> {i == 0 ? 'All partitions' : `Partition ${i}`}
                    </span>
                </div>
            );
        }
        return elements;
    };
    const getContent = () => {
        return (
            <div>
                <div className="filter-partitions-title">
                    <p>Filter</p>
                    <Divider />
                </div>
                <div className="filter-partitions-container">{getItems()}</div>
            </div>
        );
    };

    <div>{partitions_number}</div>;

    return (
        <Popover placement="bottomLeft" content={getContent()} trigger="click" onOpenChange={handleOpenChange} open={isOpen}>
            <Button
                className="modal-btn"
                width="200px"
                height={height}
                placeholder={
                    <div className="filter-partition-container">
                        <div>
                            <label className="filter-title">Filter: </label>
                            {selectedPartition == 0 ? `All partitions` : `Partition ${selectedPartition}`}
                        </div>
                        <img src={CollapseArrow} alt="CollapseArrow" />
                    </div>
                }
                colorType="black"
                radiusType="circle"
                backgroundColorType="white"
                fontSize="14px"
                fontWeight="bold"
                boxShadowStyle="float"
                onClick={() => {}}
            />
        </Popover>
    );
};
export default PartitionsFilter;
