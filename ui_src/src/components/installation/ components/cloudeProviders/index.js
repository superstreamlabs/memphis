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
                                            <img src={redirectIcon} alt="redirectIcon" />
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
                        <img width={25} height={22} src={docsPurple} alt="docsPurple" />
                        <p>Link to docs</p>
                        <ChevronRightOutlined />
                    </div>
                </Link>
            </div>
        </div>
    );
};

export default CloudeProviders;
