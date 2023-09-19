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

import React, { useContext, useState } from 'react';
import { Divider, Popover } from 'antd';
import { ReactComponent as PartitionIcon } from '../../assets/images/partitionIcon.svg';
import { ReactComponent as CollapseArrowIcon } from '../../assets/images/collapseArrow.svg';
import { StationStoreContext } from '../../domain/stationOverview';

const PartitionsFilter = ({ partitions_number }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);
    const [isOpen, setIsOpen] = useState(false);
    const [selectedPartition, setSelectedPartition] = useState(-1);

    const handleApply = (i) => {
        setSelectedPartition(i === 0 ? -1 : i);
        stationDispatch({ type: 'SET_STATION_PARTITION', payload: i === 0 ? -1 : i });
        setIsOpen(false);
    };

    const handleOpenChange = () => {
        setIsOpen(!isOpen);
    };

    const getItems = () => {
        let elements = [];
        for (let i = 0; i <= partitions_number; i++) {
            elements.push(
                <div className="partition-item" key={i} onClick={() => handleApply(i)}>
                    <span>
                        <PartitionIcon alt="PartitionIcon" />
                        {i == 0 ? 'All partitions' : `Partition ${i}`}
                    </span>
                </div>
            );
        }
        return elements;
    };
    const getContent = () => {
        return <div className="filter-partitions-container">{getItems()}</div>;
    };

    return (
        <Popover placement="bottomLeft" content={getContent()} trigger="click" onOpenChange={handleOpenChange} open={isOpen}>
            <div className="filter-partition-btn">
                <div className="filter-partition-container">
                    <PartitionIcon alt="PartitionIcon" className="partition-icon" />
                    <div>{selectedPartition == -1 ? `All partitions` : `Partition ${selectedPartition}`}</div>
                    <CollapseArrowIcon alt="CollapseArrow" />
                </div>
            </div>
        </Popover>
    );
};
export default PartitionsFilter;
