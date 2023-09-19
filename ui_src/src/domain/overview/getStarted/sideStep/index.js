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

import React, { useState, useEffect, useMemo } from 'react';
import './style.scss';

import { ReactComponent as GetStartedIcon } from '../../../../assets/images/getStartedIcon.svg';
import { ReactComponent as AppUserIcon } from '../../../../assets/images/usersIconActive.svg';
import { ReactComponent as EmptyStationIcon } from '../../../../assets/images/emptyStation.svg';
import { ReactComponent as DataProducedIcon } from '../../../../assets/images/dataProduced.svg';
import { ReactComponent as ConsumeDataIcon } from '../../../../assets/images/stationsIconActive.svg';
import { ReactComponent as FullStationIcon } from '../../../../assets/images/fullStation.svg';
import { ReactComponent as FinishFlagIcon } from '../../../../assets/images/finishFlag.svg';
import { ReactComponent as GrayAppUserIcon } from '../../../../assets/images/grayAppUserIcon.svg';
import { ReactComponent as GrayProduceDataIcon } from '../../../../assets/images/grayProduceDataImg.svg';
import { ReactComponent as GrayConsumeDataIcon } from '../../../../assets/images/grayConsumeDataImg.svg';
import { ReactComponent as GrayfinishStepIcon } from '../../../../assets/images/grayFinish.svg';
import { ReactComponent as CompletedStepIcon } from '../../../../assets/images/checkIcon.svg';

const Step = ({ stepNumber, stepName, currentStep, completedSteps, stepsDescription, onSideBarClick }) => {
    const docLinks = {
        1: 'https://docs.memphis.dev/memphis-new/dashboard-ui/stations',
        2: 'https://docs.memphis.dev/memphis-new/dashboard-ui/users',
        3: 'https://docs.memphis.dev/memphis-new/memphis/concepts/producer',
        4: 'https://docs.memphis.dev/memphis-new/memphis/concepts/consumer'
    };
    const getIcon = useMemo(() => {
        switch (stepNumber) {
            case 1:
                return <GetStartedIcon className="sidebar-image" alt="getStartedIcon" />;
            case 2:
                return completedSteps + 1 >= stepNumber ? (
                    <AppUserIcon className="sidebar-image" alt="getStartedIcon" />
                ) : (
                    <GrayAppUserIcon className="sidebar-image" alt="getStartedIcon" />
                );
            case 3:
                if (completedSteps + 1 > stepNumber) return <DataProducedIcon className="sidebar-image" alt="getStartedIcon" />;
                else if (completedSteps + 1 === stepNumber) return <EmptyStationIcon className="sidebar-image" alt="getStartedIcon" />;
                else return <GrayProduceDataIcon className="sidebar-image" alt="getStartedIcon" />;
            case 4:
                if (completedSteps + 1 > stepNumber) return <ConsumeDataIcon className="sidebar-image" alt="getStartedIcon" />;
                else if (completedSteps + 1 === stepNumber) return <FullStationIcon className="sidebar-image" alt="getStartedIcon" />;
                else return <GrayConsumeDataIcon className="sidebar-image" alt="getStartedIcon" />;
            case 5:
                return completedSteps + 1 >= stepNumber ? (
                    <FinishFlagIcon className="sidebar-image" alt="getStartedIcon" />
                ) : (
                    <GrayfinishStepIcon className="sidebar-image" alt="getStartedIcon" />
                );
            default:
                return null;
        }
    }, [stepNumber, completedSteps]);
    return (
        <div
            className={completedSteps + 1 >= stepNumber ? 'side-step-container cursor-allowed' : 'side-step-container'}
            onClick={() => completedSteps + 1 >= stepNumber && onSideBarClick(stepNumber)}
        >
            <div className="side-step-header">
                {getIcon}
                <div className="step-name-completed">
                    <p className={currentStep === stepNumber ? 'step-name curr-step-name' : 'step-name'}>{stepName}</p>
                    {completedSteps >= stepNumber && stepNumber !== 5 && <CompletedStepIcon className="completed" alt="completed" />}
                </div>
            </div>
            <div className={completedSteps >= stepNumber ? 'side-step-body border-completed' : stepNumber !== 5 ? 'side-step-body border' : 'side-step-body'}>
                {stepNumber !== 5 && (
                    <p className={currentStep === stepNumber ? 'step-description curr-step-name' : 'step-description'}>
                        {stepsDescription}
                        {'. '}
                        <a href={docLinks[stepNumber]} target="_blank" rel="noopener noreferrer">
                            Learn more
                        </a>
                    </p>
                )}
            </div>
        </div>
    );
};

export default Step;
