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

import React from 'react';
import RightArrow from '../../../../assets/images/rightArrow.svg';
import './style.scss';
import Done from '../../../../assets/images/done.svg';

const SideStep = (props) => {
    const { stepNumber, stepName, currentStep, completedSteps } = props;
    return (
        <div
            className={
                currentStep === stepNumber
                    ? completedSteps + 1 >= stepNumber
                        ? 'side-step-container curr-step cursor-allowed'
                        : 'side-step-container curr-step cursor-blocked'
                    : completedSteps + 1 >= stepNumber
                    ? 'side-step-container cursor-allowed'
                    : 'side-step-container cursor-blocked'
            }
            onClick={() => completedSteps + 1 >= stepNumber && props.onSideBarClick(stepNumber)}
        >
            <div className="number-name-container">
                <div className={currentStep >= stepNumber ? 'step-number-container step-number-white' : 'step-number-container'}>
                    {stepNumber <= completedSteps ? (
                        <div className="done-image">
                            <img src={Done} alt="done" />
                        </div>
                    ) : (
                        <p className="step-number">{stepNumber}</p>
                    )}
                </div>
                <p className={currentStep === stepNumber ? 'step-name curr-step-name' : 'step-name'}>{stepName}</p>
            </div>
            <div className="arrow-container">{currentStep === stepNumber && <img src={RightArrow} alt="select-arrow" />}</div>
        </div>
    );
};
export default SideStep;
