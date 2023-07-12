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
import StatusIndication from '../../../../components/indication';

const ConsumerGroup = ({ header, details, cgMembers }) => {
    const [consumers, setConsumers] = useState([]);
    useEffect(() => {
        cgMembers?.map((row, index) => {
            let consumer = {
                name: row.name,
                is_active: row.is_active,
                is_deleted: row.is_deleted,
            };
            setConsumers([consumer]);
        });
    }, [cgMembers]);

    return (
        <div className="consumer-group">
            <header is="x3d">
                <p>CG - {header}</p>
            </header>
            <div className="content-wrapper">
                <div className="details">
                    <p className="title">Details</p>
                    {details?.map((row, index) => {
                        if (row.value !== '-1') {
                            return (
                                <content is="x3d" key={index}>
                                    <p>{row.name}</p>
                                    <span>{row.value}</span>
                                </content>
                            );
                        }
                    })}
                </div>
                <div className="consumers">
                {consumers?.map((row, index) => {
                            return (
                                <div className="consumer" key={index}>
                                        <p className="title">
                                            {row.name}
                                        </p>
                                        <status is="x3d">
                                            <StatusIndication is_active={row.is_active} is_deleted={false} />
                                        </status>
                                </div>
                            );
                    })}
                </div>
            </div>
        </div>
    );
};
export default ConsumerGroup;
