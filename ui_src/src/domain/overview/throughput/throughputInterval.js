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
import { SearchOutlined } from '@ant-design/icons';
import Button from 'components/button';
import { Popover } from 'antd';
import DatePickerComponent from 'components/datePicker';
import SearchInput from 'components/searchInput';

const ThroughputInterval = ({ createStationTrigger }) => {
    const [selectInterval, setSelectInterval] = useState(0);

    const content = (
        <div className="throughput-interval-containter">
            <div className="custom" style={{ display: selectInterval !== 7 && 'none' }}>
                <div>
                    <p className="custom-header">Custom</p>
                    <p className="custom-description">Choose custom time interval.</p>
                </div>
                <div>
                    <div className="date-container">
                        <label>From</label>
                        <DatePickerComponent width="250px" />
                    </div>
                    <div className="date-container">
                        <label>To</label>
                        <DatePickerComponent width="250px" />
                    </div>
                </div>
                <Button
                    className="modal-btn"
                    width="250px"
                    height="32px"
                    placeholder="Apply Time Range"
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="14px"
                    fontWeight="bold"
                    aria-haspopup="true"
                    // onClick={() => setOpenFunctionForm(true)}
                />
            </div>
            <div className="fixed">
                <SearchInput
                    placeholder="Search quick ranges"
                    colorType="navy"
                    backgroundColorType="gray-dark"
                    width="250px"
                    height="34px"
                    borderColorType="none"
                    boxShadowsType="none"
                    borderRadiusType="semi-round"
                    iconComponent={<SearchOutlined />}
                    // onChange={handleSearch}
                    // value={searchInput}
                />
                <div className="intervals-list">
                    <p className={selectInterval === 0 && 'selected'} onClick={() => setSelectInterval(0)}>
                        Last 5 minutes
                    </p>
                    <p className={selectInterval === 1 && 'selected'} onClick={() => setSelectInterval(1)}>
                        Last 10 minutes
                    </p>
                    <p className={selectInterval === 2 && 'selected'} onClick={() => setSelectInterval(2)}>
                        Last 15 minutes
                    </p>
                    <p className={selectInterval === 3 && 'selected'} onClick={() => setSelectInterval(3)}>
                        Last 1 hrs
                    </p>
                    <p className={selectInterval === 4 && 'selected'} onClick={() => setSelectInterval(4)}>
                        Last 3 hrs
                    </p>
                    <p className={selectInterval === 5 && 'selected'} onClick={() => setSelectInterval(5)}>
                        Last 6 hrs
                    </p>
                    <p className={selectInterval === 6 && 'selected'} onClick={() => setSelectInterval(6)}>
                        Last 2 days
                    </p>
                    <p className={selectInterval === 7 && 'selected'} onClick={() => setSelectInterval(7)}>
                        Custom
                    </p>
                </div>
            </div>
        </div>
    );

    return <Popover placement="bottomRight" content={content} trigger="click" open={true} />;
};

export default ThroughputInterval;
