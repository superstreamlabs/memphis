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

import React, { useContext } from 'react';
import { Context } from '../../../hooks/store';
import TotalMsg from '../../../assets/images/TotalMessages.svg';
import TotalPoison from '../../../assets/images/DeadLetteredMessages.svg';
import TotalStations from '../../../assets/images/TotalStations.svg';
import { Divider } from 'antd';

const GenericDetails = () => {
    const [state, dispatch] = useContext(Context);

    return (
        <div className="overview-components-wrapper">
            <div className="generic-details-container">
                <div className="data-box">
                    <img src={TotalStations} width={50} height={50} alt="Total stations" className="icon-wrapper" />
                    <div className="data-wrapper">
                        <span>Stations</span>
                        <p>{state?.monitor_data?.total_stations?.toLocaleString()}</p>
                    </div>
                </div>
                <Divider type="vertical" />
                <div className="data-box">
                    <img src={TotalMsg} width={50} height={50} alt="Total stations" className="icon-wrapper" />
                    <div className="data-wrapper">
                        <span>Messages</span>
                        <p>{state?.monitor_data?.total_messages?.toLocaleString()}</p>
                    </div>
                </div>
                <Divider type="vertical" />
                <div className="data-box">
                    <img src={TotalPoison} width={50} height={50} alt="Total stations" className="icon-wrapper" />
                    <div className="data-wrapper">
                        <span>Dead-letter messages</span>
                        <p>{state?.monitor_data?.total_dls_messages?.toLocaleString()}</p>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default GenericDetails;
