// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import './style.scss';

import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import React, { useState } from 'react';

import { INSTALLATION_GUIDE } from '../../const/installationGuide';
import Button from '../button';
import InstallationCommand from './ components/installationCommand';

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
                {installationPhase !== 'Main' && (
                    <InstallationCommand
                        steps={INSTALLATION_GUIDE[installationPhase].steps}
                        showLinks={INSTALLATION_GUIDE[installationPhase].showLinks}
                        videoLink={INSTALLATION_GUIDE[installationPhase].videoLink}
                        docsLink={INSTALLATION_GUIDE[installationPhase].docsLink}
                    />
                )}
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
