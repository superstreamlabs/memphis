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

import { ReactComponent as RemoveFunctionIcon } from 'assets/images/removeFunctionIcon.svg';
import { ReactComponent as ArrowFunctionIcon } from 'assets/images/arrowFunction.svg';
import OverflowTip from 'components/tooltip/overflowtip';

const FunctionsOverview = (props) => {
    const { functions, horizontal, editable } = props;

    const handleRemoveFunction = (funcIndex) => {};
    const handleEditFunction = (funcIndex, func) => {};

    return (
        <div className={horizontal ? 'function-overview-container horizontal' : 'function-overview-container'}>
            {functions.map((func, index) => {
                return (
                    <div className={horizontal ? 'function-list-container horizontal' : 'function-list-container'} key={index}>
                        <div className="func-wrapper">
                            {editable && (
                                <div className="remove-button" onClick={() => handleRemoveFunction(index)}>
                                    <RemoveFunctionIcon alt="edit" width={8} height={8} />
                                </div>
                            )}
                            <div
                                className={horizontal ? 'function-box-overview horizontal' : 'function-box-overview'}
                                onClick={() => handleEditFunction(index + 1, func)}
                            >
                                <div className="function-name">
                                    <OverflowTip text={func.name} width={'7vw'} cursor="pointer">
                                        {func.name}
                                    </OverflowTip>
                                </div>
                            </div>
                        </div>
                        {index < functions?.length - 1 && (
                            <ArrowFunctionIcon alt="edit" width="4vw" style={{ transform: !horizontal && 'rotate(90deg)', margin: '15px' }} />
                        )}
                    </div>
                );
            })}
        </div>
    );
};

export default FunctionsOverview;
