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
import React, { useContext } from 'react';
import UpgradePlans from '../../../components/upgradePlans';

import { Context } from '../../../hooks/store';

import './style.scss';
const Usage = () => {
    const [state, dispatch] = useContext(Context);

    const actual = state?.monitor_data?.billing_details?.actual_usage || 0;
    const total = state?.monitor_data?.billing_details?.total_included || 1;
    const widthInPercentage = (actual / total) * 100 > 100 ? 100 : (actual / total) * 100;

    const dataStyle = {
        width: `${widthInPercentage}%`,
        borderTopRightRadius: widthInPercentage > 99.5 ? 'inherit' : '0px',
        borderBottomRightRadius: widthInPercentage > 99.5 ? 'inherit' : '0px'
    };
    return (
        <div className="overview-components-wrapper">
            <div className="overview-usage-container">
                <div className="overview-components-header usage-header">
                    <p>Free plan usage</p>
                    <UpgradePlans
                        content={
                            <div className="upgrade-button-wrapper">
                                <p className="upgrade-plan">Upgrade now</p>
                            </div>
                        }
                        isExternal={false}
                    />
                </div>
                <div className="usage-body">
                    <div className="usageLeft-side">
                        <div
                            className="usageLeft-label"
                            style={{ paddingLeft: `${widthInPercentage}%`, marginLeft: widthInPercentage > 99.5 ? '-1px' : widthInPercentage < 0.1 ? '1px' : '0px' }}
                        >
                            <div className="dividerContainer">
                                <span className="labelMain">Current usage</span>
                                <span className="labelSecondary">{`${actual} GB`}</span>
                            </div>
                        </div>
                        <div className="totalContainer">
                            <div className="dataContainer" style={{ ...dataStyle }} />
                        </div>
                    </div>
                    <div className="usageRight-side">
                        <p className="mainText">{total}GB</p>
                        <p className="secondaryText">Storage included</p>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Usage;
