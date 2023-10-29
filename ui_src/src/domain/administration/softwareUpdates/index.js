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
import { ReactComponent as RedirectIcon } from '../../../assets/images/redirectIcon.svg';
import { ReactComponent as RedirectWhiteIcon } from '../../../assets/images/exportWhite.svg';
import CustomGauge from './components/customGauge';
const data = {
    labels: ['Red', 'Green', 'Blue'],
    datasets: [
        {
            data: [40, 60, 100],
            backgroundColor: ['red', 'green', 'transparent'],
            hoverBackgroundColor: ['red', 'green', 'transparent'],
            borderWidth: 2, // Add a border for the outer doughnut
            borderColor: 'white', // Set the border color for the outer doughnut
            cutoutPercentage: 70, // Adjust the cutout percentage to control the size of the inner doughnut
            borderAlign: 'inner', // Make sure the border is drawn inside the doughnut
            // hoverBorderWidth: 5 // Adjust the hover border width
            hoverBorderWidth: 0, // Adjust the hover border width
            borderRadius: 10,
            hoverBorderColor: 'transparent'
        }
    ]
};

const options = {
    cutoutPercentage: 70, // Adjust the cutout percentage to control the size of the innermost hole
    rotation: Math.PI,
    tooltips: false
};

const options2 = {
    cutoutPercentage: 10, // Adjust the cutout percentage to control the size of the innermost hole
    rotation: Math.PI,
    tooltips: false
};
function SoftwareUpates({}) {
    return (
        <div className="softwate-updates-container">
            <div className="rows">
                <div className="item-component">
                    <div className="title-component">
                        <div className="versions">
                            <LogoTexeMemphis alt="Memphis logo" />
                            <label className="curr-version">v0.4.3 - Beta</label>
                            <span className="new-version">
                                <label>New Version available </label>
                                <RedirectIcon alt="redirect" />
                            </span>
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
                            // htmlType="submit"
                            onClick={() => console.log('clicked')}
                        />
                    </div>
                </div>
                <div className="statistics">
                    <div className="item-component wrapper">
                        <label className="title">Amount of brokers</label>
                        <label className="numbers">600</label>
                    </div>
                    <div className="item-component wrapper">
                        <label className="title">Amount of brokers</label>
                        <label className="numbers">600</label>
                    </div>
                    <div className="item-component wrapper">
                        <label className="title">Amount of brokers</label>
                        <label className="numbers">600</label>
                    </div>
                    <div className="item-component wrapper">
                        <label className="title">Amount of brokers</label>
                        <label className="numbers">600</label>
                    </div>
                </div>
                <div className="charts">
                    <div className="item-component">
                        <label className="title">Amount of brokers</label>
                        <CustomGauge />
                    </div>
                    <div className="item-component">
                        <label className="title">Amount of brokers</label>
                    </div>
                    <div className="item-component">
                        <label className="title">Amount of brokers</label>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default SoftwareUpates;
