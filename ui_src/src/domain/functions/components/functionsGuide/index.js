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
import { useState } from 'react';
import { ReactComponent as FunctionIntegrateIcon } from '../../../../assets/images/functionIntegrate.svg';
import { ReactComponent as CopyIcon } from '../../../../assets/images/copy.svg';
import { MdDone } from 'react-icons/md';
import VideoPlayer from '../../../../components/videoPlayer';
import Button from '../../../../components/button';
import { CONNECT_APP_VIDEO } from '../../../../config';
import ConnectBG from '../../../../assets/images/connectBG.webp';

const steps = [
    { name: 'Create new repo and clone', description: 'Donec dictum tristique prota. Etiam convallis lorem lobortis nulla molestie' },

    {
        name: 'Clone the template or Download zip',
        description: 'Donec dictum tristique prota. Etiam convallis lorem lobortis nulla molestie'
    },
    {
        name: 'Copy the files to your repos',
        description: 'Donec dictum tristique prota. Etiam convallis lorem lobortis nulla molestie'
    },
    {
        name: 'Commit your new function',
        description: 'Donec dictum tristique prota. Etiam convallis lorem lobortis nulla molestie'
    },
    { name: 'Add the new function to Memphis', description: 'Donec dictum tristique prota. Etiam convallis lorem lobortis nulla molestie' }
];

const FunctionsGuide = ({ handleClose, handleConfirm, handleCloneClick }) => {
    const [currentStep, setCurrentStep] = useState(0);

    const handleNext = () => {
        setCurrentStep(currentStep + 1);
    };
    return (
        <div className="new-function-modal">
            <div className="header-icon">
                <FunctionIntegrateIcon width={22} height={22} />
            </div>
            <div className="title-wrapper">
                <p className="title">
                    Welcome to <span> Memphis Functions</span>
                </p>
                <p className="sub-title">A cool new way to stream processing!</p>
            </div>
            <div className="video-wrapper">
                <VideoPlayer url={CONNECT_APP_VIDEO} bgImg={ConnectBG} width={'540px'} height={'250px'} />
            </div>
            <div className="info">
                {steps.map((step, index) => (
                    <div className="step-container" key={index}>
                        <div className="step-header">
                            {index < currentStep ? (
                                <div className="done" onClick={() => setCurrentStep(index)}>
                                    <MdDone width={12} height={12} alt="Done" />
                                </div>
                            ) : (
                                <div className="icon" onClick={() => setCurrentStep(index)}>
                                    {index + 1}
                                </div>
                            )}
                            <div className="step-name">{step.name}</div>
                            {index === 1 && (
                                <Button
                                    height={'23px'}
                                    width={'125px'}
                                    placeholder={
                                        <div className="cloneButton">
                                            <CopyIcon width={12} height={12} />
                                            <span>Clone Template</span>
                                        </div>
                                    }
                                    colorType={'purple'}
                                    backgroundColorType={'white'}
                                    border={'gray-light'}
                                    onClick={handleCloneClick}
                                    radiusType={'circle'}
                                />
                            )}
                        </div>
                        <div className={`step-body ${index < currentStep && 'step-body-done'}`}>
                            <p className="description">{step.description}</p>
                        </div>
                    </div>
                ))}
            </div>

            <div className="footer">
                <Button
                    width={'100%'}
                    height={'34px'}
                    placeholder={'Cancel'}
                    colorType={'black'}
                    backgroundColorType={'white'}
                    border={'gray-light'}
                    radiusType={'circle'}
                    onClick={handleClose}
                    fontWeight={600}
                    fontSize={'12px'}
                />
                <Button
                    width={'100%'}
                    height={'34px'}
                    placeholder={currentStep < steps.length - 1 ? 'Next' : 'Done'}
                    colorType={'white'}
                    backgroundColorType={'purple'}
                    radiusType={'circle'}
                    onClick={currentStep === steps.length - 1 ? handleConfirm : handleNext}
                    fontWeight={600}
                    fontSize={'12px'}
                />
            </div>
        </div>
    );
};

export default FunctionsGuide;
