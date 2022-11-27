// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

import './style.scss';

import React, { useEffect, useState } from 'react';
import MultiCollapse from '../../../stationOverview/stationObservabilty/components/multiCollapse';

const ConsumerGroup = ({ header, details, cgMembers }) => {
    const [consumers, setConsumers] = useState([]);
    useEffect(() => {
        cgMembers.map((row, index) => {
            let consumer = {
                name: row.name,
                is_active: row.is_active,
                is_deleted: row.is_deleted,
                details: [
                    {
                        name: 'User',
                        value: row.created_by_user
                    },
                    {
                        name: 'IP',
                        value: row.client_address
                    }
                ]
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
                        return (
                            <content is="x3d" key={index}>
                                <p>{row.name}</p>
                                <span>{row.value}</span>
                            </content>
                        );
                    })}
                </div>
                <div className="consumers">
                    <MultiCollapse data={consumers} />
                </div>
            </div>
        </div>
    );
};
export default ConsumerGroup;
