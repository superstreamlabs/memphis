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

import { ChevronRightOutlined } from '@material-ui/icons';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';

import React, { useState } from 'react';
import { Link } from 'react-router-dom';

import redirectIcon from '../../../../assets/images/redirectIcon.svg';
import docsPurple from '../../../../assets/images/docsPurple.svg';
import Copy from '../../../copy';

const CloudeProviders = ({ steps }) => {
    const [cloudSelected, setcloudSelected] = useState('aws');
    const [docsLink, setDocsLink] = useState(steps?.clouds[0].docsLink);

    const handleSelectedCloud = (name, link) => {
        setcloudSelected(name);
        setDocsLink(link);
    };

    return (
        <div className="installation-command">
            <div className="steps">
                <div className="step-wrapper">
                    <p className="step-title">Choose your cloud:</p>
                    <div className="img-wrapper">
                        {steps?.clouds &&
                            steps?.clouds.map((value) => {
                                return (
                                    <div
                                        key={value.name}
                                        className={cloudSelected === value.name ? 'img-cloud selected' : 'img-cloud'}
                                        onClick={() => handleSelectedCloud(value.name, value.docsLink)}
                                    >
                                        {value.src}
                                    </div>
                                );
                            })}
                    </div>
                </div>
                {steps[cloudSelected]?.map((value, key) => {
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
            <div className="links">
                <Link to={{ pathname: docsLink }} target="_blank">
                    <div className="link-wrapper">
                        <img width={25} height={22} src={docsPurple} alt="docsPurple" />
                        <p>Link to docs</p>
                        <ChevronRightOutlined />
                    </div>
                </Link>
            </div>
        </div>
    );
};

export default CloudeProviders;
