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

import React, { useContext, useState, useEffect } from 'react';
import Button from '../../../components/button';
import { ReactComponent as LogoTexeMemphis } from '../../../assets/images/logoTexeMemphis.svg';
import { ReactComponent as RedirectWhiteIcon } from '../../../assets/images/exportWhite.svg';
import { ReactComponent as DocumentIcon } from '../../../assets/images/documentGroupIcon.svg';
import { ReactComponent as DisordIcon } from '../../../assets/images/discordGroupIcon.svg';
import { ReactComponent as WindowIcon } from '../../../assets/images/windowGroupIcon.svg';

function SoftwareUpates({}) {
    const [version, setVersion] = useState({
        currentVersion: 'v0.4.3 - Beta',
        currentVersionURL: 'https://docs.memphis.dev/memphis/release-notes/releases/v0.4.3-beta',
        isUpdateAvailable: true
    });
    return (
        <div className="softwate-updates-container">
            <div className="rows">
                <div className="item-component">
                    <div className="title-component">
                        <div className="versions">
                            <LogoTexeMemphis alt="Memphis logo" />
                            <label className="curr-version">{version.currentVersion}</label>
                            {version.isUpdateAvailable && <div className="red-dot" />}
                        </div>
                        <Button
                            width="200px"
                            height="36px"
                            placeholder={
                                <span className="">
                                    <label>View Change log </label>
                                    <RedirectWhiteIcon alt="redirect" />
                                </span>
                            }
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType={'purple'}
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => {
                                window.open('https://docs.memphis.dev/memphis/release-notes/releases', '_blank');
                            }}
                        />
                    </div>
                </div>
                <div className="statistics">
                    <div className="item-component wrapper">
                        <label className="title">Amount of brokers</label>
                        <label className="numbers">600</label>
                    </div>
                    <div className="item-component wrapper">
                        <label className="title">total stations</label>
                        <label className="numbers">600</label>
                    </div>
                    <div className="item-component wrapper">
                        <label className="title">total users</label>
                        <label className="numbers">600</label>
                    </div>
                    <div className="item-component wrapper">
                        <label className="title">total schemas</label>
                        <label className="numbers">600</label>
                    </div>
                </div>
                <div className="charts">
                    <div className="item-component">
                        <DocumentIcon />
                        <p>Read Our documentation</p>
                        <span>
                            Read our documentation to learn more about <a href="https://docs.memphis.dev/memphis/getting-started/readme"> Memphis.dev</a>
                        </span>
                    </div>
                    <div className="item-component">
                        <DisordIcon />
                        <p>Join our Discord</p>
                        <span>
                            Find <a href="https://memphis.dev/open-source-support-bundle/">Memphis.dev's</a> Open-Source contributors and maintainers here
                        </span>
                    </div>
                    <div className="item-component">
                        <WindowIcon />
                        <p>Open a service request</p>
                        <span>Lorem ipsum dolor sit amet, consectetur adipiscing elit. </span>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default SoftwareUpates;
