// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import { ChevronRightOutlined } from '@material-ui/icons';

import React from 'react';
import { Link } from 'react-router-dom';

import redirectIcon from '../../../../assets/images/redirectIcon.svg';
import videoIcon from '../../../../assets/images/videoIcon.svg';
import docsPurple from '../../../../assets/images/docsPurple.svg';
import Copy from '../../../copy';

const InstallationCommand = ({ steps, showLinks, videoLink, docsLink }) => {
    return (
        <div className="installation-command">
            {steps.length > 0 && (
                <div className="steps">
                    {steps.map((value, key) => {
                        return (
                            <div className="step-wrapper" key={key}>
                                <p className="step-title">{value.title}</p>
                                {value.command && (
                                    <div className="step-command">
                                        <span>{value.command}</span>
                                        {value.icon === 'copy' && <Copy data={value.command} key={key} />}
                                        {value.icon === 'link' && (
                                            <Link to={{ pathname: 'http://localhost:5555' }} target="_blank">
                                                <img src={redirectIcon} alt="redirectIcon" />
                                            </Link>
                                        )}
                                    </div>
                                )}
                            </div>
                        );
                    })}
                </div>
            )}
            {showLinks && (
                <div className="links">
                    <Link to={{ pathname: videoLink }} target="_blank">
                        <div className="link-wrapper">
                            <img src={videoIcon} alt="videoIcon" />
                            <p>Installation video</p>
                            <ChevronRightOutlined />
                        </div>
                    </Link>
                    <Link to={{ pathname: docsLink }} target="_blank">
                        <div className="link-wrapper">
                            <img width={25} height={22} src={docsPurple} alt="docsPurple" />
                            <p>Link to docs</p>
                            <ChevronRightOutlined />
                        </div>
                    </Link>
                </div>
            )}
        </div>
    );
};

export default InstallationCommand;
