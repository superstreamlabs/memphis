// Credit for The NATS.IO Authors
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

import './style.scss';

import { ChevronRightOutlined } from '@material-ui/icons';

import React from 'react';
import { Link } from 'react-router-dom';

import redirectIcon from '../../../../assets/images/redirectIcon.svg';
import videoIcon from '../../../../assets/images/videoIcon.svg';
import docsPurple from '../../../../assets/images/docsPurple.svg';
import Copy from '../../../copy';

const InstallationCommand = ({ steps, showLinks, videoLink, docsLink }) => {
    return (
        <div className="installation-command">
            {steps.length > 0 && (
                <div className="steps">
                    {steps.map((value, key) => {
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
            )}
            {showLinks && (
                <div className="links">
                    <Link to={{ pathname: videoLink }} target="_blank">
                        <div className="link-wrapper">
                            <img src={videoIcon} alt="videoIcon" />
                            <p>Installation video</p>
                            <ChevronRightOutlined />
                        </div>
                    </Link>
                    <Link to={{ pathname: docsLink }} target="_blank">
                        <div className="link-wrapper">
                            <img width={25} height={22} src={docsPurple} alt="docsPurple" />
                            <p>Link to docs</p>
                            <ChevronRightOutlined />
                        </div>
                    </Link>
                </div>
            )}
        </div>
    );
};

export default InstallationCommand;
