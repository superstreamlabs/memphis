// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
