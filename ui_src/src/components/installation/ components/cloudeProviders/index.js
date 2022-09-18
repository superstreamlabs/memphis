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

import { ChevronRightOutlined } from '@material-ui/icons';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';

import React, { useState } from 'react';
import { Link } from 'react-router-dom';

import redirectIcon from '../../../../assets/images/redirectIcon.svg';
import docsPurple from '../../../../assets/images/docsPurple.svg';
import Copy from '../../../copy';

const CloudeProviders = ({ steps }) => {
    const [cloudSelected, setcloudSelected] = useState('aws');
    const [docsLink, setDocsLink] = useState(steps?.clouds[0].docsLink);

    const handleSelectedCloud = (name, link) => {
        setcloudSelected(name);
        setDocsLink(link);
    };

    return (
        <div className="installation-command">
            <div className="steps">
                <div className="step-wrapper">
                    <p className="step-title">Choose your cloud:</p>
                    <div className="img-wrapper">
                        {steps?.clouds &&
                            steps?.clouds.map((value) => {
                                return (
                                    <div
                                        key={value.name}
                                        className={cloudSelected === value.name ? 'img-cloud selected' : 'img-cloud'}
                                        onClick={() => handleSelectedCloud(value.name, value.docsLink)}
                                    >
                                        {cloudSelected === value.name && (
                                            <div className="selected-icon">
                                                <CheckCircleIcon />
                                            </div>
                                        )}
                                        {value.src}
                                    </div>
                                );
                            })}
                    </div>
                </div>
                {steps[cloudSelected]?.map((value, key) => {
                    return (
                        <div className="step-wrapper" key={key}>
                            <p className="step-title">{value.title}</p>
                            {value.command && (
                                <div className="step-command">
                                    <span>{value.command}</span>
                                    {value.icon === 'copy' && <Copy data={value.command} key={key} />}
                                    {value.icon === 'link' && (
                                        <Link to={{ pathname: 'http://localhost:5555' }} target="_blank">
                                            <img src={redirectIcon} />
                                        </Link>
                                    )}
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>
            <div className="links">
                <Link to={{ pathname: docsLink }} target="_blank">
                    <div className="link-wrapper">
                        <img width={25} height={22} src={docsPurple} />
                        <p>Link to docs</p>
                        <ChevronRightOutlined />
                    </div>
                </Link>
            </div>
        </div>
    );
};

export default CloudeProviders;
