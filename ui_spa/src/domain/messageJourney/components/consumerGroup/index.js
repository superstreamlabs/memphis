// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
            setConsumers([...consumers, consumer]);
        });
    }, []);

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
