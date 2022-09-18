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

import './style.scss';

import React, { useEffect, useContext, useState } from 'react';

import LogBadge from '../../../../components/logBadge';
import Button from '../../../../components/button';
import copy from '../../../../assets/images/copy.svg';
import Copied from '../../../../assets/images/copied.svg';
import { capitalizeFirst, cutInfoLog, parsingDate } from '../../../../services/valueConvertor';

const LogContent = ({ displayedLog }) => {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        setCopied(true);
        navigator.clipboard.writeText(cutInfoLog(displayedLog?.data));
        setTimeout(() => {
            setCopied(false);
        }, 3000);
    };

    return (
        <div className="log-content-wrapper">
            <log-header is="3xd">
                <p>Log Details</p>
            </log-header>
            <log-payload is="3xd">
                <div className="log-details">
                    <div className="source">
                        <p className="title">Source</p>
                        <span className="des">{displayedLog?.source && capitalizeFirst(displayedLog?.source)}</span>
                    </div>
                    <div className="type">
                        <p className="title">Type</p>
                        <LogBadge type={displayedLog?.type} />
                    </div>
                    <div className="date">
                        <p className="title">Time</p>
                        <span className="des">{parsingDate(displayedLog?.creation_date)}</span>
                    </div>
                </div>
                <div></div>
            </log-payload>
            <log-content is="3xd">
                <p>{cutInfoLog(displayedLog?.data)}</p>
                <div className="copy-button" onClick={() => handleCopy()}>
                    <img src={copied ? Copied : copy} />
                </div>
            </log-content>
        </div>
    );
};

export default LogContent;
