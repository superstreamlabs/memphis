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

import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import React, { useState } from 'react';

import { INSTALLATION_GUIDE } from '../../const/installationGuide';
import Button from '../button';
import InstallationCommand from './ components/installationCommand';
import CloudeProviders from './ components/cloudeProviders';

const option = ['Kubernetes', 'Docker Compose', 'Cloud Providers'];

const Installation = ({ closeModal }) => {
    const [installationPhase, setInstallationPhase] = useState('Main');
    const [selectedOption, setSelectedOption] = useState(0);
    return (
        <div className="installation-modal">
            <header is="x3d">
                <p>{INSTALLATION_GUIDE[installationPhase].header}</p>
                <span>{INSTALLATION_GUIDE[installationPhase].description} </span>
            </header>
            <content is="x3d">
                {installationPhase === 'Main' && (
                    <>
                        <p className="content-title">Choose your preferred environment</p>
                        <div>
                            {option.map((value, key) => {
                                return (
                                    <div
                                        key={key}
                                        className={selectedOption === key ? 'option-wrapper selected' : 'option-wrapper'}
                                        onClick={() => setSelectedOption(key)}
                                    >
                                        <p>{value}</p>
                                        {selectedOption === key && <CheckCircleIcon className="check-icon" />}
                                        {selectedOption !== key && <div className="uncheck-icon" />}
                                    </div>
                                );
                            })}
                        </div>
                    </>
                )}
                {installationPhase !== 'Main' && installationPhase !== 'Cloud Providers' && (
                    <InstallationCommand
                        steps={INSTALLATION_GUIDE[installationPhase].steps}
                        showLinks={INSTALLATION_GUIDE[installationPhase].showLinks}
                        videoLink={INSTALLATION_GUIDE[installationPhase].videoLink}
                        docsLink={INSTALLATION_GUIDE[installationPhase].docsLink}
                    />
                )}
                {installationPhase === 'Cloud Providers' && <CloudeProviders steps={INSTALLATION_GUIDE[installationPhase]} />}
            </content>
            <buttons is="x3d">
                <Button
                    width="186px"
                    height="34px"
                    placeholder={installationPhase === 'Main' ? 'Close' : 'Back'}
                    colorType="black"
                    border="gray-light"
                    radiusType="circle"
                    backgroundColorType="white"
                    fontSize="12px"
                    fontWeight="bold"
                    boxShadowStyle="none"
                    onClick={() => {
                        installationPhase === 'Main' ? closeModal() : setInstallationPhase('Main');
                    }}
                />
                <Button
                    width="224px"
                    height="34px"
                    placeholder={installationPhase === 'Main' ? 'Next' : 'Finish'}
                    colorType="white"
                    radiusType="circle"
                    backgroundColorType="purple"
                    fontSize="12px"
                    fontWeight="bold"
                    onClick={() => (installationPhase === 'Main' ? setInstallationPhase(option[selectedOption]) : closeModal())}
                />
            </buttons>
        </div>
    );
};

export default Installation;
