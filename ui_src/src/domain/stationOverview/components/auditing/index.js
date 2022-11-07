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

import React, { useState } from 'react';

import GenericList from './genericList';

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

    return <GenericList tab={tabValue} columns={auditColumns} />;
};

export default Auditing;
