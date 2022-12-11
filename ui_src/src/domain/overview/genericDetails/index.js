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
import { Context } from '../../../hooks/store';
import { numberWithCommas } from '../../../services/valueConvertor';
import TotalMsg from '../../../assets/images/total_msg.svg';
import TotalPoison from '../../../assets/images/total_poison.svg';
import TotalStations from '../../../assets/images/total_stations.svg';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';

const GenericDetails = () => {
    const [state, dispatch] = useContext(Context);

    return (
        <div className="generic-container">
            <div className="overview-wrapper data-box">
                <img src={TotalStations} width={50} height={50} alt="Total stations" className="icon-wrapper" />
                <div className="data-wrapper">
                    <span>Total stations</span>
                    <p>{numberWithCommas(state?.monitor_data?.total_stations)}</p>
                </div>
            </div>
            <div className="overview-wrapper data-box">
                <img src={TotalMsg} width={50} height={50} alt="Total Messages" className="icon-wrapper" />
                <div className="data-wrapper">
                    <span>Total Messages</span>
                    <p>{numberWithCommas(state?.monitor_data?.total_messages)}</p>
                </div>
            </div>
            <div className="overview-wrapper data-box">
                <div className="coming-soon-wrapper">
                    <img src={comingSoonBox} width={40} height={60} alt="comingSoonBox" />
                </div>
                <img src={TotalPoison} width={50} height={50} alt="Total Poison messages" className="icon-wrapper" />
                <div className="data-wrapper">
                    <span>Total Poison messages</span>
                    <p></p>
                </div>
            </div>
        </div>
    );
};

export default GenericDetails;
