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

import React, { useRef } from 'react';
import { HiLockClosed } from 'react-icons/hi';

import TooltipComponent from '../tooltip/tooltip';
import UpgradePlans from '../upgradePlans';
import { showUpgradePlan } from '../../services/valueConvertor';

const LockFeature = ({ position, header }) => {
    const style = {
        position: position || 'absolute'
    };

    const tooltipRef = useRef(null);

    return (
        <TooltipComponent
            className="lock-feature-tooltip"
            style={style}
            placement="right"
            tooltipRef={tooltipRef}
            text={
                <div className="lock-tooltip-text">
                    <p>{header || 'Upgrade needed'}</p>
                    <span>Please upgrade your plan to unlock this feature</span>
                    {showUpgradePlan() && (
                        <UpgradePlans
                            content={
                                <div className="upgrade-button-wrapper" onClick={() => tooltipRef.current()}>
                                    <p className="upgrade-plan">Upgrade now</p>
                                </div>
                            }
                            isExternal={false}
                        />
                    )}
                </div>
            }
        >
            <HiLockClosed className="lock-feature-icon" />
        </TooltipComponent>
    );
};

export default LockFeature;
