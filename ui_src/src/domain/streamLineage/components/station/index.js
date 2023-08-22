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

import React from 'react';
import { BiSolidEnvelope } from 'react-icons/bi';
import { PiWarningFill } from 'react-icons/pi';

import redirectWhite from '../../../../assets/images/redirectWhite.svg';
import { useHistory } from 'react-router-dom/cjs/react-router-dom.min';
import pathDomains from '../../../../router';

const Station = ({ stationName, dls_messages, total_messages, schema_name }) => {
    const history = useHistory();

    const goToStation = () => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };
    return (
        <div className="station-graph-wrapper" onClick={() => goToStation()}>
            <div className="yellow-background" />
            <div className="station-details">
                <img src={redirectWhite} alt="redirectWhite" />
                <div className="station-name">{stationName}</div>
                <div className="station-messages">
                    <div className="icon-wrapper">
                        <BiSolidEnvelope />
                    </div>
                    <div className="station-messages-title">Messages</div>
                    <div className="station-messages-count">{total_messages}</div>
                </div>
                <div className="station-messages">
                    <div className="icon-wrapper">
                        <PiWarningFill />
                    </div>
                    <div className="station-messages-title">DLS messages</div>
                    <div className="station-messages-count">{dls_messages}</div>
                </div>
                {schema_name !== '' && (
                    <div className="station-messages schema-attached">
                        <div className="schema-attached-title">Schema attached</div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default Station;
