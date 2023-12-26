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
import { Collapse } from 'antd';
import { ReactComponent as CollapseArrowIcon } from '../../../../../assets/images/collapseArrow.svg';
import StatusIndication from '../../../../../components/indication';

const { Panel } = Collapse;

const MultiCollapse = ({ data }) => {
    const [activeKey, setActiveKey] = useState([]);
    const [activeChiledKey, setActiveChiledKey] = useState();
    const [collapseData, setCollapseData] = useState();

    useEffect(() => {
        setCollapseData(data);
    }, [data]);

    const onChange = (key) => {
        setActiveKey(key);
    };
    const onChiledChange = (key) => {
        setActiveChiledKey(key);
    };

    return (
        <div className="custom-collapse multi">
            <Collapse ghost accordion={true} className="collapse-child" onChange={onChiledChange}>
                {collapseData?.length > 0 &&
                    collapseData?.map((row, index) => {
                        return (
                            <Panel
                                showArrow={false}
                                header={
                                    <div className="collapse-header">
                                        <p className="title">{row.name}</p>
                                        <span className="status">
                                            <StatusIndication is_active={row.is_active} is_deleted={row.is_deleted} />
                                            <CollapseArrowIcon
                                                className={Number(activeChiledKey) === index ? 'collapse-arrow open' : 'collapse-arrow close'}
                                                alt="collapse-arrow"
                                            />
                                        </span>
                                    </div>
                                }
                                key={index}
                            >
                                {row.details?.length > 0 &&
                                    row.details?.map((row, index) => {
                                        return (
                                            <div className="panel-child" key={index}>
                                                <content is="x3d" key={index}>
                                                    <p>{row.name}</p>
                                                    <span>{row.value}</span>
                                                </content>
                                            </div>
                                        );
                                    })}
                            </Panel>
                        );
                    })}
            </Collapse>
        </div>
    );
};

export default MultiCollapse;
