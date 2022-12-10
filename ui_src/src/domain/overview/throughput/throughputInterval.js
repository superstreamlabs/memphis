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
import { useHistory } from 'react-router-dom';
import { KeyboardArrowRightRounded } from '@material-ui/icons';

import { numberWithCommas, parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import Button from '../../../components/button';
import Filter from '../../../components/filter';
import NoStations from '../../../assets/images/noStations.svg';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';
import { Popover } from 'antd';
import DatePickerComponent from '../../../components/datePicker';

const ThroughputInterval = ({ createStationTrigger }) => {
    const [s, ss] = useState(Context);

    const content = (
        <div className="throughput-interval-containter">
            <div className="custom">
                <p>Custom</p>
                <p>Choose custom time interval.</p>
                <label>From</label>
                <DatePickerComponent />
                <label>To</label>
                <DatePickerComponent />
                <Button
                    className="modal-btn"
                    width="190px"
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
            <div className="fixed"></div>
        </div>
    );

    return <Popover placement="bottomRight" content={content} trigger="click" open={true} />;
};

export default ThroughputInterval;
