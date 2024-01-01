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

import { KeyboardArrowUpRounded } from '@material-ui/icons';
import { useHistory } from 'react-router-dom';
import React, { useState } from 'react';
import { Badge, Divider } from 'antd';

import { ReactComponent as StationsActiveIcon } from 'assets/images/stationsIconActive.svg';
import { ReactComponent as RedirectIcon } from 'assets/images/redirectIcon.svg';

const ExpandIcon = ({ isActive }) => <KeyboardArrowUpRounded className={isActive ? 'collapse-arrow open' : 'collapse-arrow close'} />;

const StationLagCollapse = ({ station, index }) => {
    const history = useHistory();
    const [isOpen, setIsOpen] = useState(true);

    const toggleCollapse = () => {
        setIsOpen(!isOpen);
    };

    const redirectToStation = () => {
        history.push(`/stations/${station?.station_name}`);
    };

    return (
        <div className="station-lag-wrapper" key={index}>
            <div className="station-lag-header" onClick={() => redirectToStation()}>
                <div className="left">
                    <StationsActiveIcon alt="stationsIconActive" />
                    <p>{station?.station_name}</p>
                </div>
                <RedirectIcon alt="redirectIcon" width={14} />
            </div>
            <div className="station-lag-content">
                {!isOpen && (
                    <div className="station-lag-content-header" onClick={toggleCollapse}>
                        <p>Show more</p>
                        <ExpandIcon isActive={isOpen} />
                    </div>
                )}
                {isOpen && (
                    <div className="collapse-content">
                        {station?.cgs?.map((cg, index) => (
                            <React.Fragment key={index}>
                                <div className="collapse-row">
                                    <p className="station-name">{cg.cg_name}</p>
                                    <Badge className="station-badge" count={cg.num_of_delayed_msgs} overflowCount={999} />
                                </div>
                                <Divider />
                            </React.Fragment>
                        ))}
                        <div className="station-lag-content-header" onClick={toggleCollapse}>
                            <p>Show less</p>
                            <ExpandIcon isActive={isOpen} />
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default StationLagCollapse;
