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

import React, { useContext } from 'react';

import stationIdleIcon from '../../../assets/images/stationIdleIcon.svg';
import liveMessagesIcon from '../../../assets/images/liveMessagesIcon.svg';
import stationActionIcon from '../../../assets/images/stationActionIcon.svg';
import comingSoonBox from '../../../assets/images/comingSoonBox.svg';
import { Context } from '../../../hooks/store';
import { numberWithCommas } from '../../../services/valueConvertor';

const GenericDetails = () => {
    const [state, dispatch] = useContext(Context);

    return (
        <div className="generic-container">
            <div className="overview-wrapper data-box">
                {/* <div className="coming-soon-small">
                    <img src={comingSoonBox} width={25} height={45} />
                    <p>Coming soon</p>
                </div> */}
                <div className="icon-wrapper sta-act">
                    <img src={stationActionIcon} width={35} height={27} alt="stationActionIcon" />
                </div>
                <div className="data-wrapper">
                    <span>Total stations</span>
                    <p>{numberWithCommas(state?.monitor_data?.total_stations)}</p>
                </div>
            </div>
            <div className="overview-wrapper data-box">
                {/* <div className="coming-soon-small">
                    <img src={comingSoonBox} width={25} height={45} />
                    <p>Coming soon</p>
                </div> */}
                <div className="icon-wrapper lve-msg">
                    <img src={liveMessagesIcon} width={35} height={26} alt="liveMessagesIcon" />
                </div>
                <div className="data-wrapper">
                    <span>Total messages</span>
                    <p> {numberWithCommas(state?.monitor_data?.total_messages)}</p>
                </div>
            </div>
            {/* <div className="overview-wrapper data-box">
                <div className="coming-soon-small">
                    <img src={comingSoonBox} width={25} height={45} />
                    <p>Coming soon</p>
                </div>
                <div className="icon-wrapper sta-idl">
                    <img src={stationIdleIcon} width={35} height={27} alt="stationIdleIcon" />
                </div>
                <div className="data-wrapper">
                    <span>Total stations</span>
                    <p>
                        3 <span>on idle</span>
                    </p>
                </div>
            </div> */}
        </div>
    );
};

export default GenericDetails;
