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
import { useHistory } from 'react-router-dom';
import debeziumIcon from '../../../../src/assets/images/debeziumIcon.svg';
import slackLogo from '../../../../src/assets/images/slackLogo.svg';
import s3Logo from '../../../../src/assets/images/s3Logo.svg';
import pathDomains from '../../../router';

const Integrations = () => {
    const history = useHistory();

    return (
        <div className="overview-components-wrapper">
            <div className="overview-integrations-container">
                <div className="overview-components-header integrations-header">
                    <p>Integrations</p>
                    <label className="link-to-page" onClick={() => history.push(`${pathDomains.administration}/integrations`)}>
                        Explore more Integration
                    </label>
                </div>
                <div className="integrations-list">
                    <div className="integration-item">
                        <img className="img-icon" src={slackLogo} alt="slack" />
                        <label className="integration-name">Slack</label>
                    </div>
                    <div className="integration-item">
                        <img className="img-icon" src={s3Logo} alt="s3" />
                        <label className="integration-name">S3 Bucket</label>
                    </div>
                    <div className="integration-item">
                        <img className="img-icon" src={debeziumIcon} alt="debezium" />
                        <label className="integration-name">Debezium and Postgres</label>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Integrations;
