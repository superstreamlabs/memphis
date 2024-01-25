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

import React from 'react';
import OverflowTip from 'components/tooltip/overflowtip';

const inputsColumns = [
    {
        key: '1',
        title: 'Key',
        width: '300px'
    },
    {
        key: '2',
        title: 'Value',
        width: '300px'
    }
];

const FunctionInformation = ({ inputs }) => {
    return (
        <div className="function-inputs-container">
            <p className="title">Inputs</p>
            <div className="generic-list-wrapper">
                <div className="list">
                    <div className="coulmns-table">
                        {inputsColumns?.map((column, index) => {
                            return (
                                <span key={index} style={{ width: column.width }}>
                                    {column.title}
                                </span>
                            );
                        })}
                    </div>
                    {(!inputs || Object.entries(inputs)?.length === 0) && (
                        <div className="rows-wrapper">
                            <p className="no-inputs">This function has no inputs</p>
                        </div>
                    )}
                    <div className="rows-wrapper">
                        {Object.entries(inputs)?.map((key, index) => {
                            return (
                                <div className="pubSub-row" key={key[0]}>
                                    <OverflowTip text={key[0]} width={'300px'}>
                                        {key[0]}
                                    </OverflowTip>
                                    <OverflowTip text={key[1]} width={'300px'}>
                                        {key[1]}
                                    </OverflowTip>
                                </div>
                            );
                        })}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default FunctionInformation;
