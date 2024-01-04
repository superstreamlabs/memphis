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
import { ReactComponent as FunctionIntegrateIcon } from 'assets/images/functionIntegrate.svg';
import { ReactComponent as CloneModalIcon } from 'assets/images/cloneModalIcon.svg';
import { MdDone } from 'react-icons/md';
import VideoPlayer from 'components/videoPlayer';
import Button from 'components/button';
import Modal from 'components/modal';
import CloneModal from 'components/cloneModal';
import { FUNCTION_GUIDE_VIDEO } from 'config';
import ConnectBG from 'assets/images/functionsWelcomeBanner.webp';
import { LuInfo } from 'react-icons/lu';

const FunctionsGuide = ({ handleClose, handleConfirm }) => {
    const [currentStep, setCurrentStep] = useState(0);
    const [isCloneModalOpen, setIsCloneModalOpen] = useState(false);
    const [cloneType, setCloneType] = useState('functions');

    const handleCloneClick = (type) => {
        setCloneType(type);
        setIsCloneModalOpen(true);
    };

    const steps = [
        {
            name: (
                <>
                    <label>Clone or create a new GitHub repository </label>
                    <label className="link" onClick={() => handleCloneClick('functions')}>
                        {' '}
                        (templates can be found here)
                    </label>
                </>
            )
        },
        {
            name: (
                <>
                    <label>Code your function based on the following </label>
                    <label
                        className="link"
                        onClick={() => window.open(`https://docs.memphis.dev/memphis/memphis-functions/getting-started#how-to-develop-a-new-private-function`)}
                    >
                        guide
                    </label>
                </>
            )
        },
        { name: 'Commit your function' },
        { name: 'Connect the newly created repository with Memphis' }
    ];

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
                <p className="sub-title">Say Goodbye To Writing Business Logic In Your Clients!</p>
                <p className="sub-title">Embrace Lightning-Speed Serverless Stream Processing.</p>
            </div>
            <div className="video-wrapper">
                <VideoPlayer url={FUNCTION_GUIDE_VIDEO} bgImg={ConnectBG} width={'540px'} height={'250px'} tracePlay />
            </div>
            <div className="info">
                <p className="info-title">Getting Started</p>
                <>
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
                            </div>
                            {index < steps.length - 1 && (
                                <div className={`step-body`}>
                                    <p className="description"></p>
                                </div>
                            )}
                        </div>
                    ))}
                </>
            </div>
            <div className="need-help">
                <LuInfo className="msg" />
                <label className="bold">Require assistance?</label>
                <label> Submit a service request and we will come to the rescue!</label>
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
                    placeholder={'Next'}
                    colorType={'white'}
                    backgroundColorType={'purple'}
                    radiusType={'circle'}
                    onClick={currentStep === steps.length - 1 ? handleConfirm : handleNext}
                    fontWeight={600}
                    fontSize={'12px'}
                />
            </div>
            <Modal
                header={<CloneModalIcon alt="cloneModalIcon" />}
                width="540px"
                displayButtons={false}
                clickOutside={() => setIsCloneModalOpen(false)}
                open={isCloneModalOpen}
            >
                <CloneModal type={cloneType} />
            </Modal>
        </div>
    );
};

export default FunctionsGuide;
