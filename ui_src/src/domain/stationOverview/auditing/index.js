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

import React, { useState } from 'react';

import CustomTabs from '../../../components/Tabs';
import GenericList from './genericList';
import { Divider } from 'antd';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';

const auditColumns = [
    {
        key: '1',
        title: 'Message',
        width: '300px'
    },
    {
        key: '2',
        title: 'User',
        width: '200px'
    },
    {
        key: '3',
        title: 'Date',
        width: '200px'
    }
];

const Auditing = () => {
    const [tabValue, setTabValue] = useState(0);
    const tabs = ['Audit'];

    const handleChangeMenuItem = (_, newValue) => {
        setTabValue(newValue);
    };

    return (
        // <div className="auditing-container">
        //     {tabValue === 0 && <p className="audit-hint">*last 30 days</p>}
        //     <CustomTabs value={tabValue} onChange={handleChangeMenuItem} tabs={tabs}></CustomTabs>
        //     <Divider />
        //     <div className="auditing-body">{tabValue === 0 && <GenericList tab={tabValue} columns={auditColumns} />}</div>
        // </div>
        <GenericList tab={tabValue} columns={auditColumns} />
    );
};

export default Auditing;
