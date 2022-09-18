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
