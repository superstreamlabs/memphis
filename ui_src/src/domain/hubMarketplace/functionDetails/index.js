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
// limitations under the License.package server

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
