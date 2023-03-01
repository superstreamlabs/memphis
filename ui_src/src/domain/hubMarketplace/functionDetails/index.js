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

import '../functionsList/style.scss';
import './style.scss';

import React, { useState } from 'react';

import Button from '../../../components/button';
import FunctionForm from '../functionForm';

function FunctionDetails(props) {
    const [isInstalled, setInstall] = useState(false); // Placeholder -  will be received from state
    const [openFunctionForm, setOpenFunctionForm] = useState(false);

    return (
        <div className="functions-details-container">
            <FunctionForm open={openFunctionForm} chosenFunction={props.chosenFunction} closeModal={() => setOpenFunctionForm(false)} />

            <div className="functions-details-header">
                <p>Details</p>
            </div>
            <div className="functions-details-body">
                {props.chosenFunction && (
                    <div>
                        <div className="function-details">
                            {props.chosenFunction.funcImg ? (
                                <img src={props.chosenFunction.funcImg} alt="function" width="50" height="50" className="img-placeholder" />
                            ) : (
                                <div className="img-placeholder" />
                            )}
                            <div>
                                <p className="function-name">{props.chosenFunction.funcName}</p>
                                <p className="data-type">Data type: {props.chosenFunction.inputDataType}</p>
                            </div>
                        </div>
                        <div className="functions-details-section">
                            <div className="func-description">{props.chosenFunction.funcDesc}</div>
                            <p className="visit-hub">Visit hub</p>
                        </div>
                    </div>
                )}
            </div>
            {props.chosenFunction && (
                <div className="functions-details-footer">
                    <Button
                        className="modal-btn"
                        width="90px"
                        height="32px"
                        placeholder={isInstalled ? 'Uninstall' : 'Install'}
                        colorType={isInstalled ? 'purple' : 'purple'}
                        backgroundColorType={isInstalled ? 'none' : 'purple'}
                        border={isInstalled ? 'purple' : null}
                        radiusType="circle"
                        fontSize="14px"
                        fontWeight="bold"
                        aria-haspopup="true"
                        onClick={() => setInstall(!isInstalled)}
                    />
                    <Button
                        className="modal-btn"
                        width="90px"
                        height="32px"
                        placeholder="Use"
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="14px"
                        fontWeight="bold"
                        aria-haspopup="true"
                        disabled={!isInstalled}
                        onClick={() => setOpenFunctionForm(true)}
                    />
                </div>
            )}
        </div>
    );
}
export default FunctionDetails;
