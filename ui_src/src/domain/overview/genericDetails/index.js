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

import React, { useContext } from 'react';

import liveMessagesIcon from '../../../assets/images/liveMessagesIcon.svg';
import stationActionIcon from '../../../assets/images/stationActionIcon.svg';
import { Context } from '../../../hooks/store';
import { numberWithCommas } from '../../../services/valueConvertor';

const GenericDetails = () => {
    const [state, dispatch] = useContext(Context);

    return (
        <div className="generic-container">
            <div className="overview-wrapper data-box">
                <div className="icon-wrapper sta-act">
                    <img src={stationActionIcon} width={35} height={27} alt="stationActionIcon" />
                </div>
                <div className="data-wrapper">
                    <span>Total stations</span>
                    <p>{numberWithCommas(state?.monitor_data?.total_stations)}</p>
                </div>
            </div>
            <div className="overview-wrapper data-box">
                <div className="icon-wrapper lve-msg">
                    <img src={liveMessagesIcon} width={35} height={26} alt="liveMessagesIcon" />
                </div>
                <div className="data-wrapper">
                    <span>Total messages</span>
                    <p> {numberWithCommas(state?.monitor_data?.total_messages)}</p>
                </div>
            </div>
        </div>
    );
};

export default GenericDetails;
